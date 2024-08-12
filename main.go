package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
	"proof/analysis"
	"proof/lsp"
	"proof/rpc"
)

func main() {
	logger := getLogger("D:/code/proof/log.txt")
	logger.Println("Starting proof")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(rpc.Split)

	state := analysis.NewState()
	writer := os.Stdout

	for scanner.Scan() {
		bytes := scanner.Bytes()
		method, content, err := rpc.DecodeMessage(bytes)

		if err != nil {
			logger.Println(err)
			continue
		}

		handleMessage(logger, writer, state, method, content)
	}
}

func handleMessage(logger *log.Logger, writer io.Writer, state analysis.State, method string, content []byte) {
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

		state.OpenDocument(request.Params.TextDocument.URI, request.Params.TextDocument.Text)

	case "textDocument/didChange":
		var request lsp.DidChangeTextDocumentNotification

		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("Can't parse method 'textDocumet/didChange' | %s", err)
			return
		}

		logger.Printf("Changed: %s",
			request.Params.TextDocument.URI)

		for _, change := range request.Params.ContentChanges {
			state.UpdateDocument(request.Params.TextDocument.URI, change.Text)
		}

	case "textDocument/hover":
		var request lsp.HoverTextRequest

		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("Can't parse method 'textDocument/hover' | %s", err)
			return
		}

		logger.Printf("Hovered: %s",
			request.Params.TextDocument.URI)

		response := state.Hover(
			request.ID,
			request.Params.TextDocument.URI,
			request.Params.Position)

		writeResponse(writer, response)

		logger.Print("Sent initialize response")

	case "textDocument/codeAction":
		var request lsp.CodeActionRequest

		if err := json.Unmarshal(content, &request); err != nil {
			logger.Printf("Can't parse method 'textDocument/codeAction' | %s", err)
			return
		}

		logger.Printf("Code action: %s",
			request.Params.TextDocument.URI)

		response := state.CodeAction(request.ID, request.Params.TextDocument.URI)

		writeResponse(writer, response)
	}
}

func getLogger(filename string) *log.Logger {
	logfile, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)

	if err != nil {
		panic("Bad file bro")
	}

	return log.New(logfile, "[proof]", log.Ldate|log.Ltime|log.Lshortfile)
}

func writeResponse(writer io.Writer, msg any) {
	reply := rpc.EncodeMessage(msg)
	writer.Write([]byte(reply))
}
