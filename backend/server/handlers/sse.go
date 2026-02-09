package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SendSSE writes a single Server-Sent Event to w.
// eventType becomes the "event:" line; data is JSON-marshalled into the "data:" line.
func SendSSE(w http.ResponseWriter, eventType string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: %s\n", eventType)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
}
