package openapi

import (
	"fmt"
	"net/http"

	"google.golang.org/protobuf/compiler/protogen"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-pocket-extensions/internal/pocket"
)

type RequestBody struct {
	Required    bool
	Description string
	Content     map[string]*Media
}

func newRequestBody(method *descriptor.MethodDescriptorProto, plugin *protogen.Plugin, extensions *pocket.MethodExtensions) (*RequestBody, error) {
	var (
		required      = extensions.HttpMethod() == http.MethodPost
		description   = getRequestBodyDescription(method, plugin)
		refSchemaName string
	)

	if extensions.HttpMethod() == http.MethodPost {
		refSchemaName = method.GetInputType()
	}

	if extensions.HttpMethod() == http.MethodPut {
		name, err := getRequestBodyRefSchemaNameForPut(method, plugin, extensions)
		if err != nil {
			return nil, err
		}
		refSchemaName = name
	}

	return &RequestBody{
		Required:    required,
		Description: description,
		Content: map[string]*Media{
			"application/json": NewMedia(
				NewSchema(&SchemaOptions{
					Ref: refComponentsSchemas + trimPackagePath(refSchemaName),
				}),
			),
		},
	}, nil
}

func getRequestBodyDescription(method *descriptor.MethodDescriptorProto, plugin *protogen.Plugin) string {
	messageExtensions := pocket.GetMessageExtensions(
		findMessageByName(trimPackagePath(method.GetInputType()), plugin),
	)

	if messageExtensions.OpenapiMessage != nil {
		if op := messageExtensions.OpenapiMessage.GetOperation(); op != nil {
			if rb := op.GetRequestBody(); rb != nil {
				return rb.GetDescription()
			}
		}
	}

	return ""
}

func getRequestBodyRefSchemaNameForPut(method *descriptor.MethodDescriptorProto, plugin *protogen.Plugin, extensions *pocket.MethodExtensions) (string, error) {
	// We shouldn't find a body annotated as "*" but we suport it.
	if extensions.EndpointDetails.Body == "*" {
		return method.GetInputType(), nil
	}

	msgName := trimPackagePath(method.GetInputType())
	msg := findMessageByName(msgName, plugin)
	if msg == nil {
		return "", fmt.Errorf("could not find message with name '%s'", msgName)
	}

	for _, f := range msg.Field {
		if f.GetName() == extensions.EndpointDetails.Body {
			return f.GetTypeName(), nil
		}
	}

	return "", fmt.Errorf("could not find member '%s' for the request body", extensions.EndpointDetails.Body)
}
