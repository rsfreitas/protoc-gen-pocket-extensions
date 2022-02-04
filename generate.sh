#!/bin/bash

INCLUDE="-I ."

(cd options/micro && protoc $INCLUDE --go_out=. --go_opt=paths=source_relative micro.proto)

exit 0

