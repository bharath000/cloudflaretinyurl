package e2etest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Struct for API request & response
type URLRequest struct {
	LongURL   string     `json:"long_url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type URLResponse struct {
	ShortURL  string     `json:"short_url"`
	LongURL   string     `json:"long_url"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type ClickCounts struct {
	AllTime     int `json:"all_time"`
	Last1Min    int `json:"last_1min"`
	Last24Hours int `json:"last_24_hours"`
	LastWeek    int `json:"last_week"`
}

var baseAPI = "http://cloudflaretinyurl_service:8080/api/v1"
var shortURLs = make(map[string]string) // Stores created URLs for testing
var testURLs = make(map[string]string)  // Stores short URLs for delete validation

// Function to convert returned short URL to a full URL with baseAPI
func formatShortURL(returnedURL string) string {
	parsedURL, err := url.Parse(returnedURL)
	if err != nil {
		fmt.Println("Error parsing returned short URL:", err)
		return ""
	}
	shortCode := strings.TrimPrefix(parsedURL.Path, "/api/v1/") // Extract the last part
	return fmt.Sprintf("%s/%s", baseAPI, shortCode)             // Construct new URL with baseAPI
}

// Test 1: Create 10 unique short URLs and validate uniqueness
func TestCreateUniqueShortURLsE2E(t *testing.T) {
	for i := 0; i < 10; i++ {
		longURL := fmt.Sprintf("https://example.com/page%d", i)

		requestBody := URLRequest{LongURL: longURL}
		jsonData, err := json.Marshal(requestBody)
		assert.NoError(t, err)

		resp, err := http.Post(baseAPI+"/create", "application/json", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response URLResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		resp.Body.Close()

		// Validate uniqueness
		assert.NotEmpty(t, response.ShortURL, "Generated short URL should not be empty")
		_, exists := shortURLs[response.ShortURL]
		assert.False(t, exists, "Duplicate short URL detected: "+response.ShortURL)

		// Store short URL
		formattedShortURL := formatShortURL(response.ShortURL)
		shortURLs[formattedShortURL] = longURL
	}

	// Ensure exactly 10 unique URLs were created
	assert.Equal(t, 10, len(shortURLs), "Generated URLs are not unique")
}

// Test 2: Validate short URLs redirect correctly
func isRedirectSuccessful(shortURL string) bool {
	fmt.Println("Redirecting to:", shortURL)

	resp, err := http.Get(shortURL)
	if err != nil {
		fmt.Println("Error during redirect:", err)
		return false
	}
	defer resp.Body.Close()

	fmt.Println("Received status:", resp.StatusCode)

	// Only validate the HTTP status code, ignore if the redirected page does not exist (404)
	return resp.StatusCode == http.StatusFound // HTTP 302
}

func TestRedirectShortURLsE2E(t *testing.T) {
	assert.Greater(t, len(shortURLs), 0, "No short URLs generated, run TestCreateUniqueShortURLsE2E first")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Prevent the client from following redirects
			return http.ErrUseLastResponse
		},
	}

	for shortURL := range shortURLs {
		fmt.Println("Testing redirect for:", shortURL)

		resp, err := client.Get(shortURL)
		assert.NoError(t, err)
		defer resp.Body.Close()

		fmt.Println("Received status:", resp.StatusCode)
		assert.Equal(t, http.StatusFound, resp.StatusCode, "Expected HTTP 302 Redirect")
	}
}

// Test 3: Click Tracking & Expiry Validation
func TestShortURLClickTrackingE2E(t *testing.T) {
	longURL := "https://example.com/testpage"
	requestBody := URLRequest{LongURL: longURL}
	jsonData, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	resp, err := http.Post(baseAPI+"/create", "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response URLResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	resp.Body.Close()

	testShortURL := formatShortURL(response.ShortURL)
	assert.NotEmpty(t, testShortURL, "Failed to create test short URL")
	fmt.Println("Created short URL:", testShortURL)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Prevent following redirects
		},
	}
	// Step 2: Redirect 5 times with 5-sec delay
	for i := 0; i < 5; i++ {
		resp, err := client.Get(testShortURL) // Use custom client
		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, resp.StatusCode) // Expecting HTTP 302
		time.Sleep(5 * time.Second)
	}

	// Step 3: Validate Click Counts
	time.Sleep(2 * time.Second) // Allow Redis updates to process
	clicks := getClickCounts(t, testShortURL)
	// assert.Equal(t, 5, clicks.AllTime)
	assert.Equal(t, 5, clicks.Last1Min)

	// Step 4: Sleep 1 min, validate last_1min count is zero
	time.Sleep(1 * time.Minute)
	clicks = getClickCounts(t, testShortURL)
	assert.Equal(t, 0, clicks.Last1Min)

	// Step 5: Redirect once & validate last_1min count is 1
	http.Get(testShortURL)
	time.Sleep(2 * time.Second)
	clicks = getClickCounts(t, testShortURL)
	assert.Equal(t, 1, clicks.Last1Min)

	// Step 6: Sleep 1 min, validate last_1min count is 0
	time.Sleep(1 * time.Minute)
	clicks = getClickCounts(t, testShortURL)
	assert.Equal(t, 0, clicks.Last1Min)

	fmt.Println("Click tracking tests passed!")
}

// Helper function to get click counts
func getClickCounts(t *testing.T, shortURL string) ClickCounts {
	// Parse the short URL to extract the short code
	parsedURL, err := url.Parse(shortURL)
	if err != nil {
		t.Fatalf("Error parsing returned short URL: %v", err)
	}

	// Extract the short code from the URL path
	shortCode := strings.TrimPrefix(parsedURL.Path, "/api/v1/")

	// Construct the API request URL
	clicksAPIURL := fmt.Sprintf("%s/clicks/%s", baseAPI, shortCode)
	resp, err := http.Get(clicksAPIURL)

	// Handle HTTP request errors
	assert.NoError(t, err, " Error making request to clicks API")
	assert.Equal(t, http.StatusOK, resp.StatusCode, " Unexpected response status from clicks API")

	// Decode the response JSON
	var counts ClickCounts
	err = json.NewDecoder(resp.Body).Decode(&counts)
	assert.NoError(t, err, "Error decoding clicks API response")
	resp.Body.Close()

	// Print Click Counts
	fmt.Printf("Click counts → AllTime=%d, Last1Min=%d, Last24Hours=%d, LastWeek=%d\n",
		counts.AllTime, counts.Last1Min, counts.Last24Hours, counts.LastWeek)

	return counts
}

// Test 4: Create, Redirect, and Delete 5 Short URLs
func TestCreateRedirectDeleteURLsE2E(t *testing.T) {
	for i := 0; i < 5; i++ {
		longURL := fmt.Sprintf("https://example.com/testpage%d", i)

		requestBody := URLRequest{LongURL: longURL}
		jsonData, err := json.Marshal(requestBody)
		assert.NoError(t, err)

		resp, err := http.Post(baseAPI+"/create", "application/json", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response URLResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		resp.Body.Close()

		formattedShortURL := formatShortURL(response.ShortURL)
		testURLs[formattedShortURL] = longURL
	}

	// Step 3: Delete URLs & Confirm Deletion
	for shortURL := range testURLs {
		client := &http.Client{}
		log.Println("Deleting url:", shortURL)
		req, _ := http.NewRequest("DELETE", shortURL, nil)
		resp, _ := client.Do(req)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	}
}
