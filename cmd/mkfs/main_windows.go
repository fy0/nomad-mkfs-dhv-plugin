//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type StdErrResponse struct {
	Error string `json:"error"`
}

func main() {
	resp := StdErrResponse{
		Error: "mkfs dynamic host volume plugin is not supported on Windows; it relies on Unix-style mkfs, mount, and filesystem tooling",
	}

	if payload, err := json.Marshal(resp); err == nil {
		fmt.Print(string(payload))
	} else {
		fmt.Fprint(os.Stderr, resp.Error)
	}

	os.Exit(1)
}
