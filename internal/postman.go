package internal

import (
	"encoding/json"
	"fmt"
	"github.com/MaiBeng/protoc-gen-postman/google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/reflect/protoreflect"
	"reflect"
	"strings"
	"time"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// protoc --go_out=. ./proto/annotations.proto ./proto/http.proto

const (
	FILENAME           = "./source.postman_collection.json"
	SCHEMA             = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
	COMMENTS_HEADER    = "@reqMetadata"
	GRPC_HEADER_PREFIX = "Grpc-Metadata-"
)

type Postman struct{}

type Info struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
}

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type Raw struct {
	Language string `json:"language"` // 固定值 json
}
type Options struct {
	Raw *Raw `json:"raw"`
}
type Body struct {
	Mode    string   `json:"mode"` // 固定值 raw
	Raw     string   `json:"raw"`
	Options *Options `json:"options"`
}

type Query struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type URL struct {
	Raw   string   `json:"raw"`
	Host  []string `json:"host"`
	Path  []string `json:"path"`
	Query []*Query `json:"query"`
}

type Request struct {
	Method string    `json:"method"`
	Header []*Header `json:"header"`
	Body   *Body     `json:"body"`
	URL    *URL      `json:"url"`
}

type Item struct {
	Name    string   `json:"name"`
	Request *Request `json:"request"` // empty when is folder
	Item    []*Item  `json:"item"`
}

type PostmanGenerated struct {
	Info *Info   `json:"info"`
	Item []*Item `json:"item,omitempty"`
}

func (p Postman) Generate(plugin *protogen.Plugin) error {
	if len(plugin.Files) < 1 {
		return nil
	}

	// 指定生成文件的文件名
	version := time.Now().Format("20060102030405")

	// 创建一个文件生成器对象
	g := plugin.NewGeneratedFile(FILENAME, plugin.Files[len(plugin.Files)-1].GoImportPath)

	var out = PostmanGenerated{
		Info: &Info{
			Name:   "version." + version,
			Schema: SCHEMA,
		},
		Item: nil,
	}

	// 通过plugin.Fiels，我们可以拿到所有的输入的proto文件
	// 如果我们需要对这个文件生成代码的话，那么就进入到generateFile()逻辑

	// 合并同 packageName 的service
	var packageFile = make(map[string][]*protogen.File)
	for _, file := range plugin.Files {
		if file.Generate {
			packageFile[string(file.GoPackageName)] = append(packageFile[string(file.GoPackageName)], file)
		}
	}
	for name, files := range packageFile {
		item, err := p.GetFilesItem(name, files)
		if err != nil {
			return err
		}

		out.Item = append(out.Item, item)
	}

	// 调用g.P就是往文件开始写入自己期待的代码
	outStr, err := json.Marshal(out)
	if err != nil {
		return err
	}
	g.P(fmt.Sprintf("%s", outStr))

	return nil
}

func (p Postman) GetFilesItem(name string, files []*protogen.File) (*Item, error) {
	var fileItem = &Item{
		Name: name,
	}

	// Traverse Services
	for _, file := range files {
		for _, service := range file.Services {
			serviceItem, err := p.GetServiceItem(service)
			if err != nil {
				return nil, err
			}

			fileItem.Item = append(fileItem.Item, serviceItem)
		}
	}

	return fileItem, nil

}

func (p Postman) GetServiceItem(service *protogen.Service) (*Item, error) {
	var serviceItem = &Item{
		Name: service.GoName,
	}

	// Traverse Methods
	for _, method := range service.Methods {
		methodItem, err := p.GetMethodItem(method)
		if err != nil {
			return nil, err
		}

		serviceItem.Item = append(serviceItem.Item, methodItem)
	}

	return serviceItem, nil
}

