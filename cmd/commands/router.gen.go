// CODE GEGENEREERD DOOR MCL. NIET HANDMATIG AANPASSEN.
package commands

import (
	_ "embed"
	"github.com/randr79/cadmus/runtime"
)

//go:embed manifest.json
var embeddedManifest []byte
var Dispatch func() error

func init() {
	router := runtime.NewRouter(embeddedManifest, "cadmus")

	router.Register("build", New_BUILD)
	router.Register("extract", New_EXTRACT)

	Dispatch = runtime.NewDispatcher(router).Dispatch

}
