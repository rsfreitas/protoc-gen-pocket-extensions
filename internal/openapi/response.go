package openapi

import (
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-pocket-extensions/internal/pocket"
	pocketpb "github.com/rsfreitas/protoc-gen-pocket-extensions/options/pocket"
)

type Response struct {
	Description string
	Content     map[string]*Media
}

// buildPathItemResponses builds up all HTTP responses of a protobuf RPC method.
func buildPathItemResponses(extensions *pocket.MethodExtensions, method *descriptor.MethodDescriptorProto) (map[string]*Response, error) {
	// containsCode checks inside the method responses for a specific response
	// code.
	containsCode := func(code pocketpb.ResponseCode) (int, bool) {
		for index, res := range extensions.OpenapiMethod.GetResponse() {
			if res.GetCode() == code {
				return index, true
			}
		}

		return 0, false
	}

	var (
		responses = make(map[string]*Response)
	)

	if index, ok := containsCode(pocketpb.ResponseCode_RESPONSE_CODE_OK); ok {
		res := extensions.OpenapiMethod.GetResponse()[index]
		responses[pocket.ResponseCodeToHttpCode(res.GetCode())] = &Response{
			Description: res.GetDescription(),
			Content: map[string]*Media{
				"application/json": NewMedia(
					NewSchema(&SchemaOptions{
						Ref: refComponentsSchemas + trimPackagePath(method.GetOutputType()),
					}),
				),
			},
		}
	}

	if index, ok := containsCode(pocketpb.ResponseCode_RESPONSE_CODE_CREATED); ok {
		res := extensions.OpenapiMethod.GetResponse()[index]
		responses[pocket.ResponseCodeToHttpCode(res.GetCode())] = &Response{
			Description: res.GetDescription(),
			Content: map[string]*Media{
				"application/json": NewMedia(
					NewSchema(&SchemaOptions{
						Ref: refComponentsSchemas + trimPackagePath(method.GetOutputType()),
					}),
				),
			},
		}
	}

	if index, ok := containsCode(pocketpb.ResponseCode_RESPONSE_CODE_BAD_REQUEST); ok {
		res := extensions.OpenapiMethod.GetResponse()[index]
		responses[pocket.ResponseCodeToHttpCode(res.GetCode())] = &Response{
			Description: res.GetDescription(),
			Content: map[string]*Media{
				"application/json": NewMedia(
					NewSchema(&SchemaOptions{
						Ref: refComponentsSchemas + "ValidationError",
					}),
				),
			},
		}
	}

	if index, ok := containsCode(pocketpb.ResponseCode_RESPONSE_CODE_UNAUTHORIZED); ok {
		res := extensions.OpenapiMethod.GetResponse()[index]
		responses[pocket.ResponseCodeToHttpCode(res.GetCode())] = &Response{
			Description: res.GetDescription(),
			Content: map[string]*Media{
				"application/json": NewMedia(
					NewSchema(&SchemaOptions{
						Ref: refComponentsSchemas + "DefaultError",
					}),
				),
			},
		}
	}

	if index, ok := containsCode(pocketpb.ResponseCode_RESPONSE_CODE_NOT_FOUND); ok {
		res := extensions.OpenapiMethod.GetResponse()[index]
		responses[pocket.ResponseCodeToHttpCode(res.GetCode())] = &Response{
			Description: res.GetDescription(),
			Content: map[string]*Media{
				"application/json": NewMedia(
					NewSchema(&SchemaOptions{
						Ref: refComponentsSchemas + "DefaultError",
					}),
				),
			},
		}
	}

	if index, ok := containsCode(pocketpb.ResponseCode_RESPONSE_CODE_PRECONDITION_FAILED); ok {
		res := extensions.OpenapiMethod.GetResponse()[index]
		responses[pocket.ResponseCodeToHttpCode(res.GetCode())] = &Response{
			Description: res.GetDescription(),
			Content: map[string]*Media{
				"application/json": NewMedia(
					NewSchema(&SchemaOptions{
						Ref: refComponentsSchemas + "DefaultError",
					}),
				),
			},
		}
	}

	if index, ok := containsCode(pocketpb.ResponseCode_RESPONSE_CODE_INTERNAL_ERROR); ok {
		res := extensions.OpenapiMethod.GetResponse()[index]
		responses[pocket.ResponseCodeToHttpCode(res.GetCode())] = &Response{
			Description: res.GetDescription(),
			Content: map[string]*Media{
				"application/json": NewMedia(
					NewSchema(&SchemaOptions{
						Ref: refComponentsSchemas + "DefaultError",
					}),
				),
			},
		}
	}

	return responses, nil
}
