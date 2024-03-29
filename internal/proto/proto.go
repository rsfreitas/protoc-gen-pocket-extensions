package proto

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-pocket-extensions/internal/pocket"
)

type Spec struct {
	AppName     string
	PackageName string
	ServiceName string
	Methods     []*Method
}

type FieldAttribute struct {
	Name      string
	Attribute string
}

func GetPackageName(plugin *protogen.Plugin) (string, error) {
	file, err := GetProtoFile(plugin)
	if err != nil {
		return "", err
	}
	if file == nil {
		return "", nil
	}

	return string(file.GoPackageName), nil
}

func GetProtoFilePath(plugin *protogen.Plugin) (string, error) {
	file, err := GetProtoFile(plugin)
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
		extensions := pocket.GetFieldExtensions(field)
		if extensions.Database != nil {
			fields = append(fields, &FieldAttribute{
				Name:      fmt.Sprintf(".%v.%v.%v", packageName, message.GetName(), field.GetName()),
				Attribute: fmt.Sprintf(`#[serde(rename(serialize = \"%v\", deserialize = \"%v\"))]`, extensions.Database.GetName(), extensions.Database.GetName()),
			})
		}
	}

	return fields
}

func Parse(plugin *protogen.Plugin) (*Spec, error) {
	file, err := GetProtoFile(plugin)
	if err != nil {
		return nil, err
	}

	methods, err := parseMethods(file)
	if err != nil {
		return nil, err
	}

	extensions := pocket.GetFileExtensions(file.Proto)
	if extensions.AppName == "" {
		return nil, errors.New("cannot handle a service without 'pocket.pocket_app_name' option")
	}

	return &Spec{
		AppName:     extensions.AppName,
		Methods:     methods,
		ServiceName: file.Proto.Service[0].GetName(),
		PackageName: file.Proto.GetPackage(),
	}, nil
}

func GetProtoFile(plugin *protogen.Plugin) (*protogen.File, error) {
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
		return nil, nil
	}

	return file, nil
}

// searchMessageByName searches for a protobuf message by its name. It returns
// both the protogen.Message and a descriptor.DescriptorProto if found.
func searchPackageMessageByName(file *protogen.File, fullyQualifiedName string) (msg *protogen.Message, msgDescriptor *descriptor.DescriptorProto, err error) {
	packageName, msgName := splitPackageAndMessage(fullyQualifiedName)

	// Assures that we only search for messages that belongs to the same
	// package.
	if packageName != string(file.Desc.FullName()) {
		return nil, nil, fmt.Errorf("message '%s' does not belong to the current package", fullyQualifiedName)
	}

	for _, message := range file.Messages {
		if message.GoIdent.GoName == msgName {
			msg = message
			break
		}
	}

	if msg != nil {
		for _, message := range file.Proto.MessageType {
			if message.GetName() == msgName {
				msgDescriptor = message
				break
			}
		}
	}

	if msg == nil {
		err = fmt.Errorf("could not find message with name '%s' inside the package", msgName)
	}

	return
}

// splitPackageAndMessage splits a protobuf message name into its package name
// and its proper name.
func splitPackageAndMessage(messageName string) (string, string) {
	parts := strings.Split(messageName, ".")
	return strings.Join(parts[1:len(parts)-1], "."), parts[len(parts)-1]
}

func isIn(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}

	return false
}
