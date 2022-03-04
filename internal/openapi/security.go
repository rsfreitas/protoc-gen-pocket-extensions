package openapi

import (
	"fmt"

	"github.com/rsfreitas/protoc-gen-krill-extensions/internal/krill"
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
	schemeType := serviceExtensions.Service.GetSecurityScheme().GetType()

	switch schemeType.String() {
	case "SECURITY_SCHEME_HTTP":
		attrs["type"] = "http"

		httpScheme := httpSchemeToString(serviceExtensions)
		attrs["scheme"] = httpScheme

		if httpScheme == "bearer" {
			attrs["bearerFormat"] = bearerFormatToString(serviceExtensions)
		}

	case "SECURITY_SCHEME_API_KEY":
		attrs["name"] = serviceExtensions.Service.GetSecurityScheme().GetName()
		attrs["in"] = serviceExtensions.Service.GetSecurityScheme().GetIn()

	case "SECURITY_SCHEME_OAUTH2":
	case "SECURITY_SCHEME_OPEN_ID_CONNECT":
	}

	if description := serviceExtensions.Service.GetSecurityScheme().GetDescription(); description != "" {
		attrs["description"] = description
	}

	return attrs
}

func httpSchemeToString(serviceExtensions *krill.ServiceExtensions) string {
	switch serviceExtensions.Service.GetSecurityScheme().GetScheme().String() {
	case "SECURITY_SCHEME_SCHEME_BASIC":
		return "basic"
	case "SECURITY_SCHEME_SCHEME_BEARER":
		return "bearer"
	case "SECURITY_SCHEME_SCHEME_DIGEST":
		return "digest"
	case "SECURITY_SCHEME_SCHEME_OAUTH":
		return "oauth"
	}

	return "unspecified"
}

func bearerFormatToString(serviceExtensions *krill.ServiceExtensions) string {
	switch serviceExtensions.Service.GetSecurityScheme().GetBearerFormat().String() {
	case "SECURITY_SCHEME_BEARER_FORMAT_JWT":
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
