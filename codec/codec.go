package codec

import "io"

type Header struct {
	ServiceMethod string // 服务名或者是方法名
	Seq           uint64 // 识别某个请求
	Error         string
}

type Codec interface {
	io.Closer
	ReadHeader(header *Header) error
	ReadBody(body interface{}) error
	Write(header *Header, body interface{}) error
}

type NewCodecFunc func(at io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

var NewCodeFuncMap map[Type]NewCodecFunc

func init()  {
	NewCodeFuncMap = make(map[Type]NewCodecFunc)
	NewCodeFuncMap[GobType] = NewGobCodec
}


