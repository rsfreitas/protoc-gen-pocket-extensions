package krill

//go:generate protoc -I. --go_out=. --go_opt=paths=source_relative krill.proto
//go:generate protoc -I. --go_out=. --go_opt=paths=source_relative krill_http.proto
//go:generate protoc -I. --go_out=. --go_opt=paths=source_relative krill_openapi.proto
