package templates

import (
	"embed"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/rsfreitas/go-micro-utils/template"
)

//go:embed *.tmpl
var files embed.FS

type LoadOptions struct {
	SingleProtobuf bool
	UseRocket      bool
	OutputDir      string
	PrototoolPath  string
	IncludePaths   []string
	Plugin         *protogen.Plugin
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
