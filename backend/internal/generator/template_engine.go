package generator

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TemplateEngine handles template rendering with custom functions
type TemplateEngine struct {
	funcMap template.FuncMap
}

// NewTemplateEngine creates a new template engine with helper functions
func NewTemplateEngine() *TemplateEngine {
	caser := cases.Title(language.English)
	funcMap := template.FuncMap{
		"toLower":     strings.ToLower,
		"toUpper":     strings.ToUpper,
		"toTitle":     caser.String,
		"toCamel":     toCamelCase,
		"toPascal":    toPascalCase,
		"toSnake":     toSnakeCase,
		"toKebab":     toKebabCase,
		"pluralize":   pluralize,
		"singularize": singularize,
		"join":        strings.Join,
		"replace":     strings.ReplaceAll,
		"contains":    strings.Contains,
		"hasPrefix":   strings.HasPrefix,
		"hasSuffix":   strings.HasSuffix,
		"trim":        strings.TrimSpace,
		"split":       strings.Split,
		"repeat":      strings.Repeat,
		"indent":      indent,
		"quote":       quote,
		"backquote":   backquote,
		"goType":      toGoType,
		"jsonTag":     toJSONTag,
		"dbTag":       toDBTag,
	}

	return &TemplateEngine{
		funcMap: funcMap,
	}
}

// Render renders a template with the given data
func (te *TemplateEngine) Render(templateStr string, data interface{}) (string, error) {
	tmpl, err := template.New("template").Funcs(te.funcMap).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// RenderFile renders a template file with the given data
func (te *TemplateEngine) RenderFile(templatePath string, data interface{}) (string, error) {
	tmpl, err := template.New("").Funcs(te.funcMap).ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template file: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, templatePath, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// Helper functions for template

// toCamelCase converts string to camelCase
func toCamelCase(s string) string {
	words := splitWords(s)
	if len(words) == 0 {
		return ""
	}
	caser := cases.Title(language.English)
	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		result += caser.String(strings.ToLower(words[i]))
	}
	return result
}

// toPascalCase converts string to PascalCase
func toPascalCase(s string) string {
	words := splitWords(s)
	result := ""
	caser := cases.Title(language.English)
	for _, word := range words {
		result += caser.String(strings.ToLower(word))
	}
	return result
}

// toSnakeCase converts string to snake_case
func toSnakeCase(s string) string {
	words := splitWords(s)
	return strings.Join(words, "_")
}

// ToSnakeCase converts string to snake_case (exported version)
func ToSnakeCase(s string) string {
	return toSnakeCase(s)
}

// toKebabCase converts string to kebab-case
func toKebabCase(s string) string {
	words := splitWords(s)
	return strings.Join(words, "-")
}

// splitWords splits a string into words
func splitWords(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}
	}

	var words []string
	var currentWord strings.Builder

	for i, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			if currentWord.Len() > 0 {
				words = append(words, strings.ToLower(currentWord.String()))
				currentWord.Reset()
			}
			continue
		}

		if i > 0 && isUpperCase(r) && !isUpperCase(rune(s[i-1])) {
			if currentWord.Len() > 0 {
				words = append(words, strings.ToLower(currentWord.String()))
				currentWord.Reset()
			}
		}

		currentWord.WriteRune(r)
	}

	if currentWord.Len() > 0 {
		words = append(words, strings.ToLower(currentWord.String()))
	}

	return words
}

// isUpperCase checks if a rune is uppercase
func isUpperCase(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// pluralize converts a word to plural form (simple implementation)
func pluralize(s string) string {
	if strings.HasSuffix(s, "y") {
		return s[:len(s)-1] + "ies"
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") || strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	return s + "s"
}

// singularize converts a word to singular form (simple implementation)
func singularize(s string) string {
	if strings.HasSuffix(s, "ies") {
		return s[:len(s)-3] + "y"
	}
	if strings.HasSuffix(s, "es") {
		return s[:len(s)-2]
	}
	if strings.HasSuffix(s, "s") && !strings.HasSuffix(s, "ss") {
		return s[:len(s)-1]
	}
	return s
}

// indent adds indentation to each line
func indent(spaces int, s string) string {
	lines := strings.Split(s, "\n")
	indentation := strings.Repeat(" ", spaces)
	for i, line := range lines {
		if line != "" {
			lines[i] = indentation + line
		}
	}
	return strings.Join(lines, "\n")
}

// quote wraps string in double quotes
func quote(s string) string {
	return fmt.Sprintf("%q", s)
}

// backquote wraps string in backticks
func backquote(s string) string {
	return "`" + s + "`"
}

// toGoType converts a data type string to Go type
func toGoType(dataType string) string {
	typeMap := map[string]string{
		"string":    "string",
		"text":      "string",
		"int":       "int64",
		"integer":   "int64",
		"bigint":    "int64",
		"float":     "float64",
		"decimal":   "float64",
		"bool":      "bool",
		"boolean":   "bool",
		"date":      "time.Time",
		"datetime":  "time.Time",
		"timestamp": "time.Time",
		"json":      "json.RawMessage",
		"uuid":      "uuid.UUID",
	}

	if goType, ok := typeMap[strings.ToLower(dataType)]; ok {
		return goType
	}
	return "string"
}

// toJSONTag creates a JSON struct tag
func toJSONTag(fieldName string, required bool) string {
	jsonName := toSnakeCase(fieldName)
	if required {
		return fmt.Sprintf(`json:"%s"`, jsonName)
	}
	return fmt.Sprintf(`json:"%s,omitempty"`, jsonName)
}

// toDBTag creates a database struct tag
func toDBTag(fieldName string) string {
	return fmt.Sprintf(`db:"%s"`, toSnakeCase(fieldName))
}
