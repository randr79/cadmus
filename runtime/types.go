package runtime

import "context"

type CommandEntry struct {
	Label  string
	Fields map[string]FieldInfo // Key: FieldName or FieldType (if anon)
}

type EntryPoint func() Command

type Command interface {
	Command() string
	GetMetadata() CommandEntry                                                    // Returns data from the manifest
	Prepare(arguments []string) (runable Runnable, remainder []string, err error) // Performs the type-safe hydration
}

type FieldInfo struct {
	//	Type      string   // Type string: int, string, time.Duration, []string, *[]string, *string
	Default   string   // Default value as a string
	Env       string   // Environment variable name
	Arguments []string // the command line argument keys (e.g. -a --append or */# for any positional or a number for fixed positions)
}

// Runnable is the interface implemented by your business logic structs.
type Runnable interface {
	// Run receives the context (for Core/Manifest) and
	// the slice of arguments NOT consumed by the command match or flags.
	Run(ctx context.Context, args []string) error
}
