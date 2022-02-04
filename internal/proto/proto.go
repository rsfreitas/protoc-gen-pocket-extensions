package proto

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-micro-extensions/options/micro"
)

type Spec struct {
	AppName     string
	PackageName string
	ServiceName string
	Methods     []*Method
}

type Method struct {
	Name       string
	Endpoint   string
	Method     string
	InputName  string
	OutputName string
}

type FieldAttribute struct {
	Name      string
	Attribute string
}

func GetPackageName(plugin *protogen.Plugin) (string, error) {
	file, err := getProtoFile(plugin, false)
	if err != nil {
		return "", err
	}

	return string(file.GoPackageName), nil
}

func GetProtoFilePath(plugin *protogen.Plugin) (string, error) {
	file, err := getProtoFile(plugin, false)
	if err != nil {
		return "", err
	}

	return file.Proto.GetName(), nil
}

func GetFieldAttributes(plugin *protogen.Plugin) []*FieldAttribute {
	var fields []*FieldAttribute

	for _, file := range plugin.FilesByPath {
		if !file.Generate {
			// Only deals with file that is going to be processed.
			continue
		}

		for _, msg := range file.Proto.MessageType {
			fields = append(fields, getFieldAttributesFromMessage(file.Proto.GetPackage(), msg)...)
		}
	}

	return fields
}

func getFieldAttributesFromMessage(packageName string, message *descriptor.DescriptorProto) []*FieldAttribute {
	var fields []*FieldAttribute

	for _, field := range message.Field {
		if database := getFieldBuildrsDatabaseExtension(field); database != nil {
			fields = append(fields, &FieldAttribute{
				Name:      fmt.Sprintf(".%v.%v.%v", packageName, message.GetName(), field.GetName()),
				Attribute: fmt.Sprintf(`#[serde(rename(serialize = \"%v\", deserialize = \"%v\"))]`, database.GetName(), database.GetName()),
			})
		}
	}

	return fields
}

func getFieldBuildrsDatabaseExtension(field *descriptor.FieldDescriptorProto) *micro.Database {
	if field != nil && field.Options != nil {
		f := proto.GetExtension(field.Options, micro.E_Database)

		if p, ok := f.(*micro.Database); ok {
			return p
		}
	}

	return nil
}

func Parse(plugin *protogen.Plugin) (*Spec, error) {
	file, err := getProtoFile(plugin, true)
	if err != nil {
		return nil, err
	}

	methods, err := parseMethods(file)
	if err != nil {
		return nil, err
	}

	appName := getMicroFileOptions(file.Proto)
	if appName == "" {
		return nil, errors.New("cannot handle a service without 'micro.micro_app_name' option")
	}

	return &Spec{
		AppName:     appName,
		Methods:     methods,
		ServiceName: file.Proto.Service[0].GetName(),
		PackageName: file.Proto.GetPackage(),
	}, nil
}

func getProtoFile(plugin *protogen.Plugin, withService bool) (*protogen.File, error) {
	if len(plugin.Files) == 0 {
		return nil, errors.New("cannot find the module name without .proto files")
	}

	// The last file in the slice is always the main .proto file that is being
	// "compiled" by protoc.
	file := plugin.Files[len(plugin.Files)-1]
	if !file.Generate {
		return nil, errors.New("proto file not meant to be generated")
	}

	// Proto file has no service declared. Nothing for us to handle here.
	if len(file.Services) == 0 {
		fmt.Println("1")
		return nil, nil
	}

	return file, nil
}

func parseMethods(file *protogen.File) ([]*Method, error) {
	// Probably the first service is what we want.
	service := file.Proto.Service[0]

	var methods []*Method
	for _, method := range service.Method {
		var (
			endpoint   string
			httpMethod string
		)

		api := getGoogleHttpAPIIfAny(method)
		if api != nil {
			httpMethod, endpoint = getMethodAndEndpoint(api)
		}

		methods = append(methods, &Method{
			Name:       method.GetName(),
			InputName:  filterPackageName(method.GetInputType()),
			OutputName: filterPackageName(method.GetOutputType()),
			Endpoint:   endpoint,
			Method:     httpMethod,
		})
	}

	return methods, nil
}

func filterPackageName(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

func getMicroFileOptions(file *descriptor.FileDescriptorProto) string {
	if file.Options != nil {
		n := proto.GetExtension(file.Options, micro.E_MicroAppName)
		if n != nil {
			return n.(string)
		}
	}

	return ""
}

// getGoogleHttpAPIIfAny gets the google.api.http extension of a method if exists.
func getGoogleHttpAPIIfAny(msg *descriptor.MethodDescriptorProto) *annotations.HttpRule {
	if msg.Options != nil {
		h := proto.GetExtension(msg.Options, annotations.E_Http)
		return (h.(*annotations.HttpRule))
	}

	return nil
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
