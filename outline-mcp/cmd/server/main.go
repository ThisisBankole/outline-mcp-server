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
