package runtime

import (
	"strings"
)

type Router interface {
	GetRootCmd() string
	MaxRouteDepth() int
	GetRoutes() map[string]EntryPoint
	GetManifest() []byte
	Register(name string, entrypoint EntryPoint)
}

type router struct {
	appname  string
	manifest []byte
	commands map[string]EntryPoint
}

func NewRouter(embeddedManifest []byte, appname string) Router {
	return &router{
		appname:  appname,
		manifest: embeddedManifest,
		commands: make(map[string]EntryPoint),
	}
}

func (r *router) Register(name string, entrypoint EntryPoint) {
	r.commands[name] = entrypoint
}

func (r *router) GetRoutes() map[string]EntryPoint {
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
