# WebRTC Publisher

Utility to publish to [webrtc-cdn](https://github.com/AgustinSRG/webrtc-cdn) using a video source, or an RTMP source.

It uses [FFMpeg](https://ffmpeg.org/) to convert the source to RTP, and then the [pion/webrtc](https://github.com/pion/webrtc) library to convert it to WebRTC.

## Compilation

In order to install dependencies, type:

```
go get github.com/AgustinSRG/webrtc-cdn
```

To compile the code type:

```
go build
```

The build command will create a binary in the currenct directory, called `webrtc-publish`, or `webrtc-publish.exe` if you are using Windows.

## Usage

You can use the program from the command line:

```
webrtc-publisher [OPTIONS] <SOURCE> <DESTINATION>
```

### SOURCE

The source can be a path to a video file, or an RTMP URL. Examples:

 - `./path/to/video.mp4`
 - `rtmp://localhost/test/key`

### DESTINATION

The destination must be a websocket URL of one of the webrtc-cdn nodes. Examples:

 - `ws://localhost/stream-id`
 - `wss://www.example.com/stream-id`

### OPTIONS

Here is a list of all the options:

| Option | Description |
|---|---|
| `--help, -h` | Shows the command line options |
| `--version, -v` | Shows the version |
| `--debug` | Enables debug mode (prints more messages) |
| `--ffmpeg-path <path>` | Sets the FFMpeg path. By default is `/usr/bin/ffmpeg`. You can also change it with the environment variable `FFMPEG_PATH` |
| `--loop, -l` | Enables loop (for video files). |
| `--auth, -a <auth-token>` | Sets auth token. |
| `--secret, -s <secret>` | Provides secret to generate authentication tokens. |
