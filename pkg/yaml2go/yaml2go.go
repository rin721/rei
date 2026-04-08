package yaml2go

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

type fieldDef struct {
	Name string
	Type string
	Tag  string
}

type structDef struct {
	Name   string
	Fields []fieldDef
}

type generator struct {
	defs []structDef
	seen map[string]struct{}
}

// Convert 将 YAML 内容转换为 Go 结构体定义源码。
func Convert(structName string, content []byte) (string, error) {
	if strings.TrimSpace(structName) == "" {
		return "", errors.New("struct name is required")
	}

	var raw any
	if err := yaml.Unmarshal(content, &raw); err != nil {
		return "", err
	}

	rootValue, ok := normalizeValue(raw).(map[string]any)
	if !ok {
		return "", errors.New("yaml root must be a mapping")
	}

	rootName := exportName(structName)
	gen := &generator{
		seen: make(map[string]struct{}),
	}
	gen.addStruct(rootName, rootValue)

	var builder strings.Builder
	builder.WriteString("package generated\n\n")
	for index := len(gen.defs) - 1; index >= 0; index-- {
		def := gen.defs[index]
		builder.WriteString("type ")
		builder.WriteString(def.Name)
		builder.WriteString(" struct {\n")
		for _, field := range def.Fields {
			builder.WriteString("\t")
			builder.WriteString(field.Name)
			builder.WriteString(" ")
			builder.WriteString(field.Type)
			builder.WriteString(" ")
			builder.WriteString(field.Tag)
			builder.WriteString("\n")
		}
		builder.WriteString("}\n\n")
	}

	return builder.String(), nil
}

func (g *generator) addStruct(name string, values map[string]any) string {
	if _, exists := g.seen[name]; exists {
		return name
	}
	g.seen[name] = struct{}{}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fields := make([]fieldDef, 0, len(keys))
	for _, key := range keys {
		fieldName := exportName(key)
		fields = append(fields, fieldDef{
			Name: fieldName,
			Type: g.inferType(name+fieldName, values[key]),
			Tag:  fmt.Sprintf("`yaml:\"%s\"`", key),
		})
	}

	g.defs = append(g.defs, structDef{Name: name, Fields: fields})
	return name
}

func (g *generator) inferType(typeName string, value any) string {
	switch typed := normalizeValue(value).(type) {
	case map[string]any:
		return g.addStruct(typeName, typed)
	case []any:
		return "[]" + g.sliceType(typeName+"Item", typed)
	case bool:
		return "bool"
	case int, int64, uint64:
		return "int"
	case float32, float64:
		return "float64"
	case nil:
		return "any"
	default:
		return "string"
	}
}

func (g *generator) sliceType(typeName string, values []any) string {
	for _, value := range values {
		normalized := normalizeValue(value)
		if normalized != nil {
			return g.inferType(typeName, normalized)
		}
	}
	return "any"
}

func normalizeValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, inner := range typed {
			result[key] = normalizeValue(inner)
		}
		return result
	case map[any]any:
		result := make(map[string]any, len(typed))
		for key, inner := range typed {
			result[fmt.Sprint(key)] = normalizeValue(inner)
		}
		return result
	case []any:
		result := make([]any, 0, len(typed))
		for _, inner := range typed {
			result = append(result, normalizeValue(inner))
		}
		return result
	default:
		return typed
	}
}

func exportName(value string) string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == '_' || r == '-' || unicode.IsSpace(r)
	})
	if len(parts) == 0 {
		return "Field"
	}

	var builder strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		builder.WriteRune(unicode.ToUpper(runes[0]))
		if len(runes) > 1 {
			builder.WriteString(string(runes[1:]))
		}
	}

	name := builder.String()
	if name == "" {
		return "Field"
	}
	if unicode.IsDigit([]rune(name)[0]) {
		return "Field" + name
	}
	return name
}
