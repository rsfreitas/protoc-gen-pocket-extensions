package openapi

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-krill-extensions/internal/krill"
)

const (
	refComponentsSchemas   = "#/components/schemas/"
	refComponentsResponses = "#/components/responses/"
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

		operation, err := parseOperation(method, plugin, enums, extensions)
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

func parseOperation(method *descriptor.MethodDescriptorProto, plugin *protogen.Plugin, enums map[string][]string, extensions *krill.MethodExtensions) (*Operation, error) {
	operation := &Operation{
		Name:        extensions.HttpMethod(),
		Description: extensions.OpenapiMethod.GetPath().GetDescription(),
		Summary:     extensions.OpenapiMethod.GetPath().GetSummary(),
		Id:          method.GetName(),
		Tags:        extensions.OpenapiMethod.GetPath().GetTags(),
	}

	operation.addsOperationSecuritySchemes(extensions)

	if err := operation.addsOperationRequestBody(extensions, method, plugin); err != nil {
		return nil, err
	}

	if err := operation.addsOperationParameters(extensions, method, plugin, enums); err != nil {
		return nil, err
	}

	if err := operation.addsOperationResponses(extensions, method, enums); err != nil {
		return nil, err
	}

	return operation, nil
}

func (o *Operation) addsOperationSecuritySchemes(extensions *krill.MethodExtensions) {
	if extensions.HasKrillHttpExtension() {
		o.SecuritySchemes = append(o.SecuritySchemes,
			map[string][]string{
				"authorization": extensions.Method.GetHttp().GetScope(),
			})
	}
}

func (o *Operation) addsOperationRequestBody(extensions *krill.MethodExtensions, method *descriptor.MethodDescriptorProto, plugin *protogen.Plugin) error {
	httpMethod := extensions.HttpMethod()
	if httpMethod == http.MethodPost || httpMethod == http.MethodPut {
		req, err := parseRequestBody(method, extensions, plugin)
		if err != nil {
			return err
		}

		o.RequestBody = req
	}

	return nil
}

// parseRequestBody builds up an openapi.RequestBody object according a
// protogen.Method.
func parseRequestBody(method *descriptor.MethodDescriptorProto, extensions *krill.MethodExtensions, plugin *protogen.Plugin) (*RequestBody, error) {
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

func (o *Operation) addsOperationParameters(extensions *krill.MethodExtensions, method *descriptor.MethodDescriptorProto, plugin *protogen.Plugin, enums map[string][]string) error {
	parameters, err := parseOperationParameters(
		method,
		plugin,
		extensions.EndpointDetails.Parameters,
		o.HasRequestBody(),
		enums,
	)
	if err != nil {
		return err
	}
	o.Parameters = parameters

	return nil
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

func (o *Operation) addsOperationResponses(extensions *krill.MethodExtensions, method *descriptor.MethodDescriptorProto, enums map[string][]string) error {
	res, err := buildPathItemResponses(extensions, method, enums)
	if err != nil {
		return err
	}
	o.Responses = res

	return nil
}

// buildPathItemResponses builds up all HTTP responses of a protobuf RPC method.
func buildPathItemResponses(extensions *krill.MethodExtensions, method *descriptor.MethodDescriptorProto, enums map[string][]string) (map[string]*Response, error) {
	// containsCode checks inside the method responses for a specific response
	// code.
	containsCode := func(code string) (int, bool) {
		for index, res := range extensions.OpenapiMethod.GetResponse() {
			if res.GetCode().String() == code {
				return index, true
			}
		}

		return 0, false
	}

	var (
		responses = make(map[string]*Response)
	)

	if index, ok := containsCode("RESPONSE_CODE_OK"); ok {
		res := extensions.OpenapiMethod.GetResponse()[index]
		responses["200"] = &Response{
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

	if index, ok := containsCode("RESPONSE_CODE_BAD_REQUEST"); ok {
		res := extensions.OpenapiMethod.GetResponse()[index]
		responses["400"] = &Response{
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

	if index, ok := containsCode("RESPONSE_CODE_UNAUTHORIZED"); ok {
		res := extensions.OpenapiMethod.GetResponse()[index]
		responses["401"] = &Response{
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

	if index, ok := containsCode("RESPONSE_CODE_NOT_FOUND"); ok {
		res := extensions.OpenapiMethod.GetResponse()[index]
		responses["404"] = &Response{
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

	if index, ok := containsCode("RESPONSE_CODE_PRECONDITION_FAILED"); ok {
		res := extensions.OpenapiMethod.GetResponse()[index]
		responses["412"] = &Response{
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

	if index, ok := containsCode("RESPONSE_CODE_INTERNAL_ERROR"); ok {
		res := extensions.OpenapiMethod.GetResponse()[index]
		responses["500"] = &Response{
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

func trimPackagePath(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

func findMessageByName(name string, plugin *protogen.Plugin) *descriptor.DescriptorProto {
	for _, file := range plugin.FilesByPath {
		for _, m := range file.Proto.MessageType {
			if name == m.GetName() {
				return m
			}
		}
	}

	return nil
}

func findProtogenMessageByName(name string, plugin *protogen.Plugin) *protogen.Message {
	for _, file := range plugin.FilesByPath {
		for _, m := range file.Messages {
			if name == m.GoIdent.GoName {
				return m
			}
		}
	}

	return nil
}

func fieldToSchema(field *descriptor.FieldDescriptorProto, enums map[string][]string, message *descriptor.DescriptorProto, msgSchema *protogen.Message, fieldExtensions *krill.FieldExtensions) (string, *Schema) {
	var (
		fieldName = field.GetName()
		opts      = &SchemaOptions{}
		property  = fieldExtensions.Openapi.GetProperty()
	)

	parseFieldType(field, opts, fieldExtensions)

	if field.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		// Creates the Item.Schema
		o := &SchemaOptions{
			Ref: opts.Ref,
		}
		if o.Ref == "" {
			o.Type = opts.Type
		}

		// Change the main type to array
		opts.Type = SchemaType_Array
		opts.Items = NewSchema(o)
		opts.Ref = ""
	}

	if field.GetType() == descriptor.FieldDescriptorProto_TYPE_ENUM {
		for name, values := range enums {
			if trimPackagePath(field.GetTypeName()) == name {
				opts.Enum = values
			}
		}
	}

	if isFieldRequired(fieldExtensions) {
		opts.Required = true
	}

	if property != nil {
		if property.GetHideFromSchema() {
			return "", nil
		}

		opts.Example = property.GetExample()
		opts.Description = property.GetDescription()

		if opts.Format == "" {
			format := property.GetFormat().String()
			if format != "PROPERTY_FORMAT_UNSPECIFIED" && format != "PROPERTY_FORMAT_STRING" {
				opts.Format = strcase.ToKebab(strings.TrimPrefix(format, "PROPERTY_FORMAT_"))
			}
		}

	}

	return fieldName, NewSchema(opts)
}

func parseFieldType(field *descriptor.FieldDescriptorProto, opts *SchemaOptions, fieldExtensions *krill.FieldExtensions) {
	switch field.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		opts.Type = SchemaType_String
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		opts.Type = SchemaType_String
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		opts.Type = SchemaType_Bool
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE, descriptor.FieldDescriptorProto_TYPE_FLOAT:
		opts.Type = SchemaType_Number
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		if isProtobufWrapper(field.GetTypeName()) {
			switch getWrapperType(field.GetTypeName()) {
			case "String":
				opts.Type = SchemaType_String
			case "Float", "Double":
				opts.Type = SchemaType_Number
			case "Bool":
				opts.Type = SchemaType_Bool
			default:
				opts.Type = SchemaType_Integer
			}

			return
		}

		if isProtobufTimestamp(field.GetTypeName()) {
			opts.Type = SchemaType_String
			opts.Format = "date-time"
			return
		}

		if isProtobufValue(field.GetTypeName()) {
			// TODO: handle google.protobuf.Value type
			return
		}

		opts.Ref = refComponentsSchemas + trimPackagePath(field.GetTypeName())
		opts.Type = SchemaType_Object
	default:
		opts.Type = SchemaType_Integer
	}
}

func isFieldRequired(fieldExtensions *krill.FieldExtensions) bool {
	if fieldExtensions != nil && fieldExtensions.Openapi != nil {
		if p := fieldExtensions.Openapi.GetProperty(); p != nil {
			return p.GetRequired()
		}
	}

	return false
}

func isProtobufWrapper(name string) bool {
	re := regexp.MustCompile(`google\.protobuf\..+Value`)
	return re.MatchString(name)
}

func getWrapperType(name string) string {
	re := regexp.MustCompile(`google\.protobuf\.(.+)Value`)
	s := re.FindStringSubmatch(name)
	return s[1]
}

func isProtobufTimestamp(name string) bool {
	return strings.Contains(name, "google.protobuf.Timestamp")
}

func isProtobufValue(name string) bool {
	return strings.Contains(name, "google.protobuf.Value")
}
