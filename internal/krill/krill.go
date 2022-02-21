package krill

import (
	"regexp"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	krillpb "github.com/rsfreitas/protoc-gen-krill-extensions/options/krill"
)

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
