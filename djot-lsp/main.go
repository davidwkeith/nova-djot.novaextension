package main

import (
	"github.com/dwk/djot-lsp/server"
	glspServer "github.com/tliron/glsp/server"
)

const lsName = "djot-lsp"

func main() {
	handler := server.NewHandler()
	s := glspServer.NewServer(handler, lsName, false)
	s.RunStdio()
}
