package render

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

var jsonContentType = []string{"application/json; charset=utf-8"}

// JSON common json struct.
type JSON struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Now     int64         `json:"now"`
	Data    interface{} `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, obj interface{}) (err error) {
	var jsonBytes []byte
	writeContentType(w, jsonContentType)
	if jsonBytes, err = json.Marshal(obj); err != nil {
		err = errors.WithStack(err)
		return
	}
	if _, err = w.Write(jsonBytes); err != nil {
		err = errors.WithStack(err)
	}
	return
}

// Render (JSON) writes data with json ContentType.
func (r JSON) Render(w http.ResponseWriter) error {
	if r.Now <= 0 {
		r.Now = time.Now().Unix()
	}
	return writeJSON(w, r)
}

// WriteContentType write json ContentType.
func (r JSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}

// MapJSON common map json struct.
type MapJSON map[string]interface{}

// Render (MapJSON) writes data with json ContentType.
func (m MapJSON) Render(w http.ResponseWriter) error {
	return writeJSON(w, m)
}

// WriteContentType write json ContentType.
func (m MapJSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}
