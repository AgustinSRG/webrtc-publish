// Main

package main

import (
	"fmt"
	"net/url"
	"os"
)

// Program entry point
func main() {
	// Read env vars
	ffmpegPath := os.Getenv("FFMPEG_PATH")

	if ffmpegPath == "" {
		ffmpegPath = "/usr/bin/ffmpeg"
	}

	// Read arguments
	args := os.Args

	if len(args) < 3 {
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			printHelp()
		} else if len(args) > 1 && (args[1] == "--version" || args[1] == "-v") {
			printVersion()
		} else {
			printHelp()
		}
		return
	}

	source := args[len(args)-2]

	if _, err := os.Stat(source); err != nil {
		// Not a file
		u, err := url.Parse(source)
		if err != nil || (u.Scheme != "rtmp" && u.Scheme != "rtmps") {
			fmt.Println("The source is not a valid file or RTMP URL")
			return
		}
	}

	destination := args[len(args)-1]

	u, err := url.Parse(destination)
	if err != nil || (u.Scheme != "ws" && u.Scheme != "wss") {
		fmt.Println("The destination is not a valid websocket URL")
		return
	}

	protocol := u.Scheme
	host := u.Host
	streamId := ""

	if len(u.Path) > 0 {
		streamId = u.Path[1:]
	} else {
		fmt.Println("The destination URL must contain the stream ID. Example: ws://localhost/stream-id")
	}

	wsURL := url.URL{
		Scheme: protocol,
		Host:   host,
		Path:   "/ws",
	}

	loop := false
	debug := false
	authToken := ""

	for i := 1; i < (len(args) - 2); i++ {
		arg := args[i]

		if arg == "--debug" {
			debug = true
		} else if arg == "--ffmpeg-path" {
			if i == len(args)-3 {
				fmt.Println("The option '--ffmpeg-path' requires a value")
				return
			}
			ffmpegPath = args[i+1]
			i++
		} else if arg == "--loop" || arg == "-l" {
			loop = true
		} else if arg == "--auth" || arg == "-a" {
			if i == len(args)-3 {
				fmt.Println("The option '--auth' requires a value")
				return
			}
			authToken = args[i+1]
			i++
		} else if arg == "--secret" || arg == "-s" {
			if i == len(args)-3 {
				fmt.Println("The option '--secret' requires a value")
				return
			}
			authToken = generateToken(args[i+1], streamId)
			i++
		}
	}

	if _, err := os.Stat(ffmpegPath); err != nil {
		fmt.Println("Error: Could not find 'ffmpeg' at specified location: " + ffmpegPath)
		return
	}

	runPublish(source, wsURL, streamId, PublishOptions{
		loop:      loop,
		debug:     debug,
		ffmpeg:    ffmpegPath,
		authToken: authToken,
	})
}

func printHelp() {
	fmt.Println("Usage: webrtc-publisher [OPTIONS] <SOURCE> <DESTINATION>")
	fmt.Println("    SOURCE: Can be a path to a video file or RTMP URL")
	fmt.Println("    DESTINATION: Websocket URL like ws(s)://host:port/stream-id")
	fmt.Println("    OPTIONS:")
	fmt.Println("        --help, -h                 Prints command line options.")
	fmt.Println("        --version, -v              Prints version.")
	fmt.Println("        --debug                    Enables debug mode.")
	fmt.Println("        --ffmpeg-path <path>       Sets FFMpeg path.")
	fmt.Println("        --loop, -l                 Enables loop (for videos).")
	fmt.Println("        --auth, -a <auth-token>    Sets authentication token for publishing.")
	fmt.Println("        --secret, -s <secret>      Sets secret to generate authentication tokens.")
}

func printVersion() {
	fmt.Println("webrtc-publisher 1.0.0")
}

func killProcess() {
	os.Exit(0)
}
