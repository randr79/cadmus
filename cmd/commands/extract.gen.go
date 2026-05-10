// CODE GEGENEREERD DOOR MCL. NIET HANDMATIG AANPASSEN.
package commands

import (
	"fmt"
	"github.com/randr79/cadmus/cmd/applets"
	"github.com/randr79/cadmus/runtime"
	"github.com/randr79/cadmus/runtime/types"
)

type EXTRACT struct {
	Metadata runtime.CommandEntry
}

func (a *EXTRACT) GetMetadata() runtime.CommandEntry {
	return a.Metadata
}

func (a *EXTRACT) Command() string {
	return "extract"
}

func (a *EXTRACT) Prepare(args []string) (runtime.Runnable, []string, error) {
	cmd := &applets.Extract{}
	var errs []error

	rawParsed := runtime.ParseRawArgs(args)
	parsed, err := runtime.ParseArgs(a.Metadata.Fields, rawParsed)
	if err != nil {
		return nil, nil, err
	}

	{

		cmd.Applets = types.Apply(parsed.List("Applets"), func(v string, errs *[]error) string {
			return types.NewStringArgument[string](v).Validate(errs)
		}, &errs)

	}

	{

		cmd.Manifest = func(v string, errs *[]error) string {
			return types.NewStringArgument[string](v).Validate(errs)
		}(parsed.Get("Manifest"), &errs)

	}

	{

		cmd.Title = func(v string, errs *[]error) string {
			return types.NewStringArgument[string](v).Validate(errs)
		}(parsed.Get("Title"), &errs)

	}

	if len(errs) > 0 {
		return nil, nil, fmt.Errorf("validation failed: %v", errs)
	}

	return cmd, args, nil
}

func New_EXTRACT() runtime.Command {
	return &EXTRACT{
		Metadata: runtime.CommandEntry{
			Label: "extract",
			Fields: map[string]runtime.FieldInfo{

				"Applets": {
					// Type: "[]string",
					Default:   "./applets/",
					Arguments: []string{"#"},
				},

				"Manifest": {
					// Type: "string",
					Default:   "",
					Arguments: []string{"-m", "--manifest"},
				},

				"Title": {
					// Type: "string",
					Default:   "",
					Arguments: []string{"-t", "--title"},
				},
			},
		},
	}
}
