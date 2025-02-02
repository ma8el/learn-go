package ollamastream

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// requestBody matches Ollama's /generate JSON request body structure.
type requestBody struct {
	Prompt  string `json:"prompt"`
	Model   string `json:"model,omitempty"`
	Options struct {
		Temperature float64 `json:"temperature,omitempty"`
		MaxTokens   int     `json:"max_tokens,omitempty"`
	} `json:"options,omitempty"`
}

// ollamaResponse is each line of Ollama's streaming JSON output.
type ollamaResponse struct {
	Response string `json:"response,omitempty"`
	Done     bool   `json:"done,omitempty"`
}

// GenerateStream sends a prompt to an Ollama server and streams tokens via onToken callback.
//
//   - ctx:        A context for cancellation.
//   - prompt:     The text prompt to generate from.
//   - endpoint:   The Ollama server endpoint (e.g., "http://localhost:11411/generate").
//   - model:      Name of the model to use (if not already specified at serve time).
//   - temperature A float that controls creativity (0.0 = deterministic, 1.0 = creative).
//   - maxTokens:  The maximum tokens for this completion.
//   - onToken:    A callback that receives each chunk of text as it arrives.
//
// Returns an error if the request fails or if there's an I/O issue streaming.
func GenerateStream(
	ctx context.Context,
	prompt string,
	endpoint string,
	model string,
	temperature float64,
	maxTokens int,
	onToken func(string),
) error {
	// Build the JSON request body
	reqData := requestBody{
		Prompt: prompt,
		Model:  model,
	}
	reqData.Options.Temperature = temperature
	reqData.Options.MaxTokens = maxTokens

	reqBytes, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(reqBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama returned non-200 status %d: %s", resp.StatusCode, body)
	}

	// Stream the response line by line
	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading response: %w", err)
		}

		// Each line is a JSON object with fields "response" and "done"
		var partial ollamaResponse
		if err := json.Unmarshal(line, &partial); err != nil {
			// Some lines may be empty or not valid JSON. Skip errors on those.
			continue
		}

		// Send partial text to our callback
		if partial.Response != "" && onToken != nil {
			onToken(partial.Response)
		}

		if partial.Done {
			break
		}
	}

	return nil
}
