package main

import (
	"context"
	"fmt"
	"os"

	"github.com/randr79/cadmus/cmd/applets"
)

func main() {
	//bootstrapping the javlin tool (eat our own dogfood)
	//as we are executed from main these paths are relative to main!
	extract("./cmd/applets")
	create("./cmd/commands")
}

func extract(appletsDirs ...string) {
	if err := (&applets.Extract{
		Title:   "cadmus",
		Applets: appletsDirs,
	}).Run(context.Background(), nil); err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
}
func create(commandsDir string) {
	if err := (&applets.Build{
		CommandsDir: commandsDir,
	}).Run(context.Background(), nil); err != nil {
		os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
	}

}

/*
	if err := validate("../cmd/javelin/commands", "."); err != nil {
		t.Fatal(err)
	}

	func cleanup(dir string) error {
		d, err := os.Open(dir)
		if err != nil {
			return err
		}
		defer d.Close()
		names, err := d.Readdirnames(-1)
		if err != nil {
			return err
		}
		for _, name := range names {
			if !strings.HasSuffix(name, ".gen.go") && !strings.HasSuffix(name, ".json") {
				continue
			}
			if err := os.Remove(filepath.Join(dir, name)); err != nil {
				return err
			}
		}
		return nil
	}

func getDirEntries(dir string) ([]string, error) {
	if a, err := os.Open(dir); err != nil {
		return nil, fmt.Errorf("could not read dir: %s\n", err.Error())
	} else {
		defer a.Close()
		return a.Readdirnames(-1)
	}
}
func validate(commandsDir string, testDir string) error {
	var err error
	var names [2][]string
	if names[0], err = getDirEntries(commandsDir); err != nil {
		return fmt.Errorf("could not read commands dir: %s\n", err.Error())
	} else if names[1], err = getDirEntries(testDir); err != nil {
		return fmt.Errorf("could not read test dir: %s\n", err.Error())
	}

	all := make(map[string]bool)
	for _, nameset := range names {
		for _, name := range nameset {
			if name == "tmp.json" {
				continue
			}
			if strings.HasSuffix(name, ".gen.go") || strings.HasSuffix(name, ".json") {
				all[name] = true
			}
		}
	}
	var errs []error

	for name := range all {
		if !slices.Contains(names[0], name) {
			errs = append(errs, fmt.Errorf("missing %s in commands dir\n", name))
		} else if !slices.Contains(names[1], name) {
			errs = append(errs, fmt.Errorf("missing %s in test dir\n", name))
		} else {
			//compare the contents of the files using the os package, if they differ report an error
			cmd := exec.Command("diff", filepath.Join(commandsDir, name), filepath.Join(testDir, name))
			if err := cmd.Run(); err != nil {
				errs = append(errs, fmt.Errorf("files %s and %s differ\n", filepath.Join(commandsDir, name), filepath.Join(testDir, name)))
			}
		}
	}
	return errors.Join(errs...)

}
*/
