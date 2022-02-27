package main

import (
	"github.com/MaiBeng/protoc-gen-postman/internal"

	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	p := &internal.Postman{}
	protogen.Options{}.Run(p.Generate)
}
