#!/usr/bin/env bash
set -euo pipefail

PROJECT="outline-mcp"

echo "📁 Creating project structure..."
mkdir -p "$PROJECT"/{cmd/server,internal/{outline,tools},config}
cd "$PROJECT"

echo "📄 Writing go.mod..."
cat > go.mod << 'EOF'
module github.com/yourusername/outline-mcp

go 1.22

require (
	github.com/mark3labs/mcp-go v0.18.0
	github.com/joho/godotenv v1.5.1
)
EOF

echo "📄 Writing .env.example..."
cat > .env.example << 'EOF'
OUTLINE_BASE_URL=https://notes.arrakis.house
OUTLINE_API_TOKEN=your_outline_api_token_here
MCP_BEARER_TOKEN=choose_a_strong_secret_here
PORT=8080
EOF

echo "📄 Writing .gitignore..."
cat > .gitignore << 'EOF'
.env
outline-mcp
dist/
EOF

echo "📄 Writing config/config.go..."
cat > config/config.go << 'EOF'
package config

import (
	"fmt"
	"os"
)

type Config struct {
	OutlineBaseURL  string
	OutlineAPIToken string
	MCPBearerToken  string
	Port            string
}

func Load() (*Config, error) {
	cfg := &Config{
		OutlineBaseURL:  os.Getenv("OUTLINE_BASE_URL"),
		OutlineAPIToken: os.Getenv("OUTLINE_API_TOKEN"),
		MCPBearerToken:  os.Getenv("MCP_BEARER_TOKEN"),
		Port:            os.Getenv("PORT"),
	}
	if cfg.OutlineBaseURL == "" {
		return nil, fmt.Errorf("OUTLINE_BASE_URL is required")
	}
	if cfg.OutlineAPIToken == "" {
		return nil, fmt.Errorf("OUTLINE_API_TOKEN is required")
	}
	if cfg.MCPBearerToken == "" {
		return nil, fmt.Errorf("MCP_BEARER_TOKEN is required")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	return cfg, nil
}
EOF

echo "📄 Writing internal/outline/client.go..."
cat > internal/outline/client.go << 'EOF'
package outline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) post(path string, payload, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api"+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("outline API error %d: %s", resp.StatusCode, string(raw))
	}
	return json.Unmarshal(raw, result)
}

type SearchResult struct {
	Data []struct {
		Document struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			URL   string `json:"url"`
			Text  string `json:"text"`
		} `json:"document"`
		Context string `json:"context"`
	} `json:"data"`
}

func (c *Client) SearchDocuments(query string, limit int) (*SearchResult, error) {
	var result SearchResult
	err := c.post("/documents.search", map[string]any{"query": query, "limit": limit}, &result)
	return &result, err
}

type DocumentResult struct {
	Data struct {
		ID           string `json:"id"`
		Title        string `json:"title"`
		Text         string `json:"text"`
		URL          string `json:"url"`
		CollectionID string `json:"collectionId"`
		UpdatedAt    string `json:"updatedAt"`
	} `json:"data"`
}

func (c *Client) GetDocument(id string) (*DocumentResult, error) {
	var result DocumentResult
	err := c.post("/documents.info", map[string]any{"id": id}, &result)
	return &result, err
}

