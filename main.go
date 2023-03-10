package main

import (
	"fmt"
	"path/filepath"

	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/rsfreitas/protoc-gen-pocket-extensions/internal/proto"
	"github.com/rsfreitas/protoc-gen-pocket-extensions/internal/templates"
)

func main() {
	options := newPluginOptions()

	protogen.Options{
		ParamFunc: options.FlagsSet(),
	}.Run(func(plugin *protogen.Plugin) error {
		plugin.SupportedFeatures = gengo.SupportedFeatures
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
			return fmt.Errorf("%v: %w", plugin.Request.FileToGenerate[len(plugin.Request.FileToGenerate)-1], err)
		}
		if tpl == nil {
			return nil
		}

		gen, err := tpl.Execute()
		if err != nil {
			return fmt.Errorf("%v: %w", plugin.Request.FileToGenerate[len(plugin.Request.FileToGenerate)-1], err)
		}

		for _, template := range gen {
			filePath, _ := proto.GetProtoFilePath(plugin)
			filename := filepath.Join(
				filepath.Dir(filePath),
				filepath.Base(template.Filename),
			)

			f := plugin.NewGeneratedFile(filename, ".")
			if _, err := f.Write(template.Data.Bytes()); err != nil {
				return err
			}
		}

		return nil
	})
}
