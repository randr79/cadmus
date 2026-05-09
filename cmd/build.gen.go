// CODE GEGENEREERD DOOR MCL. NIET HANDMATIG AANPASSEN.
package cmd

import (
	"fmt"
	"github.com/randr79/cadmus/applets"
	"github.com/randr79/cadmus/runtime"
	"github.com/randr79/cadmus/runtime/types"
	"regexp"
)

type BUILD struct {
	Metadata runtime.CommandEntry
}

func (a *BUILD) GetMetadata() runtime.CommandEntry {
	return a.Metadata
}

func (a *BUILD) Command() string {
	return "build"
}

func (a *BUILD) Prepare(args []string) (runtime.Runnable, []string, error) {
	cmd := &applets.Build{}
	var errs []error

	rawParsed := runtime.ParseRawArgs(args)
	parsed, err := runtime.ParseArgs(a.Metadata.Fields, rawParsed)
	if err != nil {
		return nil, nil, err
	}

	{

		cmd.CommandsDir = func(v string, errs *[]error) string {
			return types.NewStringArgument[string](v).Validate(errs)
		}(parsed.Get("CommandsDir"), &errs)

	}

	{

		// Match tag validatie voor elke ruwe string
		for _, v := range parsed.List("Manifest") {
			matched, err := regexp.MatchString(`.*`, v)
			if err != nil {
				errs = append(errs, fmt.Errorf("ongeldige regex in match tag voor Manifest: %w", err))
				break
			}
			if !matched {
				errs = append(errs, fmt.Errorf("waarde `%s` voor Manifest voldoet niet aan het patroon `%s`", v, `.*`))
			}
		}

		cmd.Manifest = func(v string, errs *[]error) string {
			return types.NewStringArgument[string](v).MinLen(1).Validate(errs)
		}(parsed.Get("Manifest"), &errs)

	}

	if len(errs) > 0 {
		return nil, nil, fmt.Errorf("validation failed: %v", errs)
	}

	return cmd, args, nil
}

func New_BUILD() runtime.Command {
	return &BUILD{
		Metadata: runtime.CommandEntry{
			Label: "build",
			Fields: map[string]runtime.FieldInfo{

				"CommandsDir": {
					// Type: "string",
					Default:   ".",
					Arguments: []string{"-o", "--out"},
				},

				"Manifest": {
					// Type: "string",
					Default:   "",
					Arguments: []string{"-m", "--manifest"},
				},
			},
		},
	}
}
