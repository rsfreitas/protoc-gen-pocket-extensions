package openapi

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

type SchemaOptions struct {
	Required    bool
	Expand      bool
	Minimum     int
	Maximum     int
	Type        SchemaType
	Format      string
	Ref         string
	Description string
	Example     string
	ExpandTo    string
	Properties  map[string]*Schema
	Enum        []string
	Items       *Schema
}

type Schema struct {
	Minimum     int                `yaml:"minimum,omitempty"`
	Maximum     int                `yaml:"maximum,omitempty"`
	Type        string             `yaml:"type,omitempty"`
	Format      string             `yaml:"format,omitempty"`
	Ref         string             `yaml:"$ref,omitempty"`
	Description string             `yaml:"description,omitempty"`
	Example     string             `yaml:"example,omitempty"`
	Items       *Schema            `yaml:"items,omitempty"`
	Enum        []string           `yaml:"enum,omitempty"`
	Required    []string           `yaml:"required,omitempty"`
	Properties  map[string]*Schema `yaml:"properties,omitempty"`

	schemaType SchemaType
	required   bool
	expand     bool
	expandTo   string
}

func (s *Schema) SchemaType() SchemaType {
	return s.schemaType
}

func (s *Schema) HasItems() bool {
	return s.schemaType == SchemaType_Array
}

func (s *Schema) IsRequired() bool {
	return s.required
}

func (s *Schema) RequiredProperties() []*Schema {
	var props []*Schema

	for _, p := range s.Properties {
		if p.IsRequired() {
			props = append(props, p)
		}
	}

	return props
}

func (s *Schema) HasRequiredProperty() bool {
	for _, p := range s.Properties {
		if p.IsRequired() {
			return true
		}
	}

	return false
}

func (s *Schema) String(prefixSpacing int) string {
	if s.Ref != "" {
		return s.asRef()
	}

	out, err := yaml.Marshal(s)
	if err != nil {
		panic(err.Error())
	}

	prefix := ""
	for i := 0; i < prefixSpacing; i += 1 {
		prefix += " "
	}

	newLines := []string{}
	for i, line := range strings.Split(string(out), "\n") {
		if len(line) == 0 {
			continue
		}

		// Skip the first line hoping that the user placed correctly the
		// cursor for it.
		if i > 0 {
			line = prefix + line
		}

		newLines = append(newLines, line)
	}

	return strings.Join(newLines, "\n")
}

func (s *Schema) asRef() string {
	return fmt.Sprintf(`$ref: "%s"`, s.Ref)
}

func (s *Schema) RefName() string {
	parts := strings.Split(s.Ref, "/")
	return parts[len(parts)-1]
}

func (s *Schema) Expand() bool {
	return s.expand
}

func (s *Schema) ExpandTo() string {
	return s.expandTo
}

func NewSchema(options *SchemaOptions) *Schema {
	s := &Schema{
		Minimum:     options.Minimum,
		Maximum:     options.Maximum,
		schemaType:  options.Type,
		Format:      options.Format,
		Ref:         options.Ref,
		Description: options.Description,
		Items:       options.Items,
		Properties:  options.Properties,
		Enum:        options.Enum,
		Example:     options.Example,
		required:    options.Required,
		expand:      options.Expand,
		expandTo:    options.ExpandTo,
	}

	var required []string
	for name, p := range s.Properties {
		if p.IsRequired() {
			required = append(required, name)
		}
	}
	s.Required = required
	sort.Strings(s.Required)

	if s.schemaType != SchemaType_Unspecified {
		s.Type = s.schemaType.String()
	}

	return s
}
