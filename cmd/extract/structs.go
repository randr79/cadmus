package extract

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/randr79/cadmus/manifest"
	"golang.org/x/tools/go/packages"
)

type CommandExtractor struct {
	pkg *packages.Package
}

func (ce CommandExtractor) parseRequiredTag(tag reflect.StructTag) bool {
	if req, ok := tag.Lookup("required"); !ok {
		return false
	} else if req == "" {
		return true //empty = true
	} else {
		v, _ := strconv.ParseBool(req)
		return v // invalid == false
	}
}

/*
	func (ce CommandExtractor) parseArgTag(tag reflect.StructTag) (short string, long string, positional bool) {
		for _, a := range strings.Split(tag.Get("arg"), " ") {
			switch {
			case a == "#":
				positional = true
			case strings.HasPrefix(a, "--"):
				long = strings.TrimPrefix(a, "--")
			case strings.HasPrefix(a, "-"):
				short = strings.TrimPrefix(a, "-")
			default:
				//invalid directive skip entire tag
				return "", "", false
			}
		}
		return
	}
*/

func (ce CommandExtractor) parseDocString(doc string) (cmd string, label string, comment string, ok bool) {
	if sdoc, ok := strings.CutPrefix(strings.TrimSpace(doc), "@"); !ok {
		// struct comment needs to start with "@"
		return "", "", "", false
	} else {
		l1, descr, _ := strings.Cut(sdoc, "\n")
		cmd, label, _ := strings.Cut(l1, ";")
		cmd = strings.TrimSpace(cmd)
		return strings.TrimSpace(cmd), strings.TrimSpace(label), strings.TrimSpace(descr), cmd != ""

	}
}

func (ce CommandExtractor) hasArguments(method *ast.FuncType, in []string, out []string) bool {
	if method.Params == nil || method.Results == nil ||
		len(method.Params.List) != len(in) || len(method.Results.List) != len(out) {
		return false
	}

	for i, ia := range method.Params.List {
		if ce.pkg.TypesInfo.TypeOf(ia.Type).String() != in[i] {
			return false
		}
	}
	for i, ia := range method.Results.List {
		if ce.pkg.TypesInfo.TypeOf(ia.Type).String() != out[i] {
			return false
		}
	}

	return true
}

func (ce CommandExtractor) getReceiver(method *ast.FuncDecl) (name string, isPointer bool) {
	if method.Recv == nil || len(method.Recv.List) != 1 {
		return "", false
	} else {
		switch t := method.Recv.List[0].Type.(type) {
		case *ast.Ident:
			// Receiver is a struct (e.g. "StatusCommand")
			return t.Name, false
		case *ast.StarExpr:
			// Receiver is a pointer to a struct (e.g. "*StatusCommand")
			if id, ok := t.X.(*ast.Ident); ok {
				return id.Name, true
			} else {
				return "", true
			}
		default:
			return "", false
		}
	}

}

func (ce CommandExtractor) isStructRunMethod(method *ast.FuncDecl) (structName string, ok bool) {
	// Check for: func (r *Receiver) Run(ctx context.Context, args []string) error
	if receiver, _ := ce.getReceiver(method); receiver == "" {
		return "", false
	} else if method.Name.Name != "Run" || !ce.hasArguments(method.Type, []string{"context.Context", "[]string"}, []string{"error"}) {
		return "", false
	} else {
		return receiver, true
	}
}

func (ce CommandExtractor) Extract(file *ast.File) ([]manifest.CommandEntry, []error) {
	candidates := make(map[string]manifest.CommandEntry)
	methods := make(map[string]bool)
	var errs []error
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok != token.TYPE || x.Doc == nil {
				//no comment on struct
				return true
			}
			for _, spec := range x.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); !ok {
					continue
				} else if st, ok := typeSpec.Type.(*ast.StructType); !ok {
					continue
				} else if cmd, label, desc, ok := ce.parseDocString(x.Doc.Text()); !ok {
					// not a command
				} else if fields, imports, err := ce.createCommandFields(st); err != nil {
					errs = append(errs, err)
				} else {
					candidates[cmd] = manifest.CommandEntry{
						Command:     cmd,
						Label:       label,
						Description: desc,
						Package:     ce.pkg.String(),
						Struct:      typeSpec.Name.Name,
						Imports:     imports,
						Fields:      fields,
					}
				}
			}
		case *ast.FuncDecl:
			if s, ok := ce.isStructRunMethod(x); ok {
				methods[s] = true
			}
		}
		return true
	})

	candidateNames := slices.Collect(maps.Keys(candidates))
	for _, cmd := range candidateNames {
		candidate := candidates[cmd]
		if !methods[candidate.Struct] {
			delete(candidates, cmd)
		}
	}
	return slices.Collect(maps.Values(candidates)), errs

}

