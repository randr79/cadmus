package extract

import (
	"errors"
	"fmt"

	"github.com/randr79/cadmus/types"
	"golang.org/x/tools/go/packages"
)

func Extract(sourceDir string) ([]types.CommandEntry, error) {
	commands := make([]types.CommandEntry, 0)
	var errs []error
	if pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedName | packages.NeedTypesInfo,
	}, sourceDir); err != nil {
		errs = append(errs, fmt.Errorf("error loading packages from %s:%w", sourceDir, err))
	} else if len(pkgs) == 0 {
		errs = append(errs, fmt.Errorf("no packages found in %s", sourceDir))
	} else {
		for _, pkg := range pkgs {
			if len(pkg.Errors) > 0 {
				for _, e := range pkg.Errors {
					errs = append(errs, e)
				}
			}
			ce := &CommandExtractor{
				pkg: pkg,
			}
			for _, file := range pkg.Syntax {
				if cmds, err := ce.Extract(file); err != nil {
					errs = append(errs, err...)
				} else {
					commands = append(commands, cmds...)
				}
			}
		}

	}
	return commands, errors.Join(errs...)
}
