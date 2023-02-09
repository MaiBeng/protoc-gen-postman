# protoc-gen-postman

> 一个帮你生成`postman`用例的`protoc`插件

### install Go
```shell
https://go.dev/dl/

# or
https://golang.google.cn/dl/
```

### install protoc
```shell
# version no later than 3.15.7
https://github.com/protocolbuffers/protobuf/releases
```

### install protoc-gen-postman
```shell
go get -u github.com/MaiBeng/protoc-gen-postman
```

### parse proto
```shell
# {{PROTO_OUT_PATH}}: proto output path
# {{PROTO_DEPEND_PATH}}: proto dependency paths, separated by `:`
# {{PROTO_PARSE_PATH}}: proto parsing path, separated by ` ` (space)
protoc --postman_out={{PROTO_OUT_PATH}} --proto_path={{PROTO_DEPEND_PATH}} {{PROTO_PARSE_PATH}}
```

### example
```shell
protoc --postman_out=. --proto_path=$GOPATH/proto:. ./proto/test.proto $GOPATH/proto/*/*.proto
protoc --postman_out=. --proto_path=$GOPATH/proto:. `grep package -rl ./proto`
```

### something error
```shell
error:
protoc-gen-go: unable to determine Go import path for ...

fix:
protoc --postman_out={{PROTO_OUT_PATH}} --postman_opt=M{{PROTO_PARSE_PATH}}=./ --proto_path={{PROTO_DEPEND_PATH}} {{PROTO_PARSE_PATH}}

example:
protoc --postman_out=. --postman_opt=Mproto/test.proto=./ --proto_path=$GOPATH/proto:. ./proto/test.proto

protoc --proto_path=src \
  --go_opt=Mprotos/test1.proto=. \
  --go_opt=Mprotos/test2.proto=. \
  ./proto/*.proto
```

> The file `source.postman_collection.json` will be generated in the current folder.
> Then we can import it into `Postman` and rename your collection.

![image](https://github.com/MaiBeng/protoc-gen-postman/blob/main/import.gif)