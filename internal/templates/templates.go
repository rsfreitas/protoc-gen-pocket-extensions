package templates

import (
	"embed"

	//	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/rsfreitas/go-pocket-utils/template"
)

//go:embed *.tmpl
var files embed.FS

// TODO: need to validate here for mandatory options
type LoadOptions struct {
	SingleProtobuf  bool
	UseRocket       bool
	ExportOpenapi   bool
	ExportRust      bool
	OpenapiSettings string
	OutputDir       string
	PrototoolPath   string
	IncludePaths    []string
	Plugin          *protogen.Plugin
}

func Load(options *LoadOptions) (*template.Templates, error) {
	ctx, err := buildContext(options)
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		return nil, nil
	}

	return template.LoadTemplates(&template.Options{
		Plugin:  options.Plugin,
		Files:   files,
		Context: ctx,
	})
}
