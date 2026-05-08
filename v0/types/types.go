package types

import (
	"context"
)

type Manifest struct {
	Project     string
	GeneratedAt string
	// Key is the full command string, e.g., "server start"
	Commands []CommandEntry
}

type CommandEntry struct {
	Command     string
	Label       string
	Description string
	Package     string               // Full import path: ://github.com
	Struct      string               // The struct name e.g. StartCmd
	Imports     map[string]string    // Key: local alias, Value: full import path
	Fields      map[string]FieldInfo // Key: FieldName or FieldType (if anon)
}

type FieldInfo struct {
	Description   string   // Comment
	Type          string   // Type string: int, string, time.Duration, []string, *[]string, *string
	Default       string   // Default value as a string
	Match         string   // Regex pattern
	Validate      string   // Validate method with args (e.g. between(1,10))
	Env           string   // Environment variable name
	Required      bool     // Whether this field is required (based on required tag)
	Arguments     []string // the command line argument keys (e.g. -a --append or */# for any positional or a number for fixed positions)
	ArgumentCount int      // the expected number of arguments
}

type EntryPoint func() Command

type Command interface {
	Command() string
	GetMetadata() CommandEntry                                                    // Returns data from the manifest
	Prepare(arguments []string) (runable Runnable, remainder []string, err error) // Performs the type-safe hydration
}

// Runnable is the interface implemented by your business logic structs.
type Runnable interface {
	// Run receives the context (for Core/Manifest) and
	// the slice of arguments NOT consumed by the command match or flags.
	Run(ctx context.Context, args []string) error
}
