package applets

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path"
	"time"

	"github.com/randr79/cadmus/javelin/extract"
	"github.com/randr79/cadmus/types"
)

// @extract
// Scans Go files for @command markers and produces an json at `OutputFile`
type Extract struct {
	Manifest string   `arg:"-m, --manifest" default:""`                         //where to temporarily store the manifest with the commands
	Title    string   `arg:"-t, --title" default:""`                            //the title of the generated router
	Applets  []string `arg:"#" default:"./applets/" required:"true" count:"-1"` //where to find the structs (applets)
}

func (c *Extract) Run(ctx context.Context, args []string) error {
	if c.Manifest == "" {
		c.Manifest = path.Join(os.TempDir(), "manifest.json")
	}
	if c.Title == "" {
		// no title, set the title to the name of the parent directory
		cwd, _ := os.Getwd()
		parentDir := path.Dir(cwd)
		c.Title = path.Base(parentDir)
	}

	var enc *json.Encoder
	if fh, err := os.OpenFile(c.Manifest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm); err != nil {
		return err
	} else {
		defer fh.Close()
		enc = json.NewEncoder(fh)
		enc.SetIndent("", "    ")
		enc.SetEscapeHTML(false)
	}

	commands := make([]types.CommandEntry, 0)
	errs := make([]error, 0, len(c.Applets))

	for _, dir := range c.Applets {
		if cmds, err := extract.Extract(dir); err != nil {
			errs = append(errs, err)
		} else {
			commands = append(commands, cmds...)

		}
	}
	manifest := types.Manifest{
		Project:     c.Title,
		GeneratedAt: time.Now().Format(time.DateTime),
		Commands:    commands,
	}

	errs = append(errs, enc.Encode(manifest))
	return errors.Join(errs...)
}
