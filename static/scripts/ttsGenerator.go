package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	htgotts "github.com/hegedustibor/htgo-tts"
	"github.com/hegedustibor/htgo-tts/handlers"
)

// TTSConfig holds TTS generation configuration
type TTSConfig struct {
	Language string
	Voice    string
	Speed    float64
}

func main() {
	// Define command line flags
	outputDir := flag.String("output", ".", "Directory to save output files")
	filename := flag.String("filename", "output", "Base filename for output files (without extension)")
	textFile := flag.String("textfile", "input.txt", "Name of text file to read (in the output directory)")
	language := flag.String("lang", "en", "Language for TTS")
	voice := flag.String("voice", "", "Voice to use")
	speed := flag.Float64("speed", 1.0, "Speech speed")

	flag.Parse()

	// Ensure the output directory exists
	if err := ensureOutputDirectoryExists(*outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Get the full path to the text file
	textFilePath := filepath.Join(*outputDir, *textFile)

	// Read the text from the file
	text, err := readTextFromFile(textFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading text file: %v\n", err)
		os.Exit(1)
	}

	if text == "" {
		fmt.Println("Error: The text file is empty")
		os.Exit(1)
	}

	config := TTSConfig{
		Language: *language,
		Voice:    *voice,
		Speed:    *speed,
	}

	if err := GenerateTTSWithSubtitles(text, *outputDir, *filename, config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// readTextFromFile reads and returns the content of a text file
func readTextFromFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ensureDirectoryExists creates the directory if it doesn't exist
func ensureOutputDirectoryExists(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// GenerateTTSWithSubtitles creates an audio file from text and generates SRT subtitles
func GenerateTTSWithSubtitles(text, outputDir, filename string, config TTSConfig) error {
	// Create output directory if it doesn't exist
	if err := ensureOutputDirectoryExists(outputDir); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Set default language if not provided
	if config.Language == "" {
		config.Language = "en"
	}

	// Split text into sentences for better subtitle timing
	sentences := splitTextIntoSentences(text)
	if len(sentences) == 0 {
		return fmt.Errorf("no valid text to process")
	}

	// Generate TTS audio
	speech := htgotts.Speech{
		Folder:   outputDir,
		Language: config.Language,
		Handler:  &handlers.Native{},
	}

	audioFilePath := filepath.Join(outputDir, filename+".mp3")
	if err := speech.Speak(text); err != nil {
		return fmt.Errorf("TTS generation failed: %w", err)
	}

	// Generate SRT file with estimated timings
	srtFilePath := filepath.Join(outputDir, filename+".srt")
	if err := generateSRTFile(sentences, srtFilePath); err != nil {
		return fmt.Errorf("subtitle generation failed: %w", err)
	}

	fmt.Printf("TTS audio saved to: %s\n", audioFilePath)
	fmt.Printf("Subtitles saved to: %s\n", srtFilePath)

	return nil
}

// splitTextIntoSentences splits text into sentences for subtitle generation
func splitTextIntoSentences(text string) []string {
	// Simple sentence splitting (you might want to improve this)
	text = strings.ReplaceAll(text, ". ", ".|")
	text = strings.ReplaceAll(text, "! ", "!|")
	text = strings.ReplaceAll(text, "? ", "?|")

	sentences := strings.Split(text, "|")
	var result []string

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence != "" {
			result = append(result, sentence)
		}
	}

	return result
}

// generateSRTFile creates an SRT subtitle file with estimated timings
func generateSRTFile(sentences []string, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	const (
		wordsPerMinute = 150.0
		charPerWord    = 5.0
	)

	writer := bufio.NewWriter(file)
	startTime := 0.0

	for i, sentence := range sentences {
		// Estimate duration based on character count
		charCount := len(sentence)
		wordCount := float64(charCount) / charPerWord
		durationSeconds := (wordCount / wordsPerMinute) * 60.0

		// Minimum duration of 1 second per subtitle
		if durationSeconds < 1.0 {
			durationSeconds = 1.0
		}

		endTime := startTime + durationSeconds

		// Format start and end times for SRT (HH:MM:SS,mmm)
		startTimeFormatted := formatSRTTime(startTime)
		endTimeFormatted := formatSRTTime(endTime)

		// Write subtitle entry
		fmt.Fprintf(writer, "%d\n", i+1)
		fmt.Fprintf(writer, "%s --> %s\n", startTimeFormatted, endTimeFormatted)
		fmt.Fprintf(writer, "%s\n\n", sentence)

		startTime = endTime
	}

	return writer.Flush()
}

// formatSRTTime converts seconds to SRT time format (HH:MM:SS,mmm)
func formatSRTTime(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	milliseconds := int((seconds - float64(int(seconds))) * 1000)

	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, milliseconds)
}

// CombineAudioAndSubtitles uses FFmpeg to embed subtitles into a video
func CombineAudioAndSubtitles(videoPath, audioPath, subtitlePath, outputPath string) error {
	cmd := exec.Command(
		"ffmpeg",
		"-i", videoPath,
		"-i", audioPath,
		"-vf", fmt.Sprintf("subtitles=%s", subtitlePath),
		"-map", "0:v",
		"-map", "1:a",
		"-c:v", "libx264",
		"-c:a", "aac",
		"-shortest",
		outputPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
