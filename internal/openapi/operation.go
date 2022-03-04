package openapi

import (
	"fmt"
	"net/http"

	"google.golang.org/protobuf/compiler/protogen"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-krill-extensions/internal/krill"
)

const (
	refComponentsSchemas   = "#/components/schemas/"
	refComponentsResponses = "#/components/responses/"
)

type Operation struct {
	Name            string
	Summary         string `yaml:"summary"`
	Description     string `yaml:"description"`
	Id              string `yaml:"operationId"`
	Tags            []string
	Parameters      []*Parameter
	Responses       map[string]*Response
	RequestBody     *RequestBody          `yaml:"requestBody"`
	SecuritySchemes []map[string][]string `yaml:"security"`

	methodExtensions *krill.MethodExtensions
}

func (o *Operation) HasRequestBody() bool {
	return o.RequestBody != nil
}

// Schemas returns all schemas that are referenced across this Operation
// object.
func (o *Operation) Schemas() []string {
	schemas := []string{}

	// Adds the schema that the body is using
	if o.HasRequestBody() {
		for _, media := range o.RequestBody.Content {
			schemas = append(schemas, media.Schema.RefName())
		}
	}

	// Adds the schemas that the responses are using
	for _, response := range o.Responses {
		for _, media := range response.Content {
			schemas = append(schemas, media.Schema.RefName())
		}
	}

	return schemas
}

func (o *Operation) ResponseErrorCodes() []string {
	var (
		strStatusOk = fmt.Sprintf("%v", http.StatusOK)
		codes       = []string{}
	)

	for code := range o.Responses {
		if code != strStatusOk {
			codes = append(codes, code)
		}
	}

	return codes
}

func parseOperations(file *protogen.File, plugin *protogen.Plugin, enums map[string][]string) (map[string]map[string]*Operation, error) {
	var (
		service   = file.Proto.Service[0]
		pathItems = make(map[string]map[string]*Operation)
	)

	for _, method := range service.GetMethod() {
		extensions := krill.GetMethodExtensions(method)
		if extensions == nil {
			return nil, fmt.Errorf("cannot handle method '%s' without HTTP API definitions", method.GetName())
		}

		operation, err := newOperation(method, plugin, enums, extensions)
		if err != nil {
			return nil, err
		}

		httpMethod, endpoint := extensions.HttpMethodAndEndpoint()
		path, ok := pathItems[endpoint]
		if ok {
			path[httpMethod] = operation
		}
		if !ok {
			pathItems[endpoint] = map[string]*Operation{
				httpMethod: operation,
			}
		}
	}

	return pathItems, nil
}

func newOperation(method *descriptor.MethodDescriptorProto, plugin *protogen.Plugin, enums map[string][]string, extensions *krill.MethodExtensions) (*Operation, error) {
	var (
		requestBody     *RequestBody
		securitySchemes []map[string][]string
	)

	httpMethod := extensions.HttpMethod()
	if httpMethod == http.MethodPost || httpMethod == http.MethodPut {
		req, err := newRequestBody(method, plugin, extensions)
		if err != nil {
			return nil, err
		}
		requestBody = req
	}

	parameters, err := parseOperationParameters(
		method,
		plugin,
		extensions.EndpointDetails.Parameters,
		requestBody != nil,
		enums,
	)
	if err != nil {
		return nil, err
	}

	if extensions.HasKrillHttpExtension() {
		securitySchemes = append(securitySchemes,
			map[string][]string{
				"authorization": extensions.Method.GetHttp().GetScope(),
			})
	}

	responses, err := buildPathItemResponses(extensions, method, enums)
	if err != nil {
		return nil, err
	}

	return &Operation{
		Name:             extensions.HttpMethod(),
		Description:      extensions.OpenapiMethod.GetPath().GetDescription(),
		Summary:          extensions.OpenapiMethod.GetPath().GetSummary(),
		Id:               method.GetName(),
		Tags:             extensions.OpenapiMethod.GetPath().GetTags(),
		RequestBody:      requestBody,
		Parameters:       parameters,
		SecuritySchemes:  securitySchemes,
		Responses:        responses,
		methodExtensions: extensions,
	}, nil
}