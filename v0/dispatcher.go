package cadmus

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/randr79/cadmus/v0/errors"
	"github.com/randr79/cadmus/v0/types"
)

type Dispatcher interface {
	Dispatch() error
}

func NewDispatcher(router Router) Dispatcher {
	return &dispatcher{
		router: router,
	}
}

type dispatcher struct {
	router Router
}

func (d *dispatcher) getArgs() []string {
	basename := filepath.Base(os.Args[0])
	if strings.HasPrefix(basename, "__debug_") {
		return os.Args[1:]
	} else if d.router.GetRootCmd() == filepath.Base(os.Args[0]) {
		return os.Args[1:]
	} else {
		return os.Args
	}
}

func (d *dispatcher) getEntrypointArgs() (types.EntryPoint, []string, error) {
	args := d.getArgs()
	scanrange := d.router.MaxRouteDepth()
	if arglen := len(args); arglen < scanrange {
		scanrange = arglen
	}

	routes := d.router.GetRoutes()
	for i := scanrange; i > 0; i-- {
		if handler, ok := routes[strings.Join(args[:i], " ")]; ok {
			return handler, args[i:], nil
		}
	}
	return nil, nil, errors.UnknownCommand(fmt.Errorf("%s found no command for `%s`", d.router.GetRootCmd(), strings.Join(args, " ")))
}

func (d *dispatcher) newContext() context.Context {
	return ContextWithDispatcher(context.Background(), d)
}

func (d *dispatcher) Dispatch() error {
	if entrypoint, args, err := d.getEntrypointArgs(); err == nil {
		if runable, remainder, err := entrypoint().Prepare(args); err != nil {
			return err
		} else {
			return runable.Run(d.newContext(), remainder)
		}
	} else if unk, ok := err.(errors.UnknownCommand); !ok {
		return err
	} else {
		return d.Help(unk)
	}
}

func (d *dispatcher) Help(err error) error {
	if help, ok := d.router.GetRoutes()["help"]; !ok {
		return d.Index(err)
	} else if runable, remainder, err := help().Prepare([]string{}); err != nil {
		return err
	} else {
		return runable.Run(d.newContext(), remainder)
	}
}

func (d *dispatcher) Index(err error) error {
	r := d.router.GetRoutes()
	maxklen := len("Command")
	maxvlen := len("usage")
	labels := make(map[string]string)
	for k, v := range r {
		u := v().GetMetadata().Label
		labels[k] = u
		maxklen = max(maxklen, len(k))
		maxvlen = max(maxvlen, len(u))
	}

	kvfmt := fmt.Sprintf("%%-%ds %%%ds\n", maxklen, maxvlen)
	if err != nil {
		fmt.Printf(kvfmt, "Error:", err.Error())
	}
	fmt.Println("The following commands are available")
	for k, v := range labels {
		fmt.Printf(kvfmt, k, v)
	}
	return err
}
