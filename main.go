// Main

package main

import "os"

// Program entry point
func main() {
	// Read env vars
	ffmpegPath := os.Getenv("FFMPEG_PATH")

	if ffmpegPath == "" {
		ffmpegPath = "/usr/bin/ffmpeg"
	}

	// Read arguments

}
