package build

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"go/format"
	"maps"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"

	"github.com/randr79/cadmus/types"
)

//go:embed templates/adapter.tmpl
var adapterTemplateStr string

//go:embed templates/router.tmpl
var routerTemplateStr string

type Builder struct {
	adapterTemplate *template.Template
	routerTemplate  *template.Template
	manifest        *types.Manifest
}

type adapterFieldData struct {
	Name            string
	Type            string
	CleanType       string
	IsSlice         bool
	IsMap           bool
	Match           string
	Validate        string
	TypeConstructor string
	Default         string
	Arguments       []string
}

type adapterTemplateData struct {
	AdapterName   string
	Command       string
	AppletPackage string
	ThisPackage   string
	Struct        string
	ImportLines   []string
	Fields        []adapterFieldData
}

func NewBuilder(manifest *types.Manifest) (*Builder, error) {
	if at, err := template.New("adapters").Funcs(template.FuncMap{
		"stringsPkgName": func(p string) string {
			parts := strings.Split(p, "/")
			return parts[len(parts)-1]
		},
	}).Parse(adapterTemplateStr); err != nil {
		return nil, fmt.Errorf("invalid adapter template %w", err)
	} else if rt, err := template.New("router").Parse(routerTemplateStr); err != nil {
		return nil, fmt.Errorf("invalid router template %w", err)
	} else {

		return &Builder{
			manifest:        manifest,
			adapterTemplate: at,
			routerTemplate:  rt,
		}, nil
	}
}

func randName() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 32)
	for i := range b {
		b[i] = letters[rand.Intn(26)]
	}
	return string(b)
}

func GetTypeConstructor(typeName string) string {
	switch {
	case typeName == "string":
		return fmt.Sprintf("NewStringArgument[%s]", typeName)
	case strings.HasPrefix(typeName, "int"):
		return fmt.Sprintf("NewIntArgument[%s]", typeName)
	case strings.HasPrefix(typeName, "uint"):
		return fmt.Sprintf("NewUintArgument[%s]", typeName)
	case strings.HasPrefix(typeName, "float"):
		return fmt.Sprintf("NewFloatArgument[%s]", typeName)
	case strings.HasPrefix(typeName, "complex"):
		return fmt.Sprintf("NewComplexArgument[%s]", typeName)
	case typeName == "bool":
		return fmt.Sprintf("NewBoolArgument[%s]", typeName)
	default:
		return fmt.Sprintf("NewTextArgument[%s]", typeName)
	}
}

func fieldTemplateData(name string, info types.FieldInfo) adapterFieldData {

	afd := adapterFieldData{
		Name:      name,
		Type:      info.Type,
		Match:     info.Match,
		Default:   info.Default,
		Arguments: info.Arguments,
	}

	if strings.HasPrefix(info.Type, "[]") {
		afd.IsSlice = true
		afd.CleanType = strings.TrimPrefix(info.Type, "[]")
		afd.TypeConstructor = GetTypeConstructor(afd.CleanType)
	} else if strings.HasPrefix(info.Type, "map[string]") {
		afd.IsMap = true
		afd.CleanType = strings.TrimPrefix(info.Type, "map[string]")
		afd.TypeConstructor = GetTypeConstructor(afd.CleanType)
	} else {
		afd.CleanType = info.Type
		afd.TypeConstructor = GetTypeConstructor(info.Type)

	}

	if info.Validate != "" && !strings.HasPrefix(info.Validate, ".") {
		afd.Validate = "." + info.Validate
	} else {
		afd.Validate = info.Validate
	}

	return afd
}

func needsRegexp(fields map[string]types.FieldInfo) bool {
	for _, info := range fields {
		if info.Match != "" {
			return true
		}
	}
	return false
}

func (b Builder) nameCommands() map[string]types.CommandEntry {
	result := make(map[string]types.CommandEntry, len(b.manifest.Commands))
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	for i, cmd := range b.manifest.Commands {
		name := strings.Trim(reg.ReplaceAllString(strings.ToUpper(cmd.Command), "_"), "_")
		if name == "" {
			result[fmt.Sprintf("%s", randName())] = cmd
		} else if _, exists := result[name]; exists {
			result[fmt.Sprintf("%s%d", name, i)] = cmd
		} else {
			result[fmt.Sprintf("%s", name)] = cmd
		}
	}
	return result
}

