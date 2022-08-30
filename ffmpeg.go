// FFMPEG

package main

import (
	"fmt"
	"os"
	"os/exec"

	child_process_manager "github.com/AgustinSRG/go-child-process-manager"
)

func runEncdingProcess(ffmpegBin string, source string, videoUDP string, audioUDP string, debug bool, loop bool) {
	args := make([]string, 1)

	args[0] = ffmpegBin

	args = append(args, "-re")

	if loop {
		args = append(args, "-stream_loop", "-1")
	}

	// INPUT
	args = append(args, "-i", source)

	// AUDIO
	args = append(args,
		"-vn",
		"-acodec", "libopus",
		"-ssrc", "1",
		"-payload_type", "111",
		"-max_delay", "0",
		"-application", "lowdelay",
		"-f", "rtp", "rtp://"+audioUDP+"?pkt_size=1200",
	)

	// VIDEO
	args = append(args,
		"-an",
		"-vcodec", "libvpx",
		"-cpu-used", "5",
		"-deadline", "1",
		"-g", "10",
		"-error-resilient", "1",
		"-auto-alt-ref", "1",
		"-f", "rtp", "rtp://"+videoUDP+"?pkt_size=1200",
	)

	cmd := exec.Command(ffmpegBin)
	cmd.Args = args

	if debug {
		cmd.Stderr = os.Stderr
		fmt.Println("Running command: " + cmd.String())
	}

	child_process_manager.ConfigureCommand(cmd)

	err := cmd.Start()

	if err != nil {
		fmt.Println("Error: ffmpeg program failed: " + err.Error())
		os.Exit(1)
	}

	child_process_manager.AddChildProcess(cmd.Process)

	err = cmd.Wait()

	if err != nil {
		fmt.Println("Error: ffmpeg program failed: " + err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
