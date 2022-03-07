# protoc-gen-pocket-extensions

**protoc-gen-pocket-extensions** is a protobuf protoc plugin capable of generating
rust source code from protobuf file with the objective of help building services
with [pocket](https://github.com/rsfreitas/pocket) framework.

## Building the plugin

In order to build the plugin, its protobuf annotation extensions must be
compiled into Go sources. This can be achieved by running the following
command:
```bash
go generate ./options/pocket
```

After successfully generated them, the plugin can be built and installed by
using the command:
```bash
go build && go install
```

## Features

## License

Apache 2.0

