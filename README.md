# protoc-gen-postman

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
# {{PROTO_DEPEND_PATH}}: proto dependency paths, separated by `:`
# {{PROTO_PARSE_PATH}}: proto parsing path, separated by ` ` (space)
protoc --postman_out=. --proto_path={{PROTO_DEPEND_PATH}} {{PROTO_PARSE_PATH}}
```

### example
```shell
protoc --postman_out=. --proto_path=$GOPATH/proto:. ./proto/test.proto $GOPATH/proto/*/*.proto
protoc --postman_out=. --proto_path=$GOPATH/proto:. `grep package -rl ./proto`
```

> The file `source.postman_collection.json` will be generated in the current folder.
> Then we can import it into `Postman` and rename your collection.

![image](https://github.com/MaiBeng/protoc-gen-postman/blob/main/import.gif)