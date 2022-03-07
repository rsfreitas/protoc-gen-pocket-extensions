package openapi

import (
	"fmt"

	"github.com/rsfreitas/protoc-gen-pocket-extensions/internal/pocket"
	pocketpb "github.com/rsfreitas/protoc-gen-pocket-extensions/options/pocket"
)

func buildSecuritySchemeFromServiceExtensions(serviceExtensions *pocket.ServiceExtensions, tabSize int) string {
	s := ""
	attrs := securitySchemeAttrs(serviceExtensions)
	padding := buildPadding(tabSize)

	for k, v := range attrs {
		s += fmt.Sprintf("%v%s: %s\n", padding, k, v)
	}

	return s
}

func securitySchemeAttrs(serviceExtensions *pocket.ServiceExtensions) map[string]string {
	attrs := make(map[string]string)

	switch serviceExtensions.Service.GetSecurityScheme().GetType() {
	case pocketpb.HttpSecuritySchemeType_HTTP_SECURITY_SCHEME_HTTP:
		attrs["type"] = "http"

		httpScheme := httpSchemeToString(serviceExtensions)
		attrs["scheme"] = httpScheme

		if httpScheme == "bearer" {
			attrs["bearerFormat"] = bearerFormatToString(serviceExtensions)
		}

	case pocketpb.HttpSecuritySchemeType_HTTP_SECURITY_SCHEME_API_KEY:
		attrs["name"] = serviceExtensions.Service.GetSecurityScheme().GetName()
		attrs["in"] = serviceExtensions.Service.GetSecurityScheme().GetIn()
	}

	if description := serviceExtensions.Service.GetSecurityScheme().GetDescription(); description != "" {
		attrs["description"] = description
	}

	return attrs
}

func httpSchemeToString(serviceExtensions *pocket.ServiceExtensions) string {
	switch serviceExtensions.Service.GetSecurityScheme().GetScheme() {
	case pocketpb.HttpSecuritySchemeScheme_HTTP_SECURITY_SCHEME_SCHEME_BASIC:
		return "basic"
	case pocketpb.HttpSecuritySchemeScheme_HTTP_SECURITY_SCHEME_SCHEME_BEARER:
		return "bearer"
	case pocketpb.HttpSecuritySchemeScheme_HTTP_SECURITY_SCHEME_SCHEME_DIGEST:
		return "digest"
	case pocketpb.HttpSecuritySchemeScheme_HTTP_SECURITY_SCHEME_SCHEME_OAUTH:
		return "oauth"
	}

	return "unspecified"
}

func bearerFormatToString(serviceExtensions *pocket.ServiceExtensions) string {
	switch serviceExtensions.Service.GetSecurityScheme().GetBearerFormat() {
	case pocketpb.HttpSecuritySchemeBearerFormat_HTTP_SECURITY_SCHEME_BEARER_FORMAT_JWT:
		return "jwt"
	}

	return "unspecified"
}

func buildPadding(tabSize int) string {
	s := ""
	for i := 0; i < tabSize; i++ {
		s += " "
	}

	return s
}
