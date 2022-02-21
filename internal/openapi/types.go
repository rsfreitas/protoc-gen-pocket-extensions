package openapi

type SchemaType int

const (
	SchemaType_Unspecified SchemaType = iota
	SchemaType_Object
	SchemaType_String
	SchemaType_Array
	SchemaType_Bool
	SchemaType_Integer
	SchemaType_Number
)

func (s SchemaType) String() string {
	switch s {
	case SchemaType_Integer:
		return "integer"

	case SchemaType_Number:
		return "number"

	case SchemaType_Bool:
		return "boolean"

	case SchemaType_Object:
		return "object"

	case SchemaType_String:
		return "string"

	case SchemaType_Array:
		return "array"
	}

	return "unspecified"
}
