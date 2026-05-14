// internal/abs/client.go
// Audiobookshelf API client for audiobook-organizer

package abs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Client for ABS REST API
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	headers    map[string]string // Custom headers for each request
}

// NewClient creates a new ABS API client
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		headers:    make(map[string]string),
	}
}

// SetHeader adds a custom header to all requests
func (c *Client) SetHeader(key, value string) {
	c.headers[key] = value
}

// LoadHeadersFromFile loads headers from a file in KEY=VALUE format
func (c *Client) LoadHeadersFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading header file: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		// Remove quotes if present
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}

		c.headers[key] = value
	}

	return nil
}

// SetHTTPClient allows customizing the HTTP client (useful for tests)
func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

// request makes an authenticated HTTP request
func (c *Client) request(method, path string, body io.Reader) (*http.Response, error) {
	// Parse base URL
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Parse the path (which may include query params)
	reqURL, err := base.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL path: %w", err)
	}

	req, err := http.NewRequest(method, reqURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Apply custom headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		return nil, fmt.Errorf("authentication failed: invalid token")
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// GetLibraries returns all libraries from ABS
func (c *Client) GetLibraries() ([]Library, error) {
	resp, err := c.request("GET", "/api/libraries", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Libraries []Library `json:"libraries"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding libraries: %w", err)
	}

	return result.Libraries, nil
}

// GetLibraryItems returns items from a library with pagination
func (c *Client) GetLibraryItems(
	libraryID string,
	limit, offset int,
) (*LibraryItemsResponse, error) {
	path := fmt.Sprintf("/api/libraries/%s/items?limit=%d&offset=%d", libraryID, limit, offset)
	resp, err := c.request("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result LibraryItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding items: %w", err)
	}

	return &result, nil
}

// GetAllLibraryItems returns all items from a library (handles pagination)
func (c *Client) GetAllLibraryItems(libraryID string) ([]LibraryItem, error) {
	const limit = 100
	var allItems []LibraryItem
	offset := 0

	for {
		resp, err := c.GetLibraryItems(libraryID, limit, offset)
		if err != nil {
			return nil, err
		}

		allItems = append(allItems, resp.Results...)

		if offset+len(resp.Results) >= resp.Total {
			break
		}
		offset += limit
	}

	return allItems, nil
}

// GetLibraryItem returns a single library item by ID
func (c *Client) GetLibraryItem(itemID string) (*LibraryItem, error) {
	path := fmt.Sprintf("/api/items/%s", itemID)
	resp, err := c.request("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var item LibraryItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("decoding item: %w", err)
	}

	return &item, nil
}

// ScanLibrary triggers a library scan
func (c *Client) ScanLibrary(libraryID string) error {
	path := fmt.Sprintf("/api/libraries/%s/scan", libraryID)
	resp, err := c.request("POST", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// ScanLibraryForce triggers a forced library scan.
func (c *Client) ScanLibraryForce(libraryID string) error {
	path := fmt.Sprintf("/api/libraries/%s/scan?force=1", libraryID)
	resp, err := c.request("POST", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// RemoveLibraryItemsWithIssues removes a library's items that ABS reports as having issues.
func (c *Client) RemoveLibraryItemsWithIssues(libraryID string) error {
	path := fmt.Sprintf("/api/libraries/%s/issues", libraryID)
	resp, err := c.request("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// GetLibrary returns a single library by ID
func (c *Client) GetLibrary(libraryID string) (*Library, error) {
	path := fmt.Sprintf("/api/libraries/%s", libraryID)
	resp, err := c.request("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var lib Library
	if err := json.NewDecoder(resp.Body).Decode(&lib); err != nil {
		return nil, fmt.Errorf("decoding library: %w", err)
	}

	return &lib, nil
}
