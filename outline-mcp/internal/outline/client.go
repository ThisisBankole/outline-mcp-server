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


type CreateCollectionResult struct {
	Data struct {
		ID    string `json:"id"`
		Name string `json:"name"`
		URL   string `json:"url"`
	} `json:"data"`
}

func (c *Client) CreateCollection(name string) (*CreateCollectionResult, error) {
	var result CreateCollectionResult
	err := c.post("/collections.create", map[string]any{"name": name}, &result)
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
