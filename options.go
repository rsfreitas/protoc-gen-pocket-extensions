package main

import (
	"flag"
	"strings"
)

type pluginOptions struct {
	exportOpenapi           *bool
	exportRust              *bool
	axumFramework           *bool
	rocketFramework         *bool
	singleProtobuf          *bool
	includePaths            *string
	outputDir               *string
	prototoolRootPath       *string
	openapiSettingsFilename *string
	flags                   flag.FlagSet
}

func (p *pluginOptions) FlagsSet() func(string, string) error {
	return p.flags.Set
}

func (p *pluginOptions) IncludePaths() []string {
	if p.includePaths == nil {
		return []string{}
	}

	return strings.Split(*p.includePaths, ";")
}

func (p *pluginOptions) Axum() bool {
	return *p.axumFramework
}

func (p *pluginOptions) Rocket() bool {
	return *p.rocketFramework
}

func (p *pluginOptions) SingleProtobuf() bool {
	return *p.singleProtobuf
}

func (p *pluginOptions) PrototoolPath() string {
	return *p.prototoolRootPath
}

func (p *pluginOptions) OutputDir() string {
	return *p.outputDir
}

func (p *pluginOptions) ExportOpenapi() bool {
	return *p.exportOpenapi
}

func (p *pluginOptions) ExportRust() bool {
	return *p.exportRust
}

func (p *pluginOptions) OpenapiSettings() string {
	return *p.openapiSettingsFilename
}

func NewPluginOptions() *pluginOptions {
	o := &pluginOptions{}

	o.axumFramework = o.flags.Bool("axum", false, "Enables/Disables axum framework.")
	o.rocketFramework = o.flags.Bool("rocket", false, "Enables/Disables rocket framework.")
	o.singleProtobuf = o.flags.Bool("single_protobuf", false, "Enables/Disables main function inside the template.")
	o.outputDir = o.flags.String("output_dir", "", "Sets the generated output directory for rust generated files.")
	o.prototoolRootPath = o.flags.String("prototool_path", "", "Sets the root path used by prototool to search for protobuf files.")
	o.includePaths = o.flags.String("include_paths", "", "Sets relative paths as include directories when compiling.")
	o.exportOpenapi = o.flags.Bool("openapi", false, "Enables/Disables openapi generation.")
	o.exportRust = o.flags.Bool("rust", false, "Enables/Disables rust source code generation.")
	o.openapiSettingsFilename = o.flags.String("openapi_settings", "", "Sets the OpenAPI additional settings file.")

	return o
}
