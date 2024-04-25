package pubsubsse

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Tests helper for the outher test files.

// -----------------------------
// General
// -----------------------------

// Start /event server
func startEventServer(ssePubSub *SSEPubSubService, t *testing.T, port int) {
	srv := http.NewServeMux()
	srv.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) { Event(ssePubSub, w, r) }) // Event SSE endpoint
	go func() {
		errx := http.ListenAndServe(":"+strconv.Itoa(port), srv)
		if errx != nil {
			t.Error("http Listen err", errx)
		}
	}()
	time.Sleep(100 * time.Millisecond)
}

// Makes an http request to localhost:8080/event
// This will be an SSE connection and will be open for 10s
func httpToEvent(t *testing.T, instance *Instance, port int, connected, done chan bool, returnValue *[]connectionData) {
	hinstance := http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:"+strconv.Itoa(port)+"/event?instance_id="+instance.GetID(), nil)
	if err != nil {
		t.Error(err)
		return
	}

	resp, err := hinstance.Do(req)
	if err != nil {
		t.Error(err)
		return
	}

	// Set a timeout for the SSE connection
	timeout := time.After(5 * time.Second)
	stream := make(chan string, 100)

	con := false

	// Goroutine to read from the SSE stream
	go func() {
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				close(stream)
				t.Logf("Error reading from SSE stream: %s", err.Error())
				return
			}
			t.Logf("%s message: %s\n", instance.GetID(), line)
			stream <- string(line)
		}
	}()

	// Listen for messages and timeout
	go func() {
		for {
			select {
			case <-timeout:
				t.Logf("timeout\n")
				// Send connected signal
				if !con {
					con = true
					connected <- true
				}
				done <- true
				resp.Body.Close()
				return
			case message, ok := <-stream:
				if !ok {
					// Stream closed, exit loop
					return
				}

				// Send connected signal
				if !con {
					con = true
					connected <- true
					t.Logf("ok: %s\n", instance.GetID())
				}

				// Remove the "data: " from message
				message = strings.TrimPrefix(message, "data: ")

				if message == "\n" {
					continue
				}

				var rvalue connectionData
				// Unmarshal the JSON data
				err = json.Unmarshal([]byte(message), &rvalue)
				if err != nil {
					t.Error(err)
					continue
				}

				*returnValue = append(*returnValue, rvalue)
			}
		}
	}()
}
