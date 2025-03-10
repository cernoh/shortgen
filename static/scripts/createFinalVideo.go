package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func finalVideoMain() {
	// Check if we have the right number of arguments
	if len(os.Args) != 3 {
		fmt.Println("Usage: createFinalVideo <input_directory> <filename>")
		os.Exit(1)
	}

	inputDir := os.Args[1]
	filename := os.Args[2]

	// Run ttsgenerator.go with input directory and filename
	fmt.Println("Running TTS generator...")
	ttsCmd := exec.Command("go", "run", "ttsgenerator.go", inputDir, filename)
	ttsCmd.Stdout = os.Stdout
	ttsCmd.Stderr = os.Stderr

	if err := ttsCmd.Run(); err != nil {
		fmt.Printf("Error running TTS generator: %v\n", err)
		os.Exit(1)
	}

	// Path to keywords.txt
	keywordsPath := filepath.Join(inputDir, "keywords.txt")

	// Check if keywords.txt exists
	if _, err := os.Stat(keywordsPath); os.IsNotExist(err) {
		fmt.Printf("Error: keywords.txt not found at %s\n", keywordsPath)
		os.Exit(1)
	}

	// Run video downloader with keywords.txt
	fmt.Println("Running video downloader...")
	videoCmd := exec.Command("go", "run", "videodownloader.go", keywordsPath)
	videoCmd.Stdout = os.Stdout
	videoCmd.Stderr = os.Stderr

	if err := videoCmd.Run(); err != nil {
		fmt.Printf("Error running video downloader: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Video creation process completed successfully!")
}
