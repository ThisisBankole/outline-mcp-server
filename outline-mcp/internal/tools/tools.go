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
	registerCreateCollection(s, client)
	registerSearchDocumentTitles(s, client)
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

func registerCreateCollection(s *server.MCPServer, client *outline.Client) {
	tool := mcp.NewTool("create_collection",
		mcp.WithDescription("Create a new collection in Outline"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the new collection")),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, _ := req.Params.Arguments["name"].(string)
		result, err := client.CreateCollection(name)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf(
			"Collection created!\nID:   %s\nName: %s\nURL:  %s",
			result.Data.ID, result.Data.Name, result.Data.URL,
		)), nil
	})
}

func registerSearchDocumentTitles(s *server.MCPServer, client *outline.Client) {
	tool := mcp.NewTool("search_document_titles",
		mcp.WithDescription("Search documents by title keyword in Outline"),
		mcp.WithString("query", mcp.Required(), mcp.Description("The title search query")),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, _ := req.Params.Arguments["query"].(string)
		result, err := client.SearchDocumentTitles(query)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		out, _ := json.MarshalIndent(result.Data, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})
}
