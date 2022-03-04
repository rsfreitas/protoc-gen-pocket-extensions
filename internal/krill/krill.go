package krill

import (
	"regexp"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	krillpb "github.com/rsfreitas/protoc-gen-krill-extensions/options/krill"
)

type FileExtensions struct {
	AppName        string
	OpenapiTitle   string
	OpenapiVersion string
	Servers        []*krillpb.OpenapiServer
}

type ServiceExtensions struct {
	Service *krillpb.Service
}

type MethodExtensions struct {
	GoogleApi       *annotations.HttpRule
	Method          *krillpb.Method
	OpenapiMethod   *krillpb.OpenapiMethod
	EndpointDetails *HttpEndpointDetails
}

// HttpEndpointDetails gathers detailed information about a HTTP endpoint of
// a method.
type HttpEndpointDetails struct {
	Method     string
	Parameters []string
	Body       string
}

type MessageExtensions struct {
	OpenapiMessage *krillpb.OpenapiMessage
}

type FieldExtensions struct {
	Database *krillpb.Database
	Openapi  *krillpb.OpenapiField
}

func (e *MethodExtensions) HasKrillHttpExtension() bool {
	return e.Method != nil && e.Method.GetHttp() != nil
}

func (e *MethodExtensions) HttpMethodAndEndpoint() (string, string) {
	return getMethodAndEndpoint(e.GoogleApi)
}

func (e *MethodExtensions) HttpMethod() string {
	if e.EndpointDetails != nil {
		return strings.ToUpper(e.EndpointDetails.Method)
	}

	return ""
}

func (f *FieldExtensions) PropertyLocation() krillpb.PropertyLocation {
	if f.Openapi != nil {
		if p := f.Openapi.GetProperty(); p != nil {
			return p.GetLocation()
		}
	}

	return krillpb.PropertyLocation_PROPERTY_LOCATION_BODY
}

func GetServiceExtensions(service *descriptor.ServiceDescriptorProto) *ServiceExtensions {
	if service.Options != nil {
		s := proto.GetExtension(service.Options, krillpb.E_Service)

		if svc, ok := s.(*krillpb.Service); ok {
			return &ServiceExtensions{
				Service: svc,
			}
		}
	}

	return nil
}

func GetMethodExtensions(method *descriptor.MethodDescriptorProto) *MethodExtensions {
	googleApi := getGoogleHttpAPIIfAny(method)

	return &MethodExtensions{
		GoogleApi:       googleApi,
		Method:          getKrillMethodExtension(method),
		OpenapiMethod:   getKrillOpenapiMethodExtension(method),
		EndpointDetails: getEndpointParameters(googleApi),
	}
}

func getKrillOpenapiMethodExtension(method *descriptor.MethodDescriptorProto) *krillpb.OpenapiMethod {
	if method.Options != nil {
		m := proto.GetExtension(method.Options, krillpb.E_OpenapiMethod)
		return (m.(*krillpb.OpenapiMethod))
	}

	return nil
}

func getKrillMethodExtension(method *descriptor.MethodDescriptorProto) *krillpb.Method {
	if method.Options != nil {
		m := proto.GetExtension(method.Options, krillpb.E_Method)
		return (m.(*krillpb.Method))
	}

	return nil
}

// getGoogleHttpAPIIfAny gets the google.api.http extension of a method if exists.
func getGoogleHttpAPIIfAny(msg *descriptor.MethodDescriptorProto) *annotations.HttpRule {
	if msg.Options != nil {
		h := proto.GetExtension(msg.Options, annotations.E_Http)
		return (h.(*annotations.HttpRule))
	}

	return nil
}

func getEndpointParameters(googleApi *annotations.HttpRule) *HttpEndpointDetails {
	details := &HttpEndpointDetails{}

	if googleApi != nil {
		method, endpoint := getMethodAndEndpoint(googleApi)
		parameters := retrieveEndpointParameters(endpoint)
		parameters = append(parameters, retrieveEndpointParametersFromAdditionalBindings(googleApi)...)

		details.Body = googleApi.GetBody()
		details.Parameters = parameters
		details.Method = strings.ToUpper(method)
	}

	return details
}

