package templates

import (
	"fmt"
	"strings"

	"github.com/rsfreitas/go-pocket-utils/template"

	"github.com/rsfreitas/protoc-gen-pocket-extensions/internal/openapi"
	"github.com/rsfreitas/protoc-gen-pocket-extensions/internal/proto"
)

// The context availble inside all template files.
type context struct {
	SingleProtobuf    bool
	AppName           string
	Module            string
	GrpcServiceName   string
	PackageName       string
	ProtoFilePath     string
	OutputDir         string
	ProtoIncludePaths []string
	FieldAttributes   []*proto.FieldAttribute
	Methods           []*proto.Method
	Openapi           *openapi.Openapi

	exportOpenapi bool
	exportRust    bool
}

func (c *context) ValidateForExecute() map[string]template.TemplateValidator {
	return map[string]template.TemplateValidator{
		"http.rs": func() bool {
			return c.exportRust && c.IsHttpService()
		},
		"build.rs": func() bool {
			// All protobuf specification files must generated this template.
			return c.exportRust
		},
		"openapi.yaml": func() bool {
			return c.exportOpenapi
		},
	}
}

func (c *context) Extension() string {
	return ""
}

// IsHttpService checks if the context of the current file corresponds to
// a HTTP service. To be a service of this type, the protobuf must include
// at least one endpoint declaration.
func (c *context) IsHttpService() bool {
	for _, m := range c.Methods {
		if m.IsHttp() {
			return true
		}
	}

	return false
}

func buildContext(options *LoadOptions) (*context, error) {
	packageName, err := proto.GetPackageName(options.Plugin)
	if err != nil {
		return nil, err
	}

	protoFilePath, err := proto.GetProtoFilePath(options.Plugin)
	if err != nil {
		return nil, err
	}

	outputDir := options.OutputDir
	if len(outputDir) > 0 {
		outputDir = fmt.Sprintf("%v/%v", outputDir, packageName)
	}

	ctx := &context{
		SingleProtobuf:    options.SingleProtobuf,
		OutputDir:         outputDir,
		PackageName:       packageName,
		ProtoFilePath:     fmt.Sprintf("%v/%v", options.PrototoolPath, protoFilePath),
		ProtoIncludePaths: options.IncludePaths,
		FieldAttributes:   proto.GetFieldAttributes(options.Plugin),
		exportOpenapi:     options.ExportOpenapi,
		exportRust:        options.ExportRust,
	}

	spec, err := proto.Parse(options.Plugin)
	if err != nil {
		return nil, err
	}
	if spec != nil {
		ctx.AppName = spec.AppName
		ctx.Methods = spec.Methods
		ctx.GrpcServiceName = spec.ServiceName
		ctx.Module = filterPackageName(spec.PackageName)
	}

	file, err := proto.GetProtoFile(options.Plugin, true)
	if err != nil {
		return nil, err
	}

	if options.ExportOpenapi {
		opApi, err := openapi.FromProto(file, options.Plugin)
		if err != nil {
			return nil, err
		}
		ctx.Openapi = opApi
	}

	return ctx, nil
}

// filterPackageName retrieves only the last part of a package name.
func filterPackageName(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}
