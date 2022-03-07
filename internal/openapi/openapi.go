package openapi

import (
	"google.golang.org/protobuf/compiler/protogen"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-pocket-extensions/internal/pocket"
	pocketpb "github.com/rsfreitas/protoc-gen-pocket-extensions/options/pocket"
)

type Openapi struct {
	Info              *Info
	Servers           []*Server
	PathItems         map[string]map[string]*Operation `yaml:"paths"`
	Components        *Components
	ServiceExtensions *pocket.ServiceExtensions
}

type Info struct {
	Title   string
	Version string
	NoAuth  bool
}

type Server struct {
	Url         string
	Description string
}

type Components struct {
	Schemas   map[string]*Schema
	Responses map[string]*Response
}

func (o *Openapi) HasAuth() bool {
	return o.ServiceExtensions.Service != nil && o.ServiceExtensions.Service.GetSecurityScheme() != nil
}

func (o *Openapi) SecurityScheme(tabSize int) string {
	return buildSecuritySchemeFromServiceExtensions(o.ServiceExtensions, tabSize)
}

func FromProto(file *protogen.File, plugin *protogen.Plugin) (*Openapi, error) {
	var (
		enums          = parseEnums(plugin)
		extensions     = pocket.GetServiceExtensions(file.Proto.Service[0])
		fileExtensions = pocket.GetFileExtensions(file.Proto)
	)

	// Initialize parser options that can be used throughout the parsing calls.
	parserOptions := &parserOptions{
		file:              file,
		plugin:            plugin,
		enums:             enums,
		serviceExtensions: extensions,
		service:           file.Proto.Service[0],
	}

	operations, err := parseOperations(parserOptions)
	if err != nil {
		return nil, err
	}

	components, err := parseComponents(parserOptions, operations)
	if err != nil {
		return nil, err
	}

	return &Openapi{
		ServiceExtensions: extensions,
		PathItems:         operations,
		Components:        components,
		Servers:           parseServersFromFileExtensions(fileExtensions),
		Info: &Info{
			Title:   fileExtensions.OpenapiTitle,
			Version: fileExtensions.OpenapiVersion,
		},
	}, nil
}

func parseComponents(options *parserOptions, pathItems map[string]map[string]*Operation) (*Components, error) {
	var (
		errorCodes = getResponseErrorCodesFromPaths(pathItems)
		schemas    = buildComponentsSchemas(getSchemaNamesFromPaths(pathItems), options)
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
	var (
		schemas = []string{}
		names   = make(map[string]bool)
	)

	for _, path := range pathItems {
		for _, operation := range path {
			schemas = append(schemas, operation.Schemas()...)
		}
	}

	// Remove duplicate schema names
	for _, n := range schemas {
		names[n] = true
	}

	schemas = []string{}
	for k := range names {
		schemas = append(schemas, k)
	}

	return schemas
}

func getResponseErrorCodesFromPaths(pathItems map[string]map[string]*Operation) map[pocketpb.ResponseCode]bool {
	stringCodeToResponseCode := func(code string) pocketpb.ResponseCode {
		switch code {
		case "400":
			return pocketpb.ResponseCode_RESPONSE_CODE_BAD_REQUEST
		case "401":
			return pocketpb.ResponseCode_RESPONSE_CODE_UNAUTHORIZED
		case "404":
			return pocketpb.ResponseCode_RESPONSE_CODE_NOT_FOUND
		case "412":
			return pocketpb.ResponseCode_RESPONSE_CODE_PRECONDITION_FAILED
		case "500":
			return pocketpb.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR
		}

		return pocketpb.ResponseCode_RESPONSE_CODE_OK
	}

	codes := make(map[pocketpb.ResponseCode]bool)
	for _, path := range pathItems {
		for _, operation := range path {
			for _, code := range operation.ResponseErrorCodes() {
				codes[stringCodeToResponseCode(code)] = true
			}
		}
	}

	return codes
}

func buildComponentsSchemas(schemaNames []string, options *parserOptions) map[string]*Schema {
	schemas := make(map[string]*Schema)

	for _, s := range schemaNames {
		name := trimPackagePath(s)
		if msg := findMessageByName(name, options.plugin); msg != nil {
			schema := messageToSchema(msg, options.enums, findProtogenMessageByName(name, options.plugin))
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
					if _, ok := schemas[refName]; !ok {
						for n, s := range buildComponentsSchemas([]string{refName}, options) {
							schemas[n] = s
						}
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
		fieldExtensions := pocket.GetFieldExtensions(f)
		if fieldExtensions.PropertyLocation() != pocketpb.HttpFieldLocation_HTTP_FIELD_LOCATION_BODY {
			continue
		}

		schemaOptions := &fieldToSchemaOptions{
			field:           f,
			enums:           enums,
			message:         message,
			msgSchema:       msgSchema,
			fieldExtensions: fieldExtensions,
		}

		if name, schema := fieldToSchema(schemaOptions); schema != nil {
			properties[name] = schema
		}
	}

	return NewSchema(&SchemaOptions{
		Type:       SchemaType_Object,
		Properties: properties,
	})
}

// responseErrorComponentsSchemas gives all error schemas that an API must have.
func responseErrorComponentsSchemas(errorCodes map[pocketpb.ResponseCode]bool) map[string]*Schema {
	var (
		dup     = make(map[pocketpb.ResponseCode]bool)
		schemas = make(map[string]*Schema)
	)

	if _, ok := errorCodes[pocketpb.ResponseCode_RESPONSE_CODE_BAD_REQUEST]; ok {
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
	delete(dup, pocketpb.ResponseCode_RESPONSE_CODE_BAD_REQUEST)

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

func parseServersFromFileExtensions(fileExtensions *pocket.FileExtensions) []*Server {
	var servers []*Server

	for _, server := range fileExtensions.Servers {
		servers = append(servers, &Server{
			Url:         server.GetUrl(),
			Description: server.GetDescription(),
		})
	}

	return servers
}
