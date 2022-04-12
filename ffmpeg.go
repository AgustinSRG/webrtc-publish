// FFMPEG

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func runEncdingProcess(ffmpegBin string, source string, videoUDP string, audioUDP string, debug bool) {
	cmd := exec.Command(ffmpegBin,
		"-re",
		"-i", source,
		// AUDIO
		"-an",
		"-c:a", "libopus",
		"-sample_fmt", "s16p",
		"-ssrc", "1",
		"-payload_type", "111",
		"-max_delay", "0",
		"-application", "lowdelay",
		"-f", "rtp", "rtp://"+audioUDP+"?pkt_size=1200",
		// VIDEO
		"-vn",
		"-vcodec", "libvpx",
		"-deadline", "1",
		"-g", "10",
		"-error-resilient", "1",
		"-auto-alt-ref", "1",
		"-f", "rtp", "rtp://"+videoUDP+"?pkt_size=1200")

	if debug {
		cmd.Stderr = os.Stderr
		fmt.Println("Running command: " + cmd.String())
	}

	err := cmd.Run()

	if err != nil {
		fmt.Println("Error: ffmpeg program failed: " + err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
