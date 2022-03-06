# protoc-gen-krill-extensions

**protoc-gen-krill-extensions** is a protobuf protoc plugin capable of generating
rust source code from protobuf file with the objective of help building services
with [krill](https://github.com/rsfreitas/krill) framework.

## Building the plugin

To build the plugin it's required to generate Go sources from the plugin's
protobuf annotation extensions. This can be achieved by running the following
command:
```bash
go generate ./options/krill
```

After successfully generated them, the plugin can be built and installed by
using the command:
```bash
go build && go install
```

## Features

## License

Apache 2.0

