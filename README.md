# protoc-gen-pocket-extensions

**protoc-gen-pocket-extensions** is a protobuf protoc plugin which allows
generating  source code extensions for services built with [pocket](https://github.com/rsfreitas/pocket)
framework. It also provides an option to generate API documenation for
services using special proto annotations.

## Building the plugin

In order to build the plugin, its protobuf annotation extensions must be
compiled into Go sources. This can be achieved by running the following
command:
```bash
go generate ./options/pocket
```

After successfully generated them, the plugin can be built and installed,
locally, by using the command:
```bash
go build && go install
```

## Features

* Extended generated source code for _pocket_ services, using proto annotations;
* Generates [OpenAPI 3.0.3](https://swagger.io/specification/v3/) spec files (YAML format).

## Using plugin annotations for pocket services

In order to use annotations to extend a pocket service, the file **pocket.proto**
must be imported inside the desired protobuf spec file and the service must be
named using one of its custom annotations. Like the example:
```!protobuf
syntax = "proto3";

package service.example.v1;

import "protoc-gen-pocket-extensions/options/pocket/pocket.proto";

// Defines the service name
option (pocket.service.app_name) = "example";

service ExampleService {
    rpc ...
}
```

## Using plugin annotations to generate OpenAPI spec

The pocket framework handles creating an HTTP server service using a protobuf
spec file by declaring all its endpoints (and which internal functions (RPC) will
handle them) and their API contracts using proto annotations.

Example:
```!protobuf
syntax = "proto3";

package service.example_resource.v1;

import "protoc-gen-pocket-extensions/options/pocket/pocket.proto";
import "protoc-gen-pocket-extensions/options/pocket/pocket_openapi.proto";
import "protoc-gen-pocket-extensions/options/pocket/pocket_http.proto";

// Sets the service name.
option (pocket.service.app_name) = "example-resource";

// Sets main OpenAPI information
option (pocket.openapi.title) = "example-resource";
option (pocket.openapi.version) = "0.1.0";

service ExampleService {
  // Sets the service authentication mode.
  option (pocket.http.service_definitions) = {
    security_scheme: {
      type: HTTP_SECURITY_SCHEME_HTTP
      scheme: HTTP_SECURITY_SCHEME_SCHEME_BEARER
    }
  };
  
  rpc GetExample(GetExampleRequest) returns (GetExampleResponse) {
    // Sets a GET endpoint for the service
    option (google.api.http) = {
      get: "/example-resource/v1/examples"
    };

    // Sets all possible status code returned by the endpoint
    option (pocket.openapi.operation) = {
      summary: "A brief summary of the endpoint."
      description: "Some more details of what the endpoint does."
      response: {
        code: RESPONSE_CODE_OK
        description: "Success."
      }

      response: {
        code: RESPONSE_CODE_BAD_REQUEST
        description: "A bad request."
      }

      response: {
        code: RESPONSE_CODE_UNAUTHORIZED
        description: "An unauthorized request."
      }

      response: {
        code: RESPONSE_CODE_INTERNAL_ERROR
        description: "An internal error occurred."
      }
    };
  };
}

message GetExampleRequest {
...
}

message GetExampleResponse {
...
}
```

## License

Apache 2.0
