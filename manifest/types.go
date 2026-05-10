package manifest

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
	Arguments     []string // the command line argument keys (e.g. -a --append or * or # for any positional or a number for fixed positions)
	ArgumentCount int      // the expected number of arguments
}
