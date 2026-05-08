package cadmus

import (
	"strings"

	"github.com/randr79/cadmus/types"
)

type Router interface {
	GetRootCmd() string
	MaxRouteDepth() int
	GetRoutes() map[string]types.EntryPoint
	GetManifest() []byte
	Register(name string, entrypoint types.EntryPoint)
}

type router struct {
	appname  string
	manifest []byte
	commands map[string]types.EntryPoint
}

func NewRouter(embeddedManifest []byte, appname string) Router {
	return &router{
		appname:  appname,
		manifest: embeddedManifest,
		commands: make(map[string]types.EntryPoint),
	}
}

func (r *router) Register(name string, entrypoint types.EntryPoint) {
	r.commands[name] = entrypoint
}

func (r *router) GetRoutes() map[string]types.EntryPoint {
	return r.commands
}

func (r *router) GetRootCmd() string {
	return r.appname
}

func (r *router) GetManifest() []byte {
	return r.manifest
}

func (r *router) MaxRouteDepth() int {
	maxDepth := 0
	for k := range r.commands {
		depth := len(strings.Fields(k))
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	return maxDepth
}
