package openapi

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/rsfreitas/protoc-gen-krill-extensions/internal/krill"
)

type Openapi struct {
	Info              *Info
	PathItems         map[string]map[string]*Operation `yaml:"paths"`
	Components        *Components
	ServiceExtensions *krill.ServiceExtensions
}

type Info struct {
	Title   string
	Version string
	NoAuth  bool
}

type Components struct {
	Schemas   map[string]*Schema
	Responses map[string]*Response
}

func (o *Openapi) HasAuth() bool {
	// TODO: use extension to set this info
	return !o.Info.NoAuth
}

func FromProto(file *protogen.File, plugin *protogen.Plugin) (*Openapi, error) {
	var (
		enums      = parseEnums(plugin)
		extensions = krill.GetServiceExtensions(file.Proto.Service[0])
	)

	fmt.Println(enums)

	return &Openapi{
		ServiceExtensions: extensions,
	}, nil
}
