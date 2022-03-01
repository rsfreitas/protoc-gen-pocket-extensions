package openapi

import (
	"net/http"

	"google.golang.org/protobuf/compiler/protogen"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-krill-extensions/internal/krill"
)

type RequestBody struct {
	Required    bool
	Description string
	Content     map[string]*Media
}

func newRequestBody(method *descriptor.MethodDescriptorProto, plugin *protogen.Plugin, extensions *krill.MethodExtensions) (*RequestBody, error) {
	var (
		required    = extensions.HttpMethod() == http.MethodPost
		description = getRequestBodyDescription(method, plugin)
	)

	return &RequestBody{
		Required:    required,
		Description: description,
		Content: map[string]*Media{
			"application/json": NewMedia(
				NewSchema(&SchemaOptions{
					Ref: refComponentsSchemas + trimPackagePath(method.GetInputType()),
				}),
			),
		},
	}, nil
}

func getRequestBodyDescription(method *descriptor.MethodDescriptorProto, plugin *protogen.Plugin) string {
	messageExtensions := krill.GetMessageExtensions(
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
