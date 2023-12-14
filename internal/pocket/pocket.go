package pocket

import (
	"regexp"
	"strings"

	"github.com/juliangruber/go-intersect"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	pocketpb "github.com/rsfreitas/protoc-gen-pocket-extensions/options/pocket"
)

type FileExtensions struct {
	AppName        string
	OpenapiTitle   string
	OpenapiVersion string
	Servers        []*pocketpb.OpenapiServer
}

type ServiceExtensions struct {
	Service *pocketpb.HttpService
}

type MethodExtensions struct {
	GoogleApi       *annotations.HttpRule
	Method          *pocketpb.HttpMethod
	OpenapiMethod   *pocketpb.OpenapiMethod
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
	OpenapiMessage *pocketpb.OpenapiMessage
}

type FieldExtensions struct {
	Database *pocketpb.Database
	Openapi  *pocketpb.Property
	Http     *pocketpb.HttpFieldProperty
}

func (e *MethodExtensions) HasKrillHttpExtension() bool {
	return e.Method != nil
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

func (e *MethodExtensions) GetHeaderMemberNames() map[string]string {
	if !e.HasKrillHttpExtension() || len(e.Method.GetHeader()) == 0 {
		return nil
	}

	names := make(map[string]string)

	for _, p := range e.Method.GetHeader() {
		names[p.GetMemberName()] = p.GetName()
	}

	return names
}

func (f *FieldExtensions) PropertyLocation() pocketpb.HttpFieldLocation {
	if f.Http != nil {
		return f.Http.GetLocation()
	}

	return pocketpb.HttpFieldLocation_HTTP_FIELD_LOCATION_BODY
}

func (s *ServiceExtensions) GetHeaderMemberNames() map[string]string {
	if s.Service == nil || len(s.Service.GetHeader()) == 0 {
		return nil
	}

	names := make(map[string]string)

	for _, p := range s.Service.GetHeader() {
		names[p.GetMemberName()] = p.GetName()
	}

	return names
}

func GetServiceExtensions(service *descriptor.ServiceDescriptorProto) *ServiceExtensions {
	if service.Options != nil {
		s := proto.GetExtension(service.Options, pocketpb.E_ServiceDefinitions)

		if svc, ok := s.(*pocketpb.HttpService); ok {
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

func getKrillOpenapiMethodExtension(method *descriptor.MethodDescriptorProto) *pocketpb.OpenapiMethod {
	if method.Options != nil {
		m := proto.GetExtension(method.Options, pocketpb.E_Operation)
		return m.(*pocketpb.OpenapiMethod)
	}

	return nil
}

func getKrillMethodExtension(method *descriptor.MethodDescriptorProto) *pocketpb.HttpMethod {
	if method.Options != nil {
		m := proto.GetExtension(method.Options, pocketpb.E_MethodDefinitions)
		return m.(*pocketpb.HttpMethod)
	}

	return nil
}

// getGoogleHttpAPIIfAny gets the google.api.http extension of a method if exists.
func getGoogleHttpAPIIfAny(msg *descriptor.MethodDescriptorProto) *annotations.HttpRule {
	if msg.Options != nil {
		h := proto.GetExtension(msg.Options, annotations.E_Http)
		return h.(*annotations.HttpRule)
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

func getKrillOpenapiMessageExtension(message *descriptor.DescriptorProto) *pocketpb.OpenapiMessage {
	if message.Options != nil {
		m := proto.GetExtension(message.Options, pocketpb.E_Message)
		return m.(*pocketpb.OpenapiMessage)
	}

	return nil
}

func GetFieldExtensions(field *descriptor.FieldDescriptorProto) *FieldExtensions {
	ext := &FieldExtensions{}

	if field != nil && field.Options != nil {
		f := proto.GetExtension(field.Options, pocketpb.E_Property)
		if p, ok := f.(*pocketpb.Property); ok {
			ext.Openapi = p
		}

		d := proto.GetExtension(field.Options, pocketpb.E_Database)
		if p, ok := d.(*pocketpb.Database); ok {
			ext.Database = p
		}

		h := proto.GetExtension(field.Options, pocketpb.E_FieldDefinitions)
		if d, ok := h.(*pocketpb.HttpFieldProperty); ok {
			ext.Http = d
		}
	}

	return ext
}

func GetFileExtensions(file *descriptor.FileDescriptorProto) *FileExtensions {
	var (
		name    string
		title   string
		version string
		servers []*pocketpb.OpenapiServer
	)

	if file.Options != nil {
		if n := proto.GetExtension(file.Options, pocketpb.E_AppName); n != nil {
			name = n.(string)
		}

		if n := proto.GetExtension(file.Options, pocketpb.E_Title); n != nil {
			title = n.(string)
		}

		if n := proto.GetExtension(file.Options, pocketpb.E_Version); n != nil {
			version = n.(string)
		}

		if s := proto.GetExtension(file.Options, pocketpb.E_Server); s != nil {
			servers = s.([]*pocketpb.OpenapiServer)
		}
	}

	return &FileExtensions{
		AppName:        name,
		OpenapiTitle:   title,
		OpenapiVersion: version,
		Servers:        servers,
	}
}

func ResponseCodeToHttpCode(code pocketpb.ResponseCode) string {
	switch code {
	case pocketpb.ResponseCode_RESPONSE_CODE_OK:
		return "200"
	case pocketpb.ResponseCode_RESPONSE_CODE_CREATED:
		return "201"
	case pocketpb.ResponseCode_RESPONSE_CODE_NOT_FOUND:
		return "404"
	case pocketpb.ResponseCode_RESPONSE_CODE_BAD_REQUEST:
		return "400"
	case pocketpb.ResponseCode_RESPONSE_CODE_UNAUTHORIZED:
		return "401"
	case pocketpb.ResponseCode_RESPONSE_CODE_PRECONDITION_FAILED:
		return "412"
	}

	// Internal Error
	return "500"
}

func PropertyFormatTrimPrefix(format pocketpb.PropertyFormat) string {
	prefix := enumStringsIntersection(
		pocketpb.PropertyFormat_PROPERTY_FORMAT_UNSPECIFIED.String(),
		pocketpb.PropertyFormat_PROPERTY_FORMAT_STRING.String(),
	)

	return strings.TrimPrefix(format.String(), prefix)
}

func enumStringsIntersection(s1, s2 string) string {
	p1 := strings.Split(s1, "_")
	p2 := strings.Split(s2, "_")

	i := intersect.Simple(p1, p2)
	if len(i) == 0 {
		return ""
	}

	var parts []string
	for _, s := range i {
		parts = append(parts, s.(string))
	}

	return strings.Join(parts, "_") + "_"
}
