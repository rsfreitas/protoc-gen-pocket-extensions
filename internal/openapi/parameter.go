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

func parseOperationParameters(method *descriptor.MethodDescriptorProto, plugin *protogen.Plugin, enums map[string][]string, serviceExtensions *krill.ServiceExtensions, methodExtensions *krill.MethodExtensions) ([]*Parameter, error) {
	var (
		msgName           = trimPackagePath(method.GetInputType())
		msgSchema         = findProtogenMessageByName(msgName, plugin)
		parameters        = []*Parameter{}
		headerMemberNames = getHeaderMemberNames(serviceExtensions, methodExtensions)
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

			if headerName, ok := headerMemberNames[name]; ok {
				delete(headerMemberNames, name)
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

	if len(headerMemberNames) > 0 {
		return nil, fmt.Errorf("could not find header members '%v' in message '%s'",
			mapToString(headerMemberNames), msgName)
	}

	return parameters, nil
}

func getHeaderMemberNames(serviceExtensions *krill.ServiceExtensions, methodExtensions *krill.MethodExtensions) map[string]string {
	var (
		global = serviceExtensions.GetHeaderMemberNames()
		local  = methodExtensions.GetHeaderMemberNames()
	)

	for k, v := range local {
		global[k] = v
	}

	return global
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