func (p Postman) GetMethodItem(method *protogen.Method) (*Item, error) {
	// 因为我们通过method.Desc.Options() 拿到的数据类型是`interface{}` 类型
	// 所以这里我们需要对Options，明确指定转换为 *descriptorpb.MethodOptions 类型
	// 这样子就能拿到我们的MethodOption对象
	options, ok := method.Desc.Options().(*descriptorpb.MethodOptions)
	if !ok {
		return nil, fmt.Errorf("method.Desc.Options err")
	}

	// PS：重点
	// 这里我们看到我们借助了一个非protogen下的包的内容
	// 原因就是，protobuf编译器会把自定义的Option全部指定为Extension，由于并非内置的属性和值
	// protobuf官方是没办法拿到和你对应的可读的内容的，只能通过拿到经过序列化之后的数据。
	// 因此，我们这里通过 proto.GetExtension的方法，把刚才annotations.proto单独编译好的 annotations.pb.proto 文件下的 annotations.E_HTTP 加载进来，
	// 指定了我需要在自定义扩展的MethodOptions中，拿到该Http下里面的value
	// 也因此，我们可以再经过一次类型转换，就可以拿到了具体的httpRule
	httpRule, ok := proto.GetExtension(options, annotations.E_Http).(*annotations.HttpRule)
	if !ok {
		return nil, fmt.Errorf("proto.GetExtension err")
	}

	var requestMethod, urlHost string
	if url := httpRule.GetPost(); url != "" {
		requestMethod = "POST"
		urlHost = url
	} else if url = httpRule.GetGet(); url != "" {
		requestMethod = "GET"
		urlHost = url
	} else {
		requestMethod = "POST"
		urlHost = url
		//return nil, fmt.Errorf("unknown method.httpRule")
	}

	// 解析 request
	inputMap := p.transField(method.Input, 3)

	// 解析注释
	desc, header := p.getMethodDescAndHeader(method.Comments.Leading)

	var methodItem = &Item{}
	if requestMethod == "POST" {
		raw, err := json.MarshalIndent(inputMap, "", "    ")
		if err != nil {
			return nil, err
		}

		methodItem = &Item{
			Name: method.GoName + "(" + desc + ")",
			Request: &Request{
				Method: requestMethod,
				Header: header,
				Body: &Body{
					Mode:    "raw",
					Raw:     string(raw),
					Options: &Options{Raw: &Raw{Language: "json"}},
				},
				URL: &URL{
					Raw:  "{{domain}}" + urlHost,
					Host: []string{"{{domain}}"},
					Path: strings.Split(strings.TrimPrefix(urlHost, "/"), "/"),
				},
			},
		}
	}

	if requestMethod == "GET" {
		var rawParams string
		var querys = p.transParmas(inputMap)
		for _, query := range querys {
			rawParams += query.Key + "=" + query.Value + "&"
		}
		rawParams = "?" + strings.TrimSuffix(rawParams, "&")

		methodItem = &Item{
			Name: method.GoName + "(" + desc + ")",
			Request: &Request{
				Method: requestMethod,
				Header: header,
				URL: &URL{
					Raw:   "{{domain}}" + urlHost + rawParams,
					Host:  []string{"{{domain}}"},
					Path:  strings.Split(strings.TrimPrefix(urlHost, "/"), "/"),
					Query: querys,
				},
			},
		}
	}

	return methodItem, nil
}

func (p Postman) transField(message *protogen.Message, recursion uint32) map[string]interface{} {
	var messageMap = make(map[string]interface{})
	for _, field := range message.Fields {
		fieldName := string(field.Desc.Name())

		switch field.Desc.Kind() {
		case protoreflect.BoolKind:
			messageMap[fieldName] = false
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind, protoreflect.Int64Kind,
			protoreflect.Sint64Kind, protoreflect.Uint64Kind, protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind,
			protoreflect.FloatKind, protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind, protoreflect.DoubleKind:
			messageMap[fieldName] = 0
		case protoreflect.EnumKind, protoreflect.StringKind, protoreflect.BytesKind:
			messageMap[fieldName] = ""
		case protoreflect.MessageKind:
			// 防止循环递归
			if recursion > 0 {
				messageMap[fieldName] = p.transField(field.Message, recursion-1)
			}
		case protoreflect.GroupKind:
			messageMap[fieldName] = "group"
		default:
			messageMap[fieldName] = "other"
		}

		if field.Desc.IsList() {
			messageMap[fieldName] = []interface{}{messageMap[fieldName]}
		}
	}

	return messageMap
}

func (p Postman) transParmas(pMap map[string]interface{}) []*Query {
	var q []*Query
	for k, v := range pMap {
		valueOf := reflect.ValueOf(v)
		if valueOf.Kind() == reflect.Map {
			var sonMap = make(map[string]interface{})
			keys := valueOf.MapKeys()
			for _, key := range keys {
				value := valueOf.MapIndex(key)
				sonMap[k+"."+fmt.Sprintf("%v", key)] = value
			}

			q = append(q, p.transParmas(sonMap)...)
		} else {
			q = append(q, &Query{
				Key:   k,
				Value: fmt.Sprintf("%v", v),
			})
		}
	}

	return q
}

func (p Postman) getMethodDescAndHeader(commentLeading protogen.Comments) (string, []*Header) {
	var desc string
	var header []*Header

	comments := strings.Split(string(commentLeading), "\n")
	for _, comment := range comments {
		commentArr := strings.Split(strings.TrimSpace(comment), " ")
		if commentArr[0] == COMMENTS_HEADER && len(commentArr) >= 2 {
			key := strings.TrimPrefix(commentArr[1], "*")
			header = append(header, &Header{
				Key:   GRPC_HEADER_PREFIX + key,
				Value: key,
				Type:  "text",
			})
		} else {
			desc += comment
		}
	}

	return desc, header
}
