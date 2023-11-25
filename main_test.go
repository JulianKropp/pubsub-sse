package pubsubsse

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"
)

// Tests helper for the outher test files.

// -----------------------------
// General
// -----------------------------

// Start /event server
func startEventServer(ssePubSub *sSEPubSubService, t *testing.T, port int) {
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
func httpToEvent(t *testing.T, client *client, port int, connected, done chan bool, returnValue *[]eventData) {
	hclient := http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:"+strconv.Itoa(port)+"/event?client_id="+client.GetID(), nil)
	if err != nil {
		t.Error(err)
		return
	}

	resp, err := hclient.Do(req)
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
			fmt.Printf("%s message: %s\n", client.GetID(), line)
			stream <- string(line)
		}
	}()

	// Listen for messages and timeout
	go func() {
		for {
			select {
			case <-timeout:
				fmt.Println("timeout")
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
					fmt.Printf("ok: %s\n", client.GetID())
				}

				var rvalue eventData
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