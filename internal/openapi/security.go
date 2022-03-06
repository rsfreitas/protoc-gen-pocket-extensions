package openapi

import (
	"fmt"

	"github.com/rsfreitas/protoc-gen-krill-extensions/internal/krill"
	krillpb "github.com/rsfreitas/protoc-gen-krill-extensions/options/krill"
)

func buildSecuritySchemeFromServiceExtensions(serviceExtensions *krill.ServiceExtensions, tabSize int) string {
	s := ""
	attrs := securitySchemeAttrs(serviceExtensions)
	padding := buildPadding(tabSize)

	for k, v := range attrs {
		s += fmt.Sprintf("%v%s: %s\n", padding, k, v)
	}

	return s
}

func securitySchemeAttrs(serviceExtensions *krill.ServiceExtensions) map[string]string {
	attrs := make(map[string]string)

	switch serviceExtensions.Service.GetSecurityScheme().GetType() {
	case krillpb.HttpSecuritySchemeType_HTTP_SECURITY_SCHEME_HTTP:
		attrs["type"] = "http"

		httpScheme := httpSchemeToString(serviceExtensions)
		attrs["scheme"] = httpScheme

		if httpScheme == "bearer" {
			attrs["bearerFormat"] = bearerFormatToString(serviceExtensions)
		}

	case krillpb.HttpSecuritySchemeType_HTTP_SECURITY_SCHEME_API_KEY:
		attrs["name"] = serviceExtensions.Service.GetSecurityScheme().GetName()
		attrs["in"] = serviceExtensions.Service.GetSecurityScheme().GetIn()
	}

	if description := serviceExtensions.Service.GetSecurityScheme().GetDescription(); description != "" {
		attrs["description"] = description
	}

	return attrs
}

func httpSchemeToString(serviceExtensions *krill.ServiceExtensions) string {
	switch serviceExtensions.Service.GetSecurityScheme().GetScheme() {
	case krillpb.HttpSecuritySchemeScheme_HTTP_SECURITY_SCHEME_SCHEME_BASIC:
		return "basic"
	case krillpb.HttpSecuritySchemeScheme_HTTP_SECURITY_SCHEME_SCHEME_BEARER:
		return "bearer"
	case krillpb.HttpSecuritySchemeScheme_HTTP_SECURITY_SCHEME_SCHEME_DIGEST:
		return "digest"
	case krillpb.HttpSecuritySchemeScheme_HTTP_SECURITY_SCHEME_SCHEME_OAUTH:
		return "oauth"
	}

	return "unspecified"
}

func bearerFormatToString(serviceExtensions *krill.ServiceExtensions) string {
	switch serviceExtensions.Service.GetSecurityScheme().GetBearerFormat() {
	case krillpb.HttpSecuritySchemeBearerFormat_HTTP_SECURITY_SCHEME_BEARER_FORMAT_JWT:
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
