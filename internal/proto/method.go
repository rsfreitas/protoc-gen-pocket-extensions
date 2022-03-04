package proto

import (
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/rsfreitas/protoc-gen-krill-extensions/internal/krill"
)

type Method struct {
	Name       string
	Input      *MethodMessage
	Output     *MethodMessage
	extensions *krill.MethodExtensions
}

type MethodMessage struct {
	Name       string
	Parameters []*Parameter
}

// HasAuthentication returns true or false if the current Method has
// authentication enabled or not.
func (m *Method) HasAuthentication() bool {
	if http := m.extensions.Method.GetHttp(); http != nil {
		return !http.GetNoAuth()
	}

	return false
}

// HasBody returns true or false if the current Method needs to parse the
// request body or not.
func (m *Method) HasBody() bool {
	if m.Input != nil {
		for _, p := range m.Input.Parameters {
			if p.Location == ParameterLocation_Body {
				return true
			}
		}
	}

	return false
}

// PathParameters gives a slice of parameters that must be read from the the
// request URL path.
func (m *Method) PathParameters() []*Parameter {
	var parameters []*Parameter

	if m.Input != nil {
		for _, p := range m.Input.Parameters {
			if p.Location == ParameterLocation_Path {
				parameters = append(parameters, p)
			}
		}
	}

	return parameters
}

// QueryParameters gives a slice of parameters that must be read from the the
// request URL as query parameters.
func (m *Method) QueryParameters() []*Parameter {
	var parameters []*Parameter

	if m.Input != nil {
		for _, p := range m.Input.Parameters {
			if p.Location == ParameterLocation_Query {
				parameters = append(parameters, p)
			}
		}
	}

	return parameters
}

// RocketEndpoint converts the method endpoint to the rocket syntax.
func (m *Method) RocketEndpoint() string {
	_, endpoint := m.extensions.HttpMethodAndEndpoint()
	re := strings.NewReplacer("{", "<", "}", ">")
	endpoint = re.Replace(endpoint)

	return m.addQueryParameters(endpoint)
}

func (m *Method) addQueryParameters(endpoint string) string {
	hasQueryParameters := len(m.Input.Parameters) != len(m.extensions.EndpointDetails.Parameters)
	if m.extensions.EndpointDetails.Body == "" && hasQueryParameters {
		endpoint += "?"
		queryParameterNames := []string{}
		for _, p := range m.Input.Parameters {
			if !isIn(m.extensions.EndpointDetails.Parameters, p.ProtoName) {
				queryParameterNames = append(queryParameterNames, p.ProtoName)
			}
		}

		for i, name := range queryParameterNames {
			if i > 0 {
				endpoint += "&"
			}

			endpoint += fmt.Sprintf("<%v>", name)
		}
	}

	return endpoint
}

// BodyArgumentType gives the variable type of the body.
func (m *Method) BodyArgumentType() string {
	// If we're dealing with a POST method, probably the entire RPC input message
	// will be used as the body.
	if m.extensions.EndpointDetails.Method == http.MethodPost {
		return m.Input.Name
	}

	if p := m.searchInputParameterByProtoName(m.extensions.EndpointDetails.Body); p != nil {
		return p.RustType()
	}

	if len(m.extensions.EndpointDetails.Parameters) > 0 {
		return m.Input.Name
	}

	return ""
}

// BodyArgumentType gives the variable name of the body.
func (m *Method) BodyArgumentName() string {
	return m.extensions.EndpointDetails.Body
}

func (m *Method) searchInputParameterByProtoName(protoName string) *Parameter {
	for _, p := range m.Input.Parameters {
		if p.ProtoName == protoName {
			return p
		}
	}

	return nil
}

// NeedsInitializeBody returns true or false if the method input struct must be
// built using path and body parameters.
func (m *Method) NeedsInitializeInput() bool {
	return m.extensions.EndpointDetails.Body != "*"
}

func (m *Method) HttpMethod() string {
	method, _ := m.extensions.HttpMethodAndEndpoint()
	return method
}

func (m *Method) IsHttp() bool {
	return m.extensions.GoogleApi != nil
}

func parseMethods(file *protogen.File) ([]*Method, error) {
	// Probably the first service is what we want.
	service := file.Proto.Service[0]

	var methods []*Method
	for _, method := range service.Method {
		extensions := krill.GetMethodExtensions(method)

		inputParameters, err := parseParametersFromMessage(file, method.GetInputType(), extensions)
		if err != nil {
			return nil, err
		}

		methods = append(methods, &Method{
			extensions: extensions,
			Name:       method.GetName(),
			Input: &MethodMessage{
				Name:       filterPackageName(method.GetInputType()),
				Parameters: inputParameters,
			},
			Output: &MethodMessage{
				Name: filterPackageName(method.GetOutputType()),
			},
		})

		// TODO: validate method?
	}

	return methods, nil
}

func filterPackageName(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}