// getMethodAndEndpoint translates a google.api.http notation of a request
// type to our supported type.
func getMethodAndEndpoint(rule *annotations.HttpRule) (string, string) {
	method := ""
	endpoint := ""

	switch rule.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		method = "get"
		endpoint = rule.GetGet()

	case *annotations.HttpRule_Post:
		method = "post"
		endpoint = rule.GetPost()

	case *annotations.HttpRule_Put:
		method = "put"
		endpoint = rule.GetPut()

	case *annotations.HttpRule_Delete:
		method = "delete"
		endpoint = rule.GetDelete()

	case *annotations.HttpRule_Patch:
		method = "patch"
		endpoint = rule.GetPatch()
	}

	return method, endpoint
}

func retrieveEndpointParameters(endpoint string) []string {
	var parameters []string
	re := regexp.MustCompile(`{[A-Za-z_.0-9]*}`)

	for _, p := range re.FindAll([]byte(endpoint), -1) {
		parameters = append(parameters, string(p[1:len(p)-1]))
	}

	return parameters
}

func retrieveEndpointParametersFromAdditionalBindings(rule *annotations.HttpRule) []string {
	var parameters []string

	for _, r := range rule.GetAdditionalBindings() {
		if _, endpoint := getMethodAndEndpoint(r); endpoint != "" {
			parameters = append(parameters, retrieveEndpointParameters(endpoint)...)
		}
	}

	return parameters
}

func GetMessageExtensions(message *descriptor.DescriptorProto) *MessageExtensions {
	return &MessageExtensions{
		OpenapiMessage: getKrillOpenapiMessageExtension(message),
	}
}

func getKrillOpenapiMessageExtension(message *descriptor.DescriptorProto) *krillpb.OpenapiMessage {
	if message.Options != nil {
		m := proto.GetExtension(message.Options, krillpb.E_OpenapiMessage)
		return (m.(*krillpb.OpenapiMessage))
	}

	return nil
}

func GetFieldExtensions(field *descriptor.FieldDescriptorProto) *FieldExtensions {
	ext := &FieldExtensions{}

	if field != nil && field.Options != nil {
		f := proto.GetExtension(field.Options, krillpb.E_Openapi)
		if p, ok := f.(*krillpb.OpenapiField); ok {
			ext.Openapi = p
		}

		d := proto.GetExtension(field.Options, krillpb.E_Database)
		if p, ok := d.(*krillpb.Database); ok {
			ext.Database = p
		}
	}

	return ext
}

func GetFileExtensions(file *descriptor.FileDescriptorProto) *FileExtensions {
	var (
		name    string
		title   string
		version string
		servers []*krillpb.OpenapiServer
	)

	if file.Options != nil {
		if n := proto.GetExtension(file.Options, krillpb.E_KrillAppName); n != nil {
			name = n.(string)
		}

		if n := proto.GetExtension(file.Options, krillpb.E_KrillOpenapiTitle); n != nil {
			title = n.(string)
		}

		if n := proto.GetExtension(file.Options, krillpb.E_KrillOpenapiVersion); n != nil {
			version = n.(string)
		}

		if s := proto.GetExtension(file.Options, krillpb.E_KrillOpenapiServer); s != nil {
			servers = s.([]*krillpb.OpenapiServer)
		}
	}

	return &FileExtensions{
		AppName:        name,
		OpenapiTitle:   title,
		OpenapiVersion: version,
		Servers:        servers,
	}
}

func ResponseCodeToHttpCode(code krillpb.ResponseCode) string {
	switch code {
	case krillpb.ResponseCode_RESPONSE_CODE_OK:
		return "200"
	case krillpb.ResponseCode_RESPONSE_CODE_NOT_FOUND:
		return "404"
	case krillpb.ResponseCode_RESPONSE_CODE_BAD_REQUEST:
		return "400"
	case krillpb.ResponseCode_RESPONSE_CODE_UNAUTHORIZED:
		return "401"
	case krillpb.ResponseCode_RESPONSE_CODE_PRECONDITION_FAILED:
		return "412"
	}

	// Internal Error
	return "500"
}
