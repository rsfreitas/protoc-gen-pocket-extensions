#!/bin/bash

INCLUDE="-I ."

(cd options/krill && protoc $INCLUDE --go_out=. --go_opt=paths=source_relative krill.proto)

exit 0

