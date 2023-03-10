package openapi

import (
	"fmt"
	"strings"

	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-pocket-extensions/internal/pocket"
	pocketpb "github.com/rsfreitas/protoc-gen-pocket-extensions/options/pocket"
)

type Parameter struct {
	Required    bool
	Location    string `yaml:"in"`
	Name        string
	Description string
	Schema      *Schema
}

func parseOperationParameters(method *descriptor.MethodDescriptorProto, options *parserOptions, methodExtensions *pocket.MethodExtensions) ([]*Parameter, error) {
	var (
		msgName           = trimPackagePath(method.GetInputType())
		msgSchema         = findProtogenMessageByName(msgName, options.plugin)
		parameters        []*Parameter
		headerMemberNames = getHeaderMemberNames(options.serviceExtensions, methodExtensions)
	)

	msg := findMessageByName(msgName, options.plugin)
	if msg == nil {
		return nil, fmt.Errorf("could not find message with name '%s'", msgName)
	}

	for _, f := range msg.Field {
		fieldExtensions := pocket.GetFieldExtensions(f)

		// We don't need to parse body parameters
		if fieldExtensions.PropertyLocation() == pocketpb.HttpFieldLocation_HTTP_FIELD_LOCATION_BODY {
			continue
		}

		schemaOptions := &fieldToSchemaOptions{
			field:           f,
			enums:           options.enums,
			message:         msg,
			msgSchema:       msgSchema,
			fieldExtensions: fieldExtensions,
		}

		if name, schema := fieldToSchema(schemaOptions); schema != nil {
			required := schema.IsRequired()
			if fieldExtensions.PropertyLocation() == pocketpb.HttpFieldLocation_HTTP_FIELD_LOCATION_PATH {
				// The field is always required when it's located at the endpoint
				// path
				required = true
			}

			if headerName, ok := headerMemberNames[name]; ok {
				delete(headerMemberNames, name)
				name = headerName
			}

			parameters = append(parameters, &Parameter{
				Location: toOpenapiLocation(fieldExtensions.PropertyLocation()),
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

func getHeaderMemberNames(serviceExtensions *pocket.ServiceExtensions, methodExtensions *pocket.MethodExtensions) map[string]string {
	var (
		global = serviceExtensions.GetHeaderMemberNames()
		local  = methodExtensions.GetHeaderMemberNames()
	)

	if global == nil {
		global = make(map[string]string)
	}

	for k, v := range local {
		global[k] = v
	}

	return global
}

func mapToString(values map[string]string) string {
	var sl []string
	for k := range values {
		sl = append(sl, k)
	}

	return strings.Join(sl, ", ")
}

func toOpenapiLocation(location pocketpb.HttpFieldLocation) string {
	switch location {
	case pocketpb.HttpFieldLocation_HTTP_FIELD_LOCATION_BODY:
		return "body"

	case pocketpb.HttpFieldLocation_HTTP_FIELD_LOCATION_PATH:
		return "path"

	case pocketpb.HttpFieldLocation_HTTP_FIELD_LOCATION_HEADER:
		return "header"

	case pocketpb.HttpFieldLocation_HTTP_FIELD_LOCATION_QUERY:
		return "query"
	}

	return "unknown"
}
