package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"proof/analysis"
	"proof/lsp"
	"proof/rpc"
	"strings"

	"github.com/f1monkey/spellchecker"
)

const word_list = "https://raw.githubusercontent.com/makifdb/spellcheck/main/words.txt"

func main() {
	args := os.Args
	var log_file_path string = args[1]

	logger := getLogger(log_file_path)
	logger.Println("Starting proof")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(rpc.Split)

	sc, err := spellchecker.New(
		spellchecker.DefaultAlphabet, // Allowed symbols
		spellchecker.WithMaxErrors(2),
		spellchecker.WithSplitter(bufio.ScanLines),
	)

	if err != nil {
		panic(err)
	}

	response, err := http.Get(word_list)

	if err != nil {
		logger.Fatal("Word list not found")
		return
	}

	defer response.Body.Close()
	reader := strings.NewReader(analysis.Abbreviations)

	sc.AddFrom(response.Body)
	sc.AddFrom(reader)

	state := analysis.NewState(sc)
	writer := os.Stdout

	for scanner.Scan() {
		bytes := scanner.Bytes()
		method, content, err := rpc.DecodeMessage(bytes)

		if err != nil {
			logger.Println(err)
			continue
		}

		handleMessage(logger, writer, &state, method, content)
	}
}

func handleMessage(logger *log.Logger, writer io.Writer, state *analysis.State, method string, content []byte) {
	logger.Printf("Received message with method '%s'", method)

	switch method {
	case "initialize":
		var request lsp.InitializeRequest

		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("Can't parse method 'initialize' | %s", err)
			return
		}

		logger.Printf("Connected to: %s %s",
			request.Params.ClientInfo.Name,
			request.Params.ClientInfo.Version)

		msg := lsp.NewInitializeResponse(request.ID)
		writeResponse(writer, msg)

		logger.Print("Sent initialize response")

	case "textDocument/didOpen":
		var request lsp.DidOpenTextDocumentNotification

		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("Can't parse method 'textDocumet/didOpen' | %s", err)
			return
		}

		logger.Printf("Opened: %s",
			request.Params.TextDocument.URI)

		diagnostics := state.OpenDocument(request.Params.TextDocument.URI, request.Params.TextDocument.Text, logger)
		logger.Printf("Diagnostics: %v", diagnostics)

		if len(diagnostics) > 0 {
			msg := lsp.NewPublishDiagnosticsNotification(request.Params.TextDocument.URI, diagnostics)
			writeResponse(writer, msg)

			logger.Print("didOpen Sent diagnostics")
		}

	case "textDocument/didChange":
		var request lsp.DidChangeTextDocumentNotification

		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("Can't parse method 'textDocumet/didChange' | %s", err)
			return
		}

		logger.Printf("Changed: %s",
			request.Params.TextDocument.URI)

		for _, change := range request.Params.ContentChanges {
			diagnostics := state.UpdateDocument(request.Params.TextDocument.URI, change.Text, logger)
			logger.Printf("Diagnostics: %v", diagnostics)

			if len(diagnostics) > 0 {
				msg := lsp.NewPublishDiagnosticsNotification(request.Params.TextDocument.URI, diagnostics)
				writeResponse(writer, msg)
				logger.Print("didChange Sent diagnostics")
			}
		}

		// case "textDocument/codeAction":
		// 	var request lsp.CodeActionRequest
		//
		// 	if err := json.Unmarshal(content, &request); err != nil {
		// 		logger.Printf("Can't parse method 'textDocument/codeAction' | %s", err)
		// 		return
		// 	}
		//
		// 	logger.Printf("Code action: %s",
		// 		request.Params.TextDocument.URI)
		//
		// 	response := state.CodeAction(request.ID, request.Params.TextDocument.URI)
		//
		// 	writeResponse(writer, response)
	}
}

func getLogger(filename string) *log.Logger {
	logfile, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)

	if err != nil {
		error_message, err := fmt.Printf("Bad file bro: %s", filename)

		if err != nil {
			panic(err)
		} else {
			panic(error_message)
		}
	}

	return log.New(logfile, "[proof]", log.Ldate|log.Ltime|log.Lshortfile)
}

func writeResponse(writer io.Writer, msg any) {
	reply := rpc.EncodeMessage(msg)
	writer.Write([]byte(reply))
}
