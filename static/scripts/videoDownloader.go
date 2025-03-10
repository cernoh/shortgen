package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// Config structure to hold settings from config.toml
type Config struct {
	PexelsAPIKeys []string `toml:"plexelsapikeys"`
}

var (
	apiKeys = []string{} // Will be populated from config file
	baseURL = "https://api.pexels.com/videos/search"
)

type Video struct {
	ID         int         `json:"id"`
	URL        string      `json:"url"`
	User       User        `json:"user"`
	VideoFiles []VideoFile `json:"video_files"`
}

type User struct {
	Name string `json:"name"`
}

type VideoFile struct {
	Link string `json:"link"`
}

type PexelsResponse struct {
	Videos []Video `json:"videos"`
}

// makeRequestWithRetry tries each API key until one succeeds
func makeRequestWithRetry(url string) (*http.Response, error) {
	// Check if we have any API keys
	if len(apiKeys) == 0 {
		return nil, fmt.Errorf("no API keys configured")
	}

	// Make a copy of the API keys to shuffle them
	keys := make([]string, len(apiKeys))
	copy(keys, apiKeys)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })

	var lastError error

	for i, key := range keys {
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			lastError = err
			continue
		}

		req.Header.Add("Authorization", key)

		// Display masked API key for debugging
		maskedKey := maskAPIKey(key)
		fmt.Printf("Trying API key %d: %s\n", i+1, maskedKey)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("API key %d failed: %v\n", i+1, err)
			lastError = err
			continue
		}

		if resp.StatusCode == http.StatusOK {
			fmt.Printf("Successfully connected using API key %d\n", i+1)
			return resp, nil
		}

		fmt.Printf("API key %d failed with status code: %d\n", i+1, resp.StatusCode)
		resp.Body.Close()
		lastError = fmt.Errorf("received status code: %d", resp.StatusCode)
	}

	return nil, fmt.Errorf("all API keys failed: %v", lastError)
}

// maskAPIKey hides most of the API key for security in logs
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func videoDownloaderMain() {
	// First, load the configuration
	if err := loadConfig(); err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}

	if len(apiKeys) == 0 {
		fmt.Println("No Pexels API keys found in configuration")
		return
	}

	// Check if required arguments are provided
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run videoDownloader.go keywords.txt output_directory")
		return
	}

	keywordsFile := os.Args[1]
	outputDir := os.Args[2]

	// Create output directory with 'plexels-videos' subfolder
	videosDir := filepath.Join(outputDir, "plexels-videos")
	if err := ensureDirectoryExists(videosDir); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	fmt.Printf("Videos will be saved to: %s\n", videosDir)

	// Read keywords from the provided file
	keywords, err := readKeywordsFromFile(keywordsFile)
	if err != nil {
		fmt.Println("Error reading keywords file:", err)
		return
	}

	if len(keywords) == 0 {
		fmt.Println("No keywords found in the file")
		return
	}

	fmt.Printf("Using keywords: %s\n", keywords)
	perPage := 1 // Number of videos to download
	url := fmt.Sprintf("%s?query=%s&per_page=%d", baseURL, keywords, perPage)

	// Try to make a request with API key rotation
	resp, err := makeRequestWithRetry(url)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	var pexelsResp PexelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&pexelsResp); err != nil {
		fmt.Println("Error decoding response:", err)
		return
	}

	if len(pexelsResp.Videos) == 0 {
		fmt.Println("No videos found for the given keywords.")
		return
	}

	for _, video := range pexelsResp.Videos {
		fmt.Printf("Downloading video by %s...\n", video.User.Name)

		// Find the first available video file or use the first one if all fail
		videoLink := ""
		if len(video.VideoFiles) > 0 {
			videoLink = video.VideoFiles[0].Link
		} else {
			fmt.Println("No video files found for this video.")
			continue
		}

		outputPath := filepath.Join(videosDir, fmt.Sprintf("video_%d.mp4", video.ID))
		if err := downloadVideo(videoLink, outputPath); err != nil {
			fmt.Println("Error downloading video:", err)
		} else {
			fmt.Println("Download completed successfully.")
		}
	}
}

// ensureDirectoryExists creates a directory if it doesn't exist
func ensureDirectoryExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Printf("Creating directory: %s\n", dirPath)
		return os.MkdirAll(dirPath, 0755)
	} else if err != nil {
		return err
	}
	return nil
}

// readKeywordsFromFile reads keywords from a file and returns them as a single string
func readKeywordsFromFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var keywords []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			keywords = append(keywords, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(keywords, " "), nil
}

func downloadVideo(url, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// loadConfig loads configuration from the config file
func loadConfig() error {
	configPath := "config.toml"

	// Check if config path is provided as an argument
	if len(os.Args) > 3 {
		configPath = os.Args[3]
	}

	// Try to find config in default locations if not specified
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Try home directory
		homeDir, err := os.UserHomeDir()
		if err == nil {
			altPath := filepath.Join(homeDir, ".config", "shortgen", "config.toml")
			if _, err := os.Stat(altPath); err == nil {
				configPath = altPath
			}
		}
	}

	fmt.Printf("Loading configuration from: %s\n", configPath)

	var config Config
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		return fmt.Errorf("failed to decode config file: %v", err)
	}

	// Set apiKeys from config
	apiKeys = config.PexelsAPIKeys
	fmt.Printf("Loaded %d Pexels API keys\n", len(apiKeys))

	return nil
}
