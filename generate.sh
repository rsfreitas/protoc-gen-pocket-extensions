#!/bin/bash

INCLUDE="-I ."

(cd options/krill && protoc $INCLUDE --go_out=. --go_opt=paths=source_relative krill.proto)
(cd options/krill && protoc $INCLUDE --go_out=. --go_opt=paths=source_relative krill_http.proto)
(cd options/krill && protoc $INCLUDE --go_out=. --go_opt=paths=source_relative krill_openapi.proto)

exit 0

