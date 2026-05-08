package parse

import (
	"fmt"
	"os"
	"strings"

	"github.com/randr79/cadmus/types"
)

type ParsedArgs map[string][]string

// De methode zoekt simpelweg de lijst met keys af
func (p ParsedArgs) Get(key string) string {
	if val, ok := p[key]; ok && len(val) == 1 {
		return val[0]
	}
	return ""
}
func (p ParsedArgs) List(key string) []string {
	if val, ok := p[key]; ok && len(val) > 0 {
		return val
	}
	return make([]string, 0)
}
func (p ParsedArgs) Map(prefix string) map[string][]string {
	res := make(map[string][]string)
	for k, v := range p {
		if bf, af, fn := strings.Cut(k, "."); fn && bf == prefix {
			res[af] = v
		} else if bf, af, fn := strings.Cut(k, "-"); fn && bf == prefix {
			res[af] = v
		}
	}
	return res
}

func explodeMultiflagArgs(args []string) []string {
	result := make([]string, 0, len(args))
	for _, arg := range args {
		runes := []rune(arg)
		la := len(runes)
		if la > 2 && runes[0] == '-' && runes[1] != '-' {
			for _, r := range runes[1:] {
				result = append(result, "-"+string(r))
			}
		} else {
			result = append(result, arg)
		}
	}
	return result
}

func ParseRawArgs(args []string) map[string][]string {
	args = explodeMultiflagArgs(args)
	result := make(map[string][]string)

	positionalCount := 0
	skipAllFlags := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// 1. Schakel vlaggen uit na '--'
		if arg == "--" && !skipAllFlags {
			skipAllFlags = true
			continue
		}

		// 2. Is het een positioneel argument?
		if len(arg) == 0 || arg[0] != '-' || skipAllFlags {
			positionalCount++
			key := fmt.Sprintf("%d", positionalCount)
			result[key] = []string{arg}
			continue
		}

		// 3. Ondersteuning voor '=' (bijv. --file=data.txt)
		if strings.Contains(arg, "=") {
			vlag, waarde, _ := strings.Cut(arg, "=")
			result[vlag] = []string{waarde}
			continue
		}

		// 4. Het is een vlag met een spatie (bijv. --file data.txt)
		if i+1 < len(args) && (len(args[i+1]) == 0 || args[i+1][0] != '-') {
			result[arg] = []string{args[i+1]}
			i++ // Sla de waarde over
			continue
		}

		// 5. Het is een losse boolean vlag (bijv. --verbose)
		result[arg] = []string{"true"}
	}

	return result
}

func ParseArgs(fields map[string]types.FieldInfo, rawArgs map[string][]string) (ParsedArgs, error) {
	result := make(map[string][]string)

	// We houden bij welke positionele argumenten we al hebben verbruikt
	usedPositionals := make(map[string]bool)

	// Stap A: Loop over alle gedefinieerde velden
	for fieldName, info := range fields {
		var values []string

		// 1. Zoek naar vlaggen in de rawArgs (bijv. "-a" of "--append")
		for _, argKey := range info.Arguments {
			if val, ok := rawArgs[argKey]; ok {
				values = append(values, val...)
			}
		}

		// 2. Zoek naar positionele argumenten (bijv. "1", "2", "#", "*")
		for _, argKey := range info.Arguments {
			if val, ok := rawArgs[argKey]; ok { // Specifieke positie (bijv. "1", "2")
				values = append(values, val...)
				usedPositionals[argKey] = true
			}

			if argKey == "#" { // Willekeurige vrije positie ("#")
				// Zoek het eerste positionele nummer dat nog niet gebruikt is
				for i := 1; ; i++ {
					posKey := fmt.Sprintf("%d", i)
					if _, exists := rawArgs[posKey]; !exists {
						break // Geen positionele args meer over
					}
					if !usedPositionals[posKey] {
						values = append(values, rawArgs[posKey]...)
						usedPositionals[posKey] = true
						break // We pakken er bij '#' slechts één
					}
				}
			}

			if argKey == "*" { // Alles-opvanger ("*")
				for i := 1; ; i++ {
					posKey := fmt.Sprintf("%d", i)
					if _, exists := rawArgs[posKey]; !exists {
						break
					}
					if !usedPositionals[posKey] {
						values = append(values, rawArgs[posKey]...)
						usedPositionals[posKey] = true
					}
				}
			}
		}

		// Stap B: Fallbacks toepassen als er niets is gevonden
		if len(values) == 0 {
			if val, ok := os.LookupEnv(info.Env); ok {
				values = []string{val}
			} else if info.Default != "" {
				values = []string{info.Default}
			}
		}

		// Sla het resultaat op voor dit veld (als er waardes zijn)
		if len(values) > 0 {
			result[fieldName] = values
		}
	}

	return result, nil
}