func (b Builder) CreateAdapter(adapterName string, cmd types.CommandEntry, targetDir string) error {
	// 1. Maak het specifieke adapter bestand aan

	fileName := filepath.Join(targetDir, fmt.Sprintf("%s.gen.go", strings.ToLower(adapterName)))
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("could not create adapter file: %w", err)
	}
	defer file.Close()

	importLines := make(map[string]bool)
	importLines[`"fmt"`] = true
	importLines[`"github.com/randr79/cadmus/arguments"`] = true
	importLines[`"github.com/randr79/cadmus/types"`] = true
	importLines[fmt.Sprintf(`"%s"`, cmd.Package)] = true

	for alias, path := range cmd.Imports {
		if alias != "" {
			importLines[fmt.Sprintf(`"%s" "%s"`, alias, path)] = true
		} else {
			importLines[fmt.Sprintf(`"%s"`, path)] = true
		}
	}
	if !importLines[`"regexp"`] && needsRegexp(cmd.Fields) {
		importLines[`"regexp"`] = true
	}

	fields := make([]adapterFieldData, len(cmd.Fields))
	for i, fieldName := range slices.Sorted(maps.Keys(cmd.Fields)) {
		fields[i] = fieldTemplateData(fieldName, cmd.Fields[fieldName])
	}

	data := adapterTemplateData{
		AdapterName:   adapterName,
		Command:       cmd.Command,
		AppletPackage: cmd.Package,
		Struct:        cmd.Struct,
		ImportLines:   slices.Sorted(maps.Keys(importLines)),
		Fields:        fields,
		ThisPackage:   filepath.Base(targetDir),
	}

	buf := new(bytes.Buffer)
	if err := b.adapterTemplate.ExecuteTemplate(buf, "adapter", data); err != nil {
		return fmt.Errorf("failed to render adapter %s: %w", cmd.Command, err)
	} else if formatted, err := format.Source(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to format adapter %s: %w", cmd.Command, err)
	} else if _, err := file.Write(formatted); err != nil {
		return fmt.Errorf("failed to write adapter %s to file: %w", cmd.Command, err)
	}
	file.Sync()
	return nil

}

func (b Builder) CreateManifest(manifestFile string) error {

	manifest, err := os.Create(manifestFile)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(manifest)
	enc.SetIndent("", "    ")
	enc.SetEscapeHTML(false)
	return enc.Encode(b.manifest)
}

func (b Builder) CreateRouter(entrypoints map[string]string, targetDir string) error {
	if err := b.CreateManifest(fmt.Sprintf("%s/manifest.json", strings.TrimSuffix(targetDir, "/"))); err != nil {
		return err
	}
	fileName := filepath.Join(targetDir, "router.gen.go")
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("could not create router file: %w", err)
	}
	defer file.Close()

	templateData := struct {
		Entrypoints map[string]string
		AppName     string
		ThisPackage string
	}{
		Entrypoints: entrypoints,
		AppName:     b.manifest.Project,
		ThisPackage: filepath.Base(targetDir),
	}
	buf := new(bytes.Buffer)
	if err := b.routerTemplate.ExecuteTemplate(buf, "router", templateData); err != nil {
		return fmt.Errorf("failed to render router: %w", err)
	} else if formatted, err := format.Source(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to format router: %w", err)
	} else if _, err := file.Write(formatted); err != nil {
		return fmt.Errorf("failed to write router to file: %w", err)
	}
	file.Sync()
	return nil
}

func (b Builder) WriteCommandAdapters(targetDir string) (map[string]string, error) {
	entrypoints := make(map[string]string)

	// Haal de map met unieke namen en bijbehorende entries op
	processedCmds := b.nameCommands()

	for adapterName, cmd := range processedCmds {
		if err := b.CreateAdapter(adapterName, cmd, targetDir); err != nil {
			return nil, fmt.Errorf("failed to create adapter for command %s: %w", cmd.Command, err)
		}
		entrypoints[cmd.Command] = adapterName
	}

	return entrypoints, nil
}
