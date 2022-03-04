package openapi

import (
	"fmt"
	"strings"

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

func parseOperationParameters(method *descriptor.MethodDescriptorProto, plugin *protogen.Plugin, endpointParameters []string, hasBody bool, enums map[string][]string, serviceExtensions *krill.ServiceExtensions) ([]*Parameter, error) {
	var (
		msgName                 = trimPackagePath(method.GetInputType())
		msgSchema               = findProtogenMessageByName(msgName, plugin)
		parameters              = []*Parameter{}
		globalHeaderMemberNames = serviceExtensions.GetGlobalHeaderMemberNames()
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

			if headerName, ok := globalHeaderMemberNames[name]; ok {
				delete(globalHeaderMemberNames, name)
				name = headerName
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

	if len(globalHeaderMemberNames) > 0 {
		return nil, fmt.Errorf("could not find header members '%v' in message '%s'",
			mapToString(globalHeaderMemberNames), msgName)
	}

	return parameters, nil
}

func mapToString(values map[string]string) string {
	sl := []string{}
	for k := range values {
		sl = append(sl, k)
	}

	return strings.Join(sl, ", ")
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
