package rpc_test

import (
	"proof/rpc"
	"testing"
)

type EncodingExample struct {
	Method bool
}

func TestEncode(t *testing.T) {
	expected := "Content-Length: 15\r\n\r\n{\"Method\":true}"
	actual := rpc.EncodeMessage(EncodingExample{Method: true})
	if expected != actual {
		t.Fatalf("Expected %s, got %s", expected, actual)
	}
}

func TestDecode(t *testing.T) {
	expected := "Content-Length: 15\r\n\r\n{\"Method\":\"hi\"}"

	method, content, err := rpc.DecodeMessage([]byte(expected))
	contentLength := len(content)
	if err != nil {
		t.Fatalf("Error decoding message: %s", err)
	}

	if contentLength != 15 {
		t.Fatalf("Expected content length of 15, got %d", contentLength)
	}

	if method != "hi" {
		t.Fatalf("Expected method 'hi', got %s", method)
	}
}
