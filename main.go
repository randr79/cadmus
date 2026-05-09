package main

import (
	_ "embed"

	"github.com/randr79/cadmus/cmd"
)

//go:generate go run ./bootstrap/generate.go
//go:generate go install .
func main() {
	cmd.Dispatch()
}
