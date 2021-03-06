package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

type channelData struct {
	id   int
	data string
}

// Check if the server is in the valid server list passed at startup
func isValidServer(server string) bool {
	for _, validServer := range setting.servers {
		if validServer == server {
			return true
		}
	}
	return false
}

// Send commands to lgproxy instances in parallel, and retrieve their responses
func batchRequest(servers []string, endpoint string, command string) []string {
	// Channel and array for storing responses
	var ch chan channelData = make(chan channelData)
	var responseArray []string = make([]string, len(servers))

	for i, server := range servers {
		if !isValidServer(server) {
			// If the server is not valid, create a dummy goroutine to return a failure
			go func(i int) {
				ch <- channelData{i, "request failed: invalid server\n"}
			}(i)
		} else {
			// Compose URL and send the request
			url := "http://" + server + "." + setting.domain + ":" + strconv.Itoa(setting.proxyPort) + "/" + url.PathEscape(endpoint) + "?q=" + url.QueryEscape(command)
			go func(url string, i int) {
				response, err := http.Get(url)
				if err != nil {
					ch <- channelData{i, "request failed: " + err.Error() + "\n"}
					return
				}
				text, _ := ioutil.ReadAll(response.Body)
				ch <- channelData{i, string(text)}
			}(url, i)
		}
	}

	// Sort the responses by their ids, to return data in order
	for range servers {
		var output channelData = <-ch
		responseArray[output.id] = output.data
		if len(responseArray[output.id]) == 0 {
			responseArray[output.id] = "node returned empty response, please refresh to try again."
		}
	}

	return responseArray
}
