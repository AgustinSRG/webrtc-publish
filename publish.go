// Publishing script

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type PublishOptions struct {
	loop      bool
	debug     bool
	ffmpeg    string
	authToken string
}

func runPublish(source string, destination url.URL, streamId string, options PublishOptions) {
	// Create UDP listeners
	listenerAudio, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		panic(err)
	}

	if options.debug {
		fmt.Println("UDP Listener openned for audio: " + fmt.Sprint(listenerAudio.LocalAddr().String()))
	}

	listenerVideo, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		panic(err)
	}

	if options.debug {
		fmt.Println("UDP Listener openned for video: " + fmt.Sprint(listenerVideo.LocalAddr().String()))
	}

	// Create tracks

	videoTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video", "pion")
	if err != nil {
		panic(err)
	}

	audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
	if err != nil {
		panic(err)
	}

	// Create peer connection
	peerConnectionConfig := loadWebRTCConfig() // Load config
	peerConnection, err := webrtc.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return
	}

	// Mutex
	lock := sync.Mutex{}

	// Connect to websocket
	if options.debug {
		fmt.Println("Connecting to " + destination.String())
	}
	c, _, err := websocket.DefaultDialer.Dial(destination.String(), nil)
	if err != nil {
		fmt.Println("Error: " + err.Error())
	}
	defer c.Close()

	// Send publish message
	pubMsg := SignalingMessage{
		method: "PUBLISH",
		params: make(map[string]string),
		body:   "",
	}
	pubMsg.params["Request-ID"] = "pub01"
	pubMsg.params["Stream-ID"] = streamId
	if options.authToken != "" {
		pubMsg.params["Auth"] = options.authToken
	}
	c.WriteMessage(websocket.TextMessage, []byte(pubMsg.serialize()))

	if options.debug {
		fmt.Println(">>>\n" + string(pubMsg.serialize()))
	}

	// ICE Candidate handler
	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		lock.Lock()
		defer lock.Unlock()

		candidateMsg := SignalingMessage{
			method: "CANDIDATE",
			params: make(map[string]string),
			body:   "",
		}
		candidateMsg.params["Request-ID"] = "pub01"
		if i != nil {
			b, e := json.Marshal(i.ToJSON())
			if e != nil {
				fmt.Println("Error: " + e.Error())
			} else {
				candidateMsg.body = string(b)
			}
		}

		c.WriteMessage(websocket.TextMessage, []byte(candidateMsg.serialize()))
	})

	// Connection status handler
	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		lock.Lock()
		defer lock.Unlock()

		if state == webrtc.PeerConnectionStateClosed || state == webrtc.PeerConnectionStateFailed {
			fmt.Println("WebRTC: Disconnected")
		} else if state == webrtc.PeerConnectionStateConnected {
			fmt.Println("WebRTC: Connected")
		}
	})

	receivedOffer := false

	// Read websocket messages
	for {
		func() {
			lock.Lock()
			defer lock.Unlock()

			_, message, err := c.ReadMessage()
			if err != nil {
				os.Exit(0)
				return // Closed
			}

			if options.debug {
				fmt.Println("<<<\n" + string(message))
			}

			msg := parseSignalingMessage(string(message))

			if msg.method == "ERROR" {
				fmt.Println("Error: " + msg.params["error-message"])
				os.Exit(1)
			} else if msg.method == "OFFER" {
				if !receivedOffer {
					receivedOffer = true

					// Set remote rescription

					sd := webrtc.SessionDescription{}

					err := json.Unmarshal([]byte(msg.body), &sd)

					if err != nil {
						fmt.Println("Error: " + err.Error())
					}

					err = peerConnection.SetRemoteDescription(sd)

					if err != nil {
						fmt.Println("Error: " + err.Error())
					}

					// Add tracks

					audioSender, err := peerConnection.AddTrack(audioTrack)
					if err != nil {
						fmt.Println("Error: " + err.Error())
					}

					go readPacketsFromRTPSender(audioSender)

					videoSender, err := peerConnection.AddTrack(videoTrack)
					if err != nil {
						fmt.Println("Error: " + err.Error())
					}

					go readPacketsFromRTPSender(videoSender)

					// Generate answer
					answer, err := peerConnection.CreateAnswer(nil)
					if err != nil {
						fmt.Println("Error: " + err.Error())
					}

					// Sets the LocalDescription, and starts our UDP listeners
					err = peerConnection.SetLocalDescription(answer)
					if err != nil {
						fmt.Println("Error: " + err.Error())
					}

					// Send ANSWER to the client

					answerJSON, e := json.Marshal(answer)

					if e != nil {
						fmt.Println("Error: " + err.Error())
					}

					answerMsg := SignalingMessage{
						method: "ANSWER",
						params: make(map[string]string),
						body:   string(answerJSON),
					}
					answerMsg.params["Request-ID"] = "pub01"

					c.WriteMessage(websocket.TextMessage, []byte(answerMsg.serialize()))

					// Pipe tracks and start FFMPEG
					go pipeTrack(listenerAudio, audioTrack)
					go pipeTrack(listenerVideo, videoTrack)

					go runEncdingProcess(options.ffmpeg, source, listenerVideo.LocalAddr().String(), listenerAudio.LocalAddr().String(), options.debug)
				}
			} else if msg.method == "CANDIDATE" {
				if receivedOffer {
					candidate := webrtc.ICECandidateInit{}

					err := json.Unmarshal([]byte(msg.body), &candidate)

					if err != nil {
						fmt.Println("Error: " + err.Error())
					}

					err = peerConnection.AddICECandidate(candidate)

					if err != nil {
						fmt.Println("Error: " + err.Error())
					}
				}
			} else if msg.method == "CLOSE" {
				fmt.Println("Connection closed by remote host.")
				os.Exit(0)
			}
		}()
	}
}
