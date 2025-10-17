package generator

import (
	"testing"
)

func TestTemplateEngine_Render(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		template string
		data     interface{}
		want     string
		wantErr  bool
	}{
		{
			name:     "simple string interpolation",
			template: "Hello {{ .Name }}!",
			data:     map[string]string{"Name": "World"},
			want:     "Hello World!",
			wantErr:  false,
		},
		{
			name:     "toLower function",
			template: "{{ toLower .Text }}",
			data:     map[string]string{"Text": "HELLO WORLD"},
			want:     "hello world",
			wantErr:  false,
		},
		{
			name:     "toUpper function",
			template: "{{ toUpper .Text }}",
			data:     map[string]string{"Text": "hello world"},
			want:     "HELLO WORLD",
			wantErr:  false,
		},
		{
			name:     "toCamel function",
			template: "{{ toCamel .Name }}",
			data:     map[string]string{"Name": "user_name"},
			want:     "userName",
			wantErr:  false,
		},
		{
			name:     "toPascal function",
			template: "{{ toPascal .Name }}",
			data:     map[string]string{"Name": "user_name"},
			want:     "UserName",
			wantErr:  false,
		},
		{
			name:     "toSnake function",
			template: "{{ toSnake .Name }}",
			data:     map[string]string{"Name": "UserName"},
			want:     "user_name",
			wantErr:  false,
		},
		{
			name:     "toKebab function",
			template: "{{ toKebab .Name }}",
			data:     map[string]string{"Name": "UserName"},
			want:     "user-name",
			wantErr:  false,
		},
		{
			name:     "pluralize function",
			template: "{{ pluralize .Word }}",
			data:     map[string]string{"Word": "user"},
			want:     "users",
			wantErr:  false,
		},
		{
			name:     "pluralize y -> ies",
			template: "{{ pluralize .Word }}",
			data:     map[string]string{"Word": "category"},
			want:     "categories",
			wantErr:  false,
		},
		{
			name:     "singularize function",
			template: "{{ singularize .Word }}",
			data:     map[string]string{"Word": "users"},
			want:     "user",
			wantErr:  false,
		},
		{
			name:     "goType function - string",
			template: "{{ goType .Type }}",
			data:     map[string]string{"Type": "string"},
			want:     "string",
			wantErr:  false,
		},
		{
			name:     "goType function - int",
			template: "{{ goType .Type }}",
			data:     map[string]string{"Type": "int"},
			want:     "int64",
			wantErr:  false,
		},
		{
			name:     "goType function - datetime",
			template: "{{ goType .Type }}",
			data:     map[string]string{"Type": "datetime"},
			want:     "time.Time",
			wantErr:  false,
		},
		{
			name:     "jsonTag function - required",
			template: "{{ jsonTag .Name true }}",
			data:     map[string]string{"Name": "UserName"},
			want:     `json:"user_name"`,
			wantErr:  false,
		},
		{
			name:     "jsonTag function - optional",
			template: "{{ jsonTag .Name false }}",
			data:     map[string]string{"Name": "UserName"},
			want:     `json:"user_name,omitempty"`,
			wantErr:  false,
		},
		{
			name:     "dbTag function",
			template: "{{ dbTag .Name }}",
			data:     map[string]string{"Name": "UserName"},
			want:     `db:"user_name"`,
			wantErr:  false,
		},
		{
			name:     "quote function",
			template: "{{ quote .Text }}",
			data:     map[string]string{"Text": "hello"},
			want:     `"hello"`,
			wantErr:  false,
		},
		{
			name:     "backquote function",
			template: "{{ backquote .Text }}",
			data:     map[string]string{"Text": "hello"},
			want:     "`hello`",
			wantErr:  false,
		},
		{
			name:     "indent function",
			template: "{{ indent 4 .Text }}",
			data:     map[string]string{"Text": "line1\nline2"},
			want:     "    line1\n    line2",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Render(tt.template, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Render() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCaseConversions(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		camel  string
		pascal string
		snake  string
		kebab  string
	}{
		{
			name:   "snake_case input",
			input:  "user_name",
			camel:  "userName",
			pascal: "UserName",
			snake:  "user_name",
			kebab:  "user-name",
		},
		{
			name:   "camelCase input",
			input:  "userName",
			camel:  "userName",
			pascal: "UserName",
			snake:  "user_name",
			kebab:  "user-name",
		},
		{
			name:   "PascalCase input",
			input:  "UserName",
			camel:  "userName",
			pascal: "UserName",
			snake:  "user_name",
			kebab:  "user-name",
		},
		{
			name:   "kebab-case input",
			input:  "user-name",
			camel:  "userName",
			pascal: "UserName",
			snake:  "user_name",
			kebab:  "user-name",
		},
		{
			name:   "complex case",
			input:  "UserProfileData",
			camel:  "userProfileData",
			pascal: "UserProfileData",
			snake:  "user_profile_data",
			kebab:  "user-profile-data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toCamelCase(tt.input); got != tt.camel {
				t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, got, tt.camel)
			}
			if got := toPascalCase(tt.input); got != tt.pascal {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, got, tt.pascal)
			}
			if got := toSnakeCase(tt.input); got != tt.snake {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, got, tt.snake)
			}
			if got := toKebabCase(tt.input); got != tt.kebab {
				t.Errorf("toKebabCase(%q) = %q, want %q", tt.input, got, tt.kebab)
			}
		})
	}
}

func TestPluralizeSingularize(t *testing.T) {
	tests := []struct {
		singular string
		plural   string
	}{
		{"user", "users"},
		{"entity", "entities"},
		{"category", "categories"},
		{"box", "boxes"},
		{"class", "classes"},
		{"person", "persons"}, // simple implementation
	}

	for _, tt := range tests {
		t.Run(tt.singular, func(t *testing.T) {
			if got := pluralize(tt.singular); got != tt.plural {
				t.Errorf("pluralize(%q) = %q, want %q", tt.singular, got, tt.plural)
			}
			if got := singularize(tt.plural); got != tt.singular {
				t.Errorf("singularize(%q) = %q, want %q", tt.plural, got, tt.singular)
			}
		})
	}
}

func TestGoTypeConversion(t *testing.T) {
	tests := []struct {
		dataType string
		goType   string
	}{
		{"string", "string"},
		{"text", "string"},
		{"int", "int64"},
		{"integer", "int64"},
		{"bigint", "int64"},
		{"float", "float64"},
		{"decimal", "float64"},
		{"bool", "bool"},
		{"boolean", "bool"},
		{"date", "time.Time"},
		{"datetime", "time.Time"},
		{"timestamp", "time.Time"},
		{"json", "json.RawMessage"},
		{"uuid", "uuid.UUID"},
		{"unknown", "string"}, // default
	}

	for _, tt := range tests {
		t.Run(tt.dataType, func(t *testing.T) {
			if got := toGoType(tt.dataType); got != tt.goType {
				t.Errorf("toGoType(%q) = %q, want %q", tt.dataType, got, tt.goType)
			}
		})
	}
}
