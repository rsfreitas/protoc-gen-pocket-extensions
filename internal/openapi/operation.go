package openapi

import (
	"fmt"
	"net/http"

	"google.golang.org/protobuf/compiler/protogen"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-krill-extensions/internal/krill"
)

type Operation struct {
	Name             string
	Summary          string `yaml:"summary"`
	Description      string `yaml:"description"`
	Id               string `yaml:"operationId"`
	Tags             []string
	Parameters       []*Parameter
	Responses        map[string]*Response
	RequestBody      *RequestBody          `yaml:"requestBody"`
	SecuritySchemes  []map[string][]string `yaml:"security"`
	MethodExtensions *krill.MethodExtensions
}

type Parameter struct {
	Required    bool
	Location    string `yaml:"in"`
	Name        string
	Description string
	Schema      *Schema
}

type RequestBody struct {
	Required    bool
	Description string
	Content     map[string]*Media
}

type Response struct {
	Description string
	Content     map[string]*Media
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
		httpMethod, endpoint := extensions.HttpMethodAndEndpoint()

		operation, err := parseOperation(method, plugin, enums, extensions)
		if err != nil {
			return nil, err
		}

		path, ok := pathItems[endpoint]
		if !ok {
			pathItems[endpoint] = map[string]*Operation{
				httpMethod: operation,
			}
		}
		if ok {
			path[httpMethod] = operation
		}
	}

	return pathItems, nil
}

func parseOperation(method *descriptor.MethodDescriptorProto, plugin *protogen.Plugin, enums map[string][]string, extensions *krill.MethodExtensions) (*Operation, error) {
	//	operationpb := getMessageOperationExtension(findMessageByName(trimPackagePath(method.GetInputType()), plugin))
	var (
		httpMethod = extensions.HttpMethod()
		hasBody    = false
	)

	operation := &Operation{
		Name:        httpMethod,
		Description: extensions.OpenapiMethod.GetPath().GetDescription(),
		Summary:     extensions.OpenapiMethod.GetPath().GetSummary(),
		Id:          method.GetName(),
		Tags:        extensions.OpenapiMethod.GetPath().GetTags(),
	}

	extensionsOperations := getMethodOperationsExtension(method)
	if extensionsOperations != nil {
		operation.SecuritySchemes = append(operation.SecuritySchemes,
			map[string][]string{
				"authorization": extensionsOperations.GetScope(),
			})
	}

	endpointParameters := retrieveEndpointParameters(endpoint)

	if httpMethod == http.MethodPost || httpMethod == http.MethodPut {
		req, err := parseRequestBody(httpMethod, method, operationpb)
		if err != nil {
			return nil, err
		}
		operation.RequestBody = req
		hasBody = true
	}

	if httpMethod != http.MethodPost {
		parameters, err := parseOperationParameters(method, plugin, endpointParameters, hasBody, enums)
		if err != nil {
			return nil, err
		}
		operation.Parameters = parameters
	}

	if extensionsOperations != nil {
		parameters, err := parseOperationHeaderParameters(extensionsOperations, method, plugin, enums)
		if err != nil {
			return nil, err
		}
		if parameters != nil {
			operation.Parameters = append(operation.Parameters, parameters...)
		}
	}

	res, err := buildPathItemResponses(method, enums)
	if err != nil {
		return nil, err
	}
	operation.Responses = res

	return operation, nil
}
