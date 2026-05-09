package applets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/randr79/cadmus/javelin/build"
	"github.com/randr79/cadmus/types"
)

// @build
// Reads a json export and generates the Go adapters and registry.
type Build struct {
	Manifest    string `arg:"-m, --manifest" default:"" match:".*" validate:"MinLen(1)"` //the manifest with the commands
	CommandsDir string `arg:"-o, --out" default:"."`                                     //where to write the commands to
}

func assertDir(target string) error {
	if _, err := os.Stat(target); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not read target directory: %w", err)
	} else if err := os.MkdirAll(target, 0750); err != nil {
		return fmt.Errorf("could not create target directory: %w", err)
	} else {
		return nil
	}

}

func readmanifest(inputfile string) (*types.Manifest, error) {
	var dec json.Decoder
	if h, err := os.Open(inputfile); err != nil {
		return nil, fmt.Errorf("could not read manifest: %w", err)
	} else {
		defer h.Close()
		dec = *json.NewDecoder(h)
		dec.DisallowUnknownFields()
	}
	var manifest types.Manifest
	if err := dec.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("could not parse JSON: %w", err)
	}
	return &manifest, nil

}

func (c *Build) Run(ctx context.Context, args []string) error {
	if c.Manifest == "" {
		c.Manifest = path.Join(os.TempDir(), "manifest.json")
	}

	if err := assertDir(c.CommandsDir); err != nil {
		return err
	} else if man, err := readmanifest(c.Manifest); err != nil {
		return err
	} else if builder, err := build.NewBuilder(man); err != nil {
		return err
	} else if entrypoints, err := builder.WriteCommandAdapters(c.CommandsDir); err != nil {
		return err
	} else if err := builder.CreateRouter(entrypoints, c.CommandsDir); err != nil {
		return fmt.Errorf("failed to create router: %w", err)
	} else {
		return nil
	}
}
