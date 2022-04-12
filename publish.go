// Publishing script

package main

import "net/url"

type PublishOptions struct {
	loop      bool
	debug     bool
	ffmpeg    string
	authToken string
}

func runPublish(source string, destination url.URL, streamId string, options PublishOptions) {

}
