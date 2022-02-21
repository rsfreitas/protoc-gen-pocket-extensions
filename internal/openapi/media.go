package openapi

type Media struct {
	Schema *Schema
}

func (m *Media) String(prefixSpacing int) string {
	return m.Schema.String(prefixSpacing)
}

func NewMedia(schema *Schema) *Media {
	return &Media{
		Schema: schema,
	}
}
