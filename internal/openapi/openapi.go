package openapi

import (
	"google.golang.org/protobuf/compiler/protogen"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-krill-extensions/internal/krill"
)

type Openapi struct {
	Info              *Info
	PathItems         map[string]map[string]*Operation `yaml:"paths"`
	Components        *Components
	ServiceExtensions *krill.ServiceExtensions
}

type Info struct {
	Title   string
	Version string
	NoAuth  bool
}

type Components struct {
	Schemas   map[string]*Schema
	Responses map[string]*Response
}

func (o *Openapi) HasAuth() bool {
	return false
	// TODO: use extension to set this info
	//	return !o.Info.NoAuth
}

func FromProto(file *protogen.File, plugin *protogen.Plugin) (*Openapi, error) {
	var (
		enums      = parseEnums(plugin)
		extensions = krill.GetServiceExtensions(file.Proto.Service[0])
	)

	operations, err := parseOperations(file, plugin, enums)
	if err != nil {
		return nil, err
	}

	components, err := parseComponents(file, plugin, operations, enums)
	if err != nil {
		return nil, err
	}

	fileExtensions := krill.GetFileExtensions(file.Proto)

	return &Openapi{
		ServiceExtensions: extensions,
		PathItems:         operations,
		Components:        components,
		Info: &Info{
			Title:   fileExtensions.OpenapiTitle,
			Version: fileExtensions.OpenapiVersion,
		},
	}, nil
}

func parseComponents(api *protogen.File, plugin *protogen.Plugin, pathItems map[string]map[string]*Operation, enums map[string][]string) (*Components, error) {
	var (
		errorCodes = getResponseErrorCodesFromPaths(pathItems)
		schemas    = buildComponentsSchemas(getSchemaNamesFromPaths(pathItems), plugin, enums)
	)

	for name, schema := range responseErrorComponentsSchemas(errorCodes) {
		schemas[name] = schema
	}

	return &Components{
		Schemas: schemas,
	}, nil
}

// getSchemaNamesFromPaths retrieves the names of all Schemas that all Paths are using
func getSchemaNamesFromPaths(pathItems map[string]map[string]*Operation) []string {
	schemas := []string{}
	for _, path := range pathItems {
		for _, operation := range path {
			schemas = append(schemas, operation.Schemas()...)
		}
	}

	return schemas
}

func getResponseErrorCodesFromPaths(pathItems map[string]map[string]*Operation) map[string]bool {
	stringCodeToResponseCode := func(code string) string {
		switch code {
		case "400":
			return "RESPONSE_CODE_BAD_REQUEST"
		case "401":
			return "RESPONSE_CODE_UNAUTHORIZED"
		case "404":
			return "RESPONSE_CODE_NOT_FOUND"
		case "412":
			return "RESPONSE_CODE_PRECONDITION_FAILED"
		case "500":
			return "RESPONSE_CODE_INTERNAL_ERROR"
		}

		return "RESPONSE_CODE_OK"
	}

	codes := make(map[string]bool)
	for _, path := range pathItems {
		for _, operation := range path {
			for _, code := range operation.ResponseErrorCodes() {
				codes[stringCodeToResponseCode(code)] = true
			}
		}
	}

	return codes
}

func buildComponentsSchemas(schemaNames []string, plugin *protogen.Plugin, enums map[string][]string) map[string]*Schema {
	schemas := make(map[string]*Schema)

	for _, s := range schemaNames {
		name := trimPackagePath(s)
		if msg := findMessageByName(name, plugin); msg != nil {
			schema := messageToSchema(msg, enums, findProtogenMessageByName(name, plugin))
			schemas[name] = schema

			for _, p := range schema.Properties {
				refName := ""
				if p.Ref != "" {
					refName = p.RefName()
				}
				if p.Items != nil {
					refName = p.Items.RefName()
				}

				if refName != "" {
					for n, s := range buildComponentsSchemas([]string{refName}, plugin, enums) {
						schemas[n] = s
					}
				}
			}
		}
	}

	return schemas
}

func messageToSchema(message *descriptor.DescriptorProto, enums map[string][]string, msgSchema *protogen.Message) *Schema {
	properties := make(map[string]*Schema)

	for _, f := range message.Field {
		fieldExtensions := krill.GetFieldExtensions(f)
		if name, schema := fieldToSchema(f, enums, message, msgSchema, fieldExtensions); schema != nil {
			properties[name] = schema
		}
	}

	return NewSchema(&SchemaOptions{
		Type:       SchemaType_Object,
		Properties: properties,
	})
}

// responseErrorComponentsSchemas gives all error schemas that an API must have.
func responseErrorComponentsSchemas(errorCodes map[string]bool) map[string]*Schema {
	var (
		dup     = make(map[string]bool)
		schemas = make(map[string]*Schema)
	)

	if _, ok := errorCodes["RESPONSE_CODE_BAD_REQUEST"]; ok {
		schemas["FieldValidationError"] = NewSchema(&SchemaOptions{
			Type: SchemaType_Object,
			Properties: map[string]*Schema{
				"field": NewSchema(&SchemaOptions{
					Type: SchemaType_String,
				}),
				"message": NewSchema(&SchemaOptions{
					Type: SchemaType_String,
				}),
				"location": NewSchema(&SchemaOptions{
					Type: SchemaType_String,
				}),
			},
		})

		schemas["ValidationError"] = NewSchema(&SchemaOptions{
			Type: SchemaType_Object,
			Properties: map[string]*Schema{
				"errors": NewSchema(&SchemaOptions{
					Type: SchemaType_Array,
					Items: NewSchema(&SchemaOptions{
						Ref: refComponentsSchemas + "FieldValidationError",
					}),
				}),
				"message": NewSchema(&SchemaOptions{
					Type: SchemaType_String,
				}),
			},
		})
	}

	// Duplicates the errorCodes map so we can remove items without affecting
	// the original one
	for k := range errorCodes {
		dup[k] = true
	}

	// Remove the BAD_REQUEST error since it's the only one that uses a
	// different schema.
	delete(dup, "RESPONSE_CODE_BAD_REQUEST")

	if len(dup) > 0 {
		schemas["DefaultError"] = NewSchema(&SchemaOptions{
			Type: SchemaType_Object,
			Properties: map[string]*Schema{
				"errors": NewSchema(&SchemaOptions{
					Type: SchemaType_Array,
					Items: NewSchema(&SchemaOptions{
						Type: SchemaType_String,
					}),
				}),
				"message": NewSchema(&SchemaOptions{
					Type: SchemaType_String,
				}),
			},
		})
	}

	return schemas
}
