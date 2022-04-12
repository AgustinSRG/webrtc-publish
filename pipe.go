// RTP -> Track pipe

package main

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/pion/webrtc/v3"
)

func pipeTrack(listener *net.UDPConn, track *webrtc.TrackLocalStaticRTP) {
	inboundRTPPacket := make([]byte, 1600) // UDP MTU
	for {
		n, _, err := listener.ReadFrom(inboundRTPPacket)
		if err != nil {
			panic(fmt.Sprintf("error during read: %s", err))
		}

		if _, err = track.Write(inboundRTPPacket[:n]); err != nil {
			if errors.Is(err, io.ErrClosedPipe) {
				// The peerConnection has been closed.
				return
			}

			panic(err)
		}
	}
}

const SENDER_READ_BUFFER_LENGTH = 1500

// Read incoming RTCP packets
// Before these packets are returned they are processed by interceptors. For things
// like NACK this needs to be called.
func readPacketsFromRTPSender(sender *webrtc.RTPSender) {
	rtcpBuf := make([]byte, SENDER_READ_BUFFER_LENGTH)
	for {
		if _, _, rtcpErr := sender.Read(rtcpBuf); rtcpErr != nil {
			return
		}
	}
}
