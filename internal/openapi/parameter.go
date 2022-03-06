package openapi

import (
	"fmt"
	"strings"

	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-krill-extensions/internal/krill"
	krillpb "github.com/rsfreitas/protoc-gen-krill-extensions/options/krill"
)

type Parameter struct {
	Required    bool
	Location    string `yaml:"in"`
	Name        string
	Description string
	Schema      *Schema
}

func parseOperationParameters(method *descriptor.MethodDescriptorProto, options *parserOptions, methodExtensions *krill.MethodExtensions) ([]*Parameter, error) {
	var (
		msgName           = trimPackagePath(method.GetInputType())
		msgSchema         = findProtogenMessageByName(msgName, options.plugin)
		parameters        = []*Parameter{}
		headerMemberNames = getHeaderMemberNames(options.serviceExtensions, methodExtensions)
	)

	msg := findMessageByName(msgName, options.plugin)
	if msg == nil {
		return nil, fmt.Errorf("could not find message with name '%s'", msgName)
	}

	for _, f := range msg.Field {
		fieldExtensions := krill.GetFieldExtensions(f)

		// We don't need to parse body parameters
		if fieldExtensions.PropertyLocation() == krillpb.HttpFieldLocation_HTTP_FIELD_LOCATION_BODY {
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
			if fieldExtensions.PropertyLocation() == krillpb.HttpFieldLocation_HTTP_FIELD_LOCATION_PATH {
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

func toOpenapiLocation(location krillpb.HttpFieldLocation) string {
	switch location {
	case krillpb.HttpFieldLocation_HTTP_FIELD_LOCATION_BODY:
		return "body"

	case krillpb.HttpFieldLocation_HTTP_FIELD_LOCATION_PATH:
		return "path"

	case krillpb.HttpFieldLocation_HTTP_FIELD_LOCATION_HEADER:
		return "header"

	case krillpb.HttpFieldLocation_HTTP_FIELD_LOCATION_QUERY:
		return "query"
	}

	return "unknown"
}
