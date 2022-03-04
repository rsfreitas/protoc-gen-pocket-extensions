package openapi

import (
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/compiler/protogen"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-krill-extensions/internal/krill"
)

func trimPackagePath(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
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
