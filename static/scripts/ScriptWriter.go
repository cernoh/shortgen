package main

// go run ScriptWriter.go -prompt "Your concept or idea here" -output "/path/to/output/directory"
import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the configuration structure
type ScriptWriterConfig struct {
	DeepSeekAPIKey string `toml:"deepseekapikey"`
}

// DeepSeekRequest represents the request structure for the DeepSeek API
type DeepSeekRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

// Message represents a message in the conversation with DeepSeek
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekResponse represents the response from the DeepSeek API
type DeepSeekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func scriptWriterMain() {
	// Parse command-line arguments
	prompt := flag.String("prompt", "", "Text prompt to generate scripts from")
	outputDir := flag.String("output", ".", "Directory to save the generated scripts")
	flag.Parse()

	if *prompt == "" {
		fmt.Println("Error: Please provide a text prompt using the -prompt flag")
		flag.Usage()
		os.Exit(1)
	}

	// Read config file
	configPath := "config.toml"
	var config ScriptWriterConfig
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		os.Exit(1)
	}

	if config.DeepSeekAPIKey == "" {
		fmt.Println("Error: DeepSeek API key not found in config.toml")
		os.Exit(1)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Generate first script
	script1, err := generateScript(*prompt, "Create a detailed script for a short video based on this concept: ", config.DeepSeekAPIKey)
	if err != nil {
		fmt.Printf("Error generating first script: %v\n", err)
		os.Exit(1)
	}

	// Generate second script with different instructions
	script2, err := generateScript(*prompt, "Create an alternative creative script for a short video using this concept: ", config.DeepSeekAPIKey)
	if err != nil {
		fmt.Printf("Error generating second script: %v\n", err)
		os.Exit(1)
	}

	// Save scripts to files
	script1File := filepath.Join(*outputDir, "script1.txt")
	script2File := filepath.Join(*outputDir, "script2.txt")

	if err := ioutil.WriteFile(script1File, []byte(script1), 0644); err != nil {
		fmt.Printf("Error saving script1.txt: %v\n", err)
		os.Exit(1)
	}

	if err := ioutil.WriteFile(script2File, []byte(script2), 0644); err != nil {
		fmt.Printf("Error saving script2.txt: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully generated and saved two scripts:")
	fmt.Printf("1. %s\n", script1File)
	fmt.Printf("2. %s\n", script2File)
}

// generateScript calls the DeepSeek API to generate a script based on the provided prompt
func generateScript(basePrompt, instruction, apiKey string) (string, error) {
	fullPrompt := instruction + basePrompt

	// Prepare DeepSeek API request
	req := DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a creative script writer for short videos. Create engaging, detailed scripts with scene descriptions and dialogue.",
			},
			{
				Role:    "user",
				Content: fullPrompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   2000,
	}

	// Convert request to JSON
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("error sending request to DeepSeek API: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse response
	var deepSeekResp DeepSeekResponse
	if err := json.NewDecoder(resp.Body).Decode(&deepSeekResp); err != nil {
		return "", fmt.Errorf("error parsing DeepSeek response: %w", err)
	}

	// Check for API errors
	if deepSeekResp.Error != nil {
		return "", fmt.Errorf("DeepSeek API error: %s", deepSeekResp.Error.Message)
	}

	// Check if we have valid choices
	if len(deepSeekResp.Choices) == 0 {
		return "", fmt.Errorf("no script generated by DeepSeek")
	}

	return deepSeekResp.Choices[0].Message.Content, nil
}
