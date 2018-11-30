// Package ghooks is a github hooks receiver in golang with net/http.
/*
Inspired by GitHub::Hooks::Receiver (https://github.com/Songmu/Github-Hooks-Receiver),
and octoks (https://github.com/hisaichi5518/octoks)

Install

		go get github.com/erikh/ghooks

Usage

		// sample.go
		package main

		import (
				"fmt"
				"log"

				"github.com/erikh/ghooks"
		)


		func main() {
				port := 8080
				hooks := ghooks.NewServer(port)

				hooks.On("push", pushHandler)
				hooks.On("pull_request", pullRequestHandler)
				hooks.Run()
		}

		func pushHandler(payload interface{}) {
				fmt.Println("puuuuush")
		}

		func pullRequestHandler(payload interface{}) {
				fmt.Println("pull_request")
		}

After starting this server:

		curl -H "X-GitHub-Event: push" -d '{"hoge":"fuga"}' http://localhost:8080
		> puuuuush
*/
package ghooks

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	// Version is the version of the library
	Version = 0.2
)

// Server is the hander pipeline for serving github hooks. One can instantiate a server
type Server struct {
	Secret string
}

// Hook is the type of hook event being processed.
type Hook struct {
	Event string
	Func  func(payload interface{})
}

// Hooks is a list of Hook.
type Hooks []Hook

var hooks Hooks

// On is the primary registration mechanism for handlers. Handlers must accept
// interface{} and process the resulting data (typically some form of
// map[string]interface{}) in its handler after receiving the hook.
func (s *Server) On(name string, handler func(payload interface{})) {
	hooks = append(hooks, Hook{Event: name, Func: handler})
}

func emit(name string, payload interface{}) {
	for _, v := range hooks {
		if strings.EqualFold(v.Event, name) {
			v.Func(payload)
		}
	}
}

// NewServer creates a new *Server.
func NewServer() *Server {
	return &Server{}
}

// Handler is the primary handler returned by the server. You can leverage it
// with the http library like so:
//
//		s := NewServer()
//		http.HandleFunc("/", s.Handler)
//		http.ListenAndServe(":2222", http.DefaultServeMux)
//
func (s *Server) Handler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		http.Error(w, "Method Not Allowd", http.StatusMethodNotAllowed)
		return
	}

	event := req.Header.Get("X-GitHub-Event")

	if event == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if req.Body == nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if s.Secret != "" {
		signature := req.Header.Get("X-Hub-Signature")
		if !s.isValidSignature(body, signature) {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	}

	var payload interface{}
	var decoder *json.Decoder

	if strings.Contains(req.Header.Get("Content-Type"), "application/json") {

		decoder = json.NewDecoder(bytes.NewReader(body))

	} else if strings.Contains(req.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {

		v, err := url.ParseQuery(string(body))
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		p := v.Get("payload")
		decoder = json.NewDecoder(strings.NewReader(p))
	}

	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	emit(event, payload)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) isValidSignature(body []byte, signature string) bool {
	if !strings.HasPrefix(signature, "sha1=") {
		return false
	}

	mac := hmac.New(sha1.New, []byte(s.Secret))
	mac.Write(body)
	actual := mac.Sum(nil)

	expected, err := hex.DecodeString(signature[5:])
	if err != nil {
		return false
	}

	return hmac.Equal(actual, expected)
}
