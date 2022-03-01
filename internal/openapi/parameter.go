package openapi

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-krill-extensions/internal/krill"
)

type Parameter struct {
	Required    bool
	Location    string `yaml:"in"`
	Name        string
	Description string
	Schema      *Schema
}

func parseOperationParameters(method *descriptor.MethodDescriptorProto, plugin *protogen.Plugin, endpointParameters []string, hasBody bool, enums map[string][]string) ([]*Parameter, error) {
	var (
		msgName    = trimPackagePath(method.GetInputType())
		msgSchema  = findProtogenMessageByName(msgName, plugin)
		parameters = []*Parameter{}
	)

	msg := findMessageByName(msgName, plugin)
	if msg == nil {
		return nil, fmt.Errorf("could not find message with name '%s'", msgName)
	}

	for _, f := range msg.Field {
		fieldExtensions := krill.GetFieldExtensions(f)

		// We don't need to parse body parameters
		if fieldExtensions.PropertyLocation().String() == "PROPERTY_LOCATION_BODY" {
			continue
		}

		if name, schema := fieldToSchema(f, enums, msg, msgSchema, fieldExtensions); schema != nil {
			required := schema.IsRequired()
			if fieldExtensions.PropertyLocation().String() == "PROPERTY_LOCATION_PATH" {
				// The field is always required when it's located at the endpoint
				// path
				required = true
			}

			parameters = append(parameters, &Parameter{
				Location: toOpenapiLocation(fieldExtensions.PropertyLocation().String()),
				Name:     name,
				Schema: NewSchema(&SchemaOptions{
					Type:    schema.SchemaType(),
					Format:  schema.Format,
					Example: schema.Example,
				}),
				Required:    required,
				Description: schema.Description,
			})
		}
	}

	return parameters, nil
}

func toOpenapiLocation(location string) string {
	switch location {
	case "PROPERTY_LOCATION_BODY":
		return "body"

	case "PROPERTY_LOCATION_PATH":
		return "path"

	case "PROPERTY_LOCATION_HEADER":
		return "header"

	case "PROPERTY_LOCATION_QUERY":
		return "query"
	}

	return "unknown"
}
