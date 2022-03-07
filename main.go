package main

import (
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/rsfreitas/protoc-gen-pocket-extensions/internal/templates"
)

func main() {
	options := NewPluginOptions()

	protogen.Options{
		ParamFunc: options.FlagsSet(),
	}.Run(func(plugin *protogen.Plugin) error {
		tpl, err := templates.Load(&templates.LoadOptions{
			Plugin:          plugin,
			SingleProtobuf:  options.SingleProtobuf(),
			OutputDir:       options.OutputDir(),
			PrototoolPath:   options.PrototoolPath(),
			IncludePaths:    options.IncludePaths(),
			UseRocket:       options.Rocket(),
			ExportOpenapi:   options.ExportOpenapi(),
			ExportRust:      options.ExportRust(),
			OpenapiSettings: options.OpenapiSettings(),
		})
		if err != nil {
			return err
		}
		if tpl == nil {
			return nil
		}

		gen, err := tpl.Execute()
		if err != nil {
			return err
		}

		for _, template := range gen {
			f := plugin.NewGeneratedFile(template.Filename, ".")
			if _, err := f.Write(template.Data.Bytes()); err != nil {
				return err
			}
		}

		return nil
	})
}
