//go:generate protoc -I. --go_out=. --go_opt=paths=source_relative pocket.proto
//go:generate protoc -I. --go_out=. --go_opt=paths=source_relative pocket_http.proto
//go:generate protoc -I. --go_out=. --go_opt=paths=source_relative pocket_openapi.proto
package pocket