func (ce CommandExtractor) getFieldImports(field ast.Expr) map[string]string {
	foundImports := make(map[string]string)
	ast.Inspect(field, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if ident, ok := sel.X.(*ast.Ident); ok {
			obj := ce.pkg.TypesInfo.Uses[ident]
			if pkgName, ok := obj.(*types.PkgName); ok {
				alias := pkgName.Name()
				path := pkgName.Imported().Path()
				if path != ce.pkg.PkgPath {
					foundImports[alias] = path
				}
			}
		}
		return false
	})
	return foundImports
}

func (ce CommandExtractor) getArgumentCount(tag string, fieldType string) (int, error) {
	typeInfo := strings.TrimLeft(fieldType, "*")
	//typeinfo is the textual presentation of the type e.g. []string, string, int etc. pointers are stripped (value is handled externally)
	if c := strings.TrimSpace(tag); c != "" {
		i, e := strconv.ParseInt(c, 10, 0)
		return int(i), e
	}
	if !strings.HasPrefix(typeInfo, "map[") {
		//noop
	} else if after, ok := strings.CutPrefix(typeInfo, "map[string]"); !ok {
		return 0, fmt.Errorf("only maps with string keys are supported, not `%s`", typeInfo)
	} else {
		/*
			map[] works with prefixes
			e.g. -map as an argument key
			--map.key1 foo --map.key2 bar
			should become
			{"key1":[]string{"foo"},"key2":[]string{"bar"}}
			so the keys would be unique and [] still needs to be checked on the map value
		*/
		typeInfo = strings.TrimLeft(after, "*") //strip pointer again
	}

	switch {
	case typeInfo == "bool":
		return 0, nil
	case strings.HasPrefix(typeInfo, "[]"):
		return -1, nil
	case strings.HasPrefix(typeInfo, "["):
		n := strings.SplitN(typeInfo[1:], "]", 2)[0]
		return strconv.Atoi(n)
	default:
		return 1, nil
	}

}

func (ce CommandExtractor) createCommandFields(st *ast.StructType) (fields map[string]manifest.FieldInfo, imports map[string]string, err error) {
	fields = make(map[string]manifest.FieldInfo)
	imports = make(map[string]string)

	for _, f := range st.Fields.List {
		var tag reflect.StructTag

		maps.Copy(imports, ce.getFieldImports(f.Type))

		//get the tag
		if f.Tag != nil && f.Tag.Value != "" {
			tag = reflect.StructTag(strings.Trim(f.Tag.Value, "`"))
		}

		var buf bytes.Buffer
		printer.Fprint(&buf, ce.pkg.Fset, f.Type)
		typestring := buf.String()
		fi := manifest.FieldInfo{
			Description: f.Comment.Text(),
			Type:        typestring,
			Validate:    tag.Get("validate"),
			Match:       tag.Get("match"),
			Default:     tag.Get("default"),
			Env:         tag.Get("env"),
			Required:    ce.parseRequiredTag(tag),
			Arguments: strings.FieldsFunc(tag.Get("arg"), func(r rune) bool {
				return r == ',' || unicode.IsSpace(r)
			})}

		if n, e := ce.getArgumentCount(tag.Get("count"), typestring); e != nil {
			return nil, nil, e
		} else {
			fi.ArgumentCount = n
		}

		if len(f.Names) > 0 {
			for _, n := range f.Names {
				if ast.IsExported(n.Name) {
					fields[n.Name] = fi
				}
			}
		} else {
			// Embedded field
			curr := f.Type
			for {
				if p, ok := curr.(*ast.StarExpr); ok {
					curr = p.X
					continue
				}
				if s, ok := curr.(*ast.SelectorExpr); ok {
					curr = s.Sel
					continue
				}
				break
			}
			if id, ok := curr.(*ast.Ident); ok {
				if ast.IsExported(id.Name) {
					fields[id.Name] = fi
				}
			}
		}
	}
	return
}
