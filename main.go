package main

import (
	"github.com/randr79/cadmus/cmd/commands"
)

//go:generate bash ./cmd/bootstrap/bootstrap.sh

//go:generate ./cadmus extract -t cadmus ./cmd/applets
//go:generate ./cadmus build -o ./cmd/applets

func main() {
	commands.Dispatch()
}
