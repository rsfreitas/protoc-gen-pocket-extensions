package proto

import (
	"net/http"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Parameter struct {
	GoName    string
	ProtoName string
	Location  ParameterLocation

	spec *protogen.Field
}

type ParameterLocation int32

const (
	ParameterLocation_Header ParameterLocation = iota
	ParameterLocation_Path
	ParameterLocation_Query
	ParameterLocation_Body
)

func (p ParameterLocation) String() string {
	switch p {
	case ParameterLocation_Header:
		return "header"

	case ParameterLocation_Path:
		return "path"

	case ParameterLocation_Query:
		return "query"

	case ParameterLocation_Body:
		return "body"
	}

	// This should not happen
	return "unknown"
}

// RustType gives the rust type of the parameter
func (p *Parameter) RustType() string {
	rt := ""

	switch p.spec.Desc.Kind() {
	case protoreflect.BoolKind:
		rt = "bool"

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind:
		rt = "i32"

	case protoreflect.Uint32Kind:
		rt = "u32"

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Fixed64Kind, protoreflect.Sfixed64Kind:
		rt = "i64"

	case protoreflect.Uint64Kind:
		rt = "u64"

	case protoreflect.StringKind:
		rt = "String"

	case protoreflect.FloatKind:
		rt = "f32"

	case protoreflect.DoubleKind:
		rt = "f64"

	case protoreflect.EnumKind:
		desc := p.spec.Desc.Enum().(protoreflect.Descriptor)
		rt = string(desc.Name())

	case protoreflect.MessageKind:
		desc := p.spec.Desc.Message().(protoreflect.Descriptor)
		rt = string(desc.Name())

	case protoreflect.BytesKind:
		// TODO: ???
	}

	return rt
}

func (p *Parameter) BodyInitCall() string {
	call := ""

	switch p.spec.Desc.Kind() {
	case protoreflect.BoolKind:

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind:

	case protoreflect.Uint32Kind:

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Fixed64Kind, protoreflect.Sfixed64Kind:

	case protoreflect.Uint64Kind:

	case protoreflect.StringKind:
		call = ".clone()"

	case protoreflect.FloatKind:

	case protoreflect.DoubleKind:

	case protoreflect.EnumKind:

	case protoreflect.MessageKind:

	case protoreflect.BytesKind:
		// TODO: ???
	}

	return call
}

func parseParametersFromMessage(file *protogen.File, messageName string, extensions *methodExtensions) ([]*Parameter, error) {
	//msg, msgDescriptor, err := searchPackageMessageByName(file, messageName)
	msg, _, err := searchPackageMessageByName(file, messageName)
	if err != nil {
		return nil, err
	}

	var parameters []*Parameter

	for _, field := range msg.Fields {
		desc := field.Desc.(protoreflect.Descriptor)
		protoName := string(desc.Name())
		parameters = append(parameters, &Parameter{
			spec:      field,
			GoName:    field.GoName,
			ProtoName: protoName,
			Location:  getFieldLocation(protoName, extensions),
		})

		// TODO: validate Parameter?
	}

	return parameters, nil
}

func getFieldLocation(name string, extensions *methodExtensions) ParameterLocation {
	var (
		location = ParameterLocation_Body
		found    = isIn(extensions.EndpointDetails.Parameters, name)
	)

	if found {
		location = ParameterLocation_Path
	}
	if !found && extensions.EndpointDetails.Method == http.MethodGet {
		location = ParameterLocation_Query
	}
	if !found && extensions.HasMicroHttpExtension() {
		for _, p := range extensions.Micro.GetHttp().GetHeader() {
			if name == p.GetName() {
				location = ParameterLocation_Header
				break
			}
		}
	}

	return location
}
