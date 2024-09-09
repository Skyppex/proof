package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"proof/analysis"
	"proof/lsp"
	"proof/rpc"
	"strings"

	"github.com/f1monkey/spellchecker"
)

func main() {
	args := os.Args

	logger := getLogger(args)
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

	word_list := analysis.WordList
	reader := strings.NewReader(word_list)
	sc.AddFrom(reader)

	state := analysis.NewState(sc)
	writer := os.Stdout

	shuttingDown := false

	for scanner.Scan() {
		bytes := scanner.Bytes()

		method, content, err := rpc.DecodeMessage(bytes)

		if err != nil {
			logger.Println(err)
			continue
		}

		shouldExit, shutdownReceived := handleMessage(logger, writer, &state, method, content)

		if shouldExit {
			break
		}

		if shutdownReceived {
			shuttingDown = true
		}
	}

	if !shuttingDown {
		logger.Print("Exiting without shutdown message")
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func handleMessage(
	logger *log.Logger,
	writer io.Writer,
	state *analysis.State,
	method string,
	content []byte) (bool, bool) {

	switch method {
	case "initialize":
		logger.Print("Initializing")
		var request lsp.InitializeRequest

		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("Can't parse method 'initialize' | %s", err)
			return false, false
		}

		logger.Printf("Connected to: %s %s",
			request.Params.ClientInfo.Name,
			request.Params.ClientInfo.Version)

		msg := lsp.NewInitializeResponse(request.ID)
		writeResponse(writer, msg)

		logger.Print("Sent initialize response")

	case "initialized":
		logger.Print("Initialized")

	case "shutdown":
		var request lsp.Shutdown

		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("Can't parse method 'shutdown' | %s", err)

			return false, false
		}

		logger.Print("Shutting down")

		msg := lsp.NewShutdownResponse(request.ID)
		writeResponse(writer, msg)
		return false, true

	case "exit":
		logger.Print("Exiting")

		return true, false

	case "workspace/didChangeConfiguration":
		var request lsp.DidChangeConfigurationRequest

		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("Can't parse method 'workspace/didChangeConfiguration' | %s", err)
			return false, false
		}

		logger.Printf("Configuration changed: %v",
			request.Params.Settings)

		state.UpdateSettings(request.Params.Settings, logger)

	case "workspace/executeCommand":
		var request lsp.ExecuteCommandRequest

		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("Can't parse method 'workspace/executeCommand' | %s", err)
			return false, false
		}

		logger.Printf("Execute command: %s",
			request.Params.Command)

		uri, diagnostics := state.ExecuteCommand(request.Params.Command, request.Params.Arguments, logger)

		if uri != "" {
			msg := lsp.NewPublishDiagnosticsNotification(uri, diagnostics)
			writeResponse(writer, msg)

			logger.Print("executeCommand sent diagnostics")
		}

	case "textDocument/didOpen":
		var request lsp.DidOpenTextDocumentNotification

		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("Can't parse method 'textDocument/didOpen' | %s", err)
			return false, false
		}

		logger.Printf("Opened: %s",
			request.Params.TextDocument.URI)

		diagnostics, diagnosticsDiffer := state.OpenDocument(request.Params.TextDocument, logger)

		if diagnosticsDiffer {
			msg := lsp.NewPublishDiagnosticsNotification(request.Params.TextDocument.URI, diagnostics)
			writeResponse(writer, msg)

			logger.Print("didOpen sent diagnostics")
		}

	case "textDocument/didChange":
		var request lsp.DidChangeTextDocumentNotification

		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("Can't parse method 'textDocument/didChange' | %s", err)
			return false, false
		}

		logger.Printf("Changed: %s",
			request.Params.TextDocument.URI)

		for _, change := range request.Params.ContentChanges {
			diagnostics, diagnosticsDiffer := state.UpdateDocument(request.Params.TextDocument, change.Text, logger)

			if diagnosticsDiffer {
				msg := lsp.NewPublishDiagnosticsNotification(request.Params.TextDocument.URI, diagnostics)
				writeResponse(writer, msg)
				logger.Print("didChange sent diagnostics")
			}
		}

	// case "textDocument/diagnostic":
	// 	var request lsp.DiagnosticRequest
	//
	// 	if err := json.Unmarshal(content, &request); err != nil {
	// 		logger.Printf("Can't parse method 'textDocument/diagnostic' | %s", err)
	// 		return false, false
	// 	}
	//
	// 	logger.Printf("Diagnostic: %s", request.Params.TextDocument.URI)
	//
	// 	diagnostics, diagnosticsDiffer := state.Diagnostic(request.Params.TextDocument.URI, logger)
	//
	// 	kind := lsp.Unchanged
	//
	// 	if diagnosticsDiffer {
	// 		kind = lsp.Full
	// 	}
	//
	// 	msg := lsp.NewDiagnosticResponse(request.ID, kind, diagnostics, request.Params.TextDocument.URI)
	// 	writeResponse(writer, msg)
	//
	// 	logger.Printf("diagnostic sent diagnostics: v%, v%", len(diagnostics), kind)

	case "textDocument/codeAction":
		logger.Print("Received Code Action Request")
		var request lsp.CodeActionRequest

		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("Can't parse method 'textDocument/codeAction' | %s", err)
			return false, false
		}

		logger.Printf("Code action: %s",
			request.Params.TextDocument.URI)

		response := state.CodeAction(request, request.Params.TextDocument.URI, logger)

		logger.Printf("Code action response: %v", response)

		writeResponse(writer, response)

	default:
		logger.Printf("Unhandled method: %s", method)

	}

	return false, false
}

func getLogger(args []string) *log.Logger {
	if len(args) <= 1 {
		return log.New(io.Discard, "[proof]", log.Ldate|log.Ltime|log.Lshortfile)
	}

	filename := args[1]
	log_file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)

	if err != nil {
		error_message, err := fmt.Printf("Bad file bro: %s", filename)

		if err != nil {
			panic(err)
		} else {
			panic(error_message)
		}
	}

	return log.New(log_file, "[proof]", log.Ldate|log.Ltime|log.Lshortfile)
}

func writeResponse(writer io.Writer, msg any) {
	reply := rpc.EncodeMessage(msg)
	writer.Write([]byte(reply))
}
