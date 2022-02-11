package proto

import (
	"regexp"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	micropb "github.com/rsfreitas/protoc-gen-micro-extensions/options/micro"
)

// methodExtensions gathers all extensions that a method might have inside its
// protobuf declaration.
type methodExtensions struct {
	GoogleApi       *annotations.HttpRule
	Micro           *micropb.Method
	EndpointDetails *httpEndpointDetails
}

// httpEndpointDetails gathers detailed information about a HTTP endpoint of
// a method.
type httpEndpointDetails struct {
	Method     string
	Parameters []string
	Body       string
}

func (e *methodExtensions) HasMicroHttpExtension() bool {
	return e.Micro != nil && e.Micro.GetHttp() != nil
}

func (e *methodExtensions) HttpMethodAndEndpoint() (string, string) {
	return getMethodAndEndpoint(e.GoogleApi)
}

func loadMethodExtensions(msg *descriptor.MethodDescriptorProto) *methodExtensions {
	googleApi := getGoogleHttpAPIIfAny(msg)

	return &methodExtensions{
		GoogleApi:       googleApi,
		Micro:           getMicroMethodExtension(msg),
		EndpointDetails: getEndpointParameters(googleApi),
	}
}

func getMicroMethodExtension(msg *descriptor.MethodDescriptorProto) *micropb.Method {
	if msg.Options != nil {
		m := proto.GetExtension(msg.Options, micropb.E_Method)
		return (m.(*micropb.Method))
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

func getEndpointParameters(googleApi *annotations.HttpRule) *httpEndpointDetails {
	details := &httpEndpointDetails{}

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