type CollectionsResult struct {
	Data []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"data"`
}

func (c *Client) ListCollections() (*CollectionsResult, error) {
	var result CollectionsResult
	err := c.post("/collections.list", map[string]any{}, &result)
	return &result, err
}

type DocumentsListResult struct {
	Data []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		URL   string `json:"url"`
	} `json:"data"`
}

func (c *Client) ListDocuments(collectionID string) (*DocumentsListResult, error) {
	var result DocumentsListResult
	err := c.post("/documents.list", map[string]any{"collectionId": collectionID}, &result)
	return &result, err
}

type CreateDocumentResult struct {
	Data struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		URL   string `json:"url"`
	} `json:"data"`
}

func (c *Client) CreateDocument(title, text, collectionID string, publish bool) (*CreateDocumentResult, error) {
	var result CreateDocumentResult
	err := c.post("/documents.create", map[string]any{
		"title":        title,
		"text":         text,
		"collectionId": collectionID,
		"publish":      publish,
	}, &result)
	return &result, err
}
EOF

echo "📄 Writing internal/tools/tools.go..."
cat > internal/tools/tools.go << 'EOF'
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/yourusername/outline-mcp/internal/outline"
)

func Register(s *server.MCPServer, client *outline.Client) {
	registerSearchDocuments(s, client)
	registerGetDocument(s, client)
	registerListCollections(s, client)
	registerListDocuments(s, client)
	registerCreateDocument(s, client)
}

func registerSearchDocuments(s *server.MCPServer, client *outline.Client) {
	tool := mcp.NewTool("search_documents",
		mcp.WithDescription("Search documents in Outline by keyword or phrase"),
		mcp.WithString("query", mcp.Required(), mcp.Description("The search query")),
		mcp.WithNumber("limit", mcp.Description("Max results to return (default 10)")),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, _ := req.Params.Arguments["query"].(string)
		limit := 10
		if l, ok := req.Params.Arguments["limit"].(float64); ok && l > 0 {
			limit = int(l)
		}
		result, err := client.SearchDocuments(query, limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		out, _ := json.MarshalIndent(result.Data, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})
}

func registerGetDocument(s *server.MCPServer, client *outline.Client) {
	tool := mcp.NewTool("get_document",
		mcp.WithDescription("Fetch the full content of a document by its ID"),
		mcp.WithString("id", mcp.Required(), mcp.Description("The document UUID")),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, _ := req.Params.Arguments["id"].(string)
		result, err := client.GetDocument(id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		out, _ := json.MarshalIndent(result.Data, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})
}

func registerListCollections(s *server.MCPServer, client *outline.Client) {
	tool := mcp.NewTool("list_collections",
		mcp.WithDescription("List all collections in your Outline workspace"),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.ListCollections()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		out, _ := json.MarshalIndent(result.Data, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})
}

func registerListDocuments(s *server.MCPServer, client *outline.Client) {
	tool := mcp.NewTool("list_documents",
		mcp.WithDescription("List documents within a specific collection"),
		mcp.WithString("collection_id", mcp.Required(), mcp.Description("The collection UUID")),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		collectionID, _ := req.Params.Arguments["collection_id"].(string)
		result, err := client.ListDocuments(collectionID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		out, _ := json.MarshalIndent(result.Data, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})
}

func registerCreateDocument(s *server.MCPServer, client *outline.Client) {
	tool := mcp.NewTool("create_document",
		mcp.WithDescription("Create a new document in Outline"),
		mcp.WithString("title", mcp.Required(), mcp.Description("Title of the new document")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Body content in Markdown")),
		mcp.WithString("collection_id", mcp.Required(), mcp.Description("UUID of the collection to create the document in")),
		mcp.WithBoolean("publish", mcp.Description("Whether to publish immediately (default true)")),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		title, _ := req.Params.Arguments["title"].(string)
		text, _ := req.Params.Arguments["text"].(string)
		collectionID, _ := req.Params.Arguments["collection_id"].(string)
		publish := true
		if p, ok := req.Params.Arguments["publish"].(bool); ok {
			publish = p
		}
		result, err := client.CreateDocument(title, text, collectionID, publish)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf(
			"Document created!\nID:    %s\nTitle: %s\nURL:   %s",
			result.Data.ID, result.Data.Title, result.Data.URL,
		)), nil
	})
}
EOF

echo "📄 Writing cmd/server/main.go..."
cat > cmd/server/main.go << 'EOF'
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"
	"github.com/yourusername/outline-mcp/config"
	"github.com/yourusername/outline-mcp/internal/outline"
	"github.com/yourusername/outline-mcp/internal/tools"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	outlineClient := outline.NewClient(cfg.OutlineBaseURL, cfg.OutlineAPIToken)

	mcpServer := server.NewMCPServer(
		"outline-mcp",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	tools.Register(mcpServer, outlineClient)

	sseHandler := server.NewSSEServer(mcpServer,
		server.WithBaseURL(fmt.Sprintf("http://localhost:%s", cfg.Port)),
	)

	http.Handle("/", authMiddleware(cfg.MCPBearerToken, sseHandler))

	addr := ":" + cfg.Port
	log.Printf("outline-mcp listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func authMiddleware(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		auth := r.Header.Get("Authorization")
		provided := strings.TrimPrefix(auth, "Bearer ")
		if provided != token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
EOF

echo "📄 Writing Makefile..."
cat > Makefile << 'EOF'
.PHONY: run build tidy

run:
	go run ./cmd/server

build:
	go build -o outline-mcp ./cmd/server

tidy:
	go mod tidy
EOF

echo ""
echo "✅ Done! To get started:"
echo ""
echo "  cd $PROJECT"
echo "  cp .env.example .env   # then fill in your tokens"
echo "  make tidy"
echo "  make run"
