package grpc

import (
	"encoding/json"
	"grpc/codec"
	"io"
	"log"
	"net"
	"sync"
)

//服务端的实现

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int        // 表示这是一个通用的请求
	CodecType   codec.Type //客户端可以选择不同的编码器来执行
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

type Server struct{} //service代表一个rpc服务

func NewService() *Server {
	return &Server{}
}

var DefaultService = NewService()

//为每个传入的连接提供服务
func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("这是accept错误", err)
			return
		}
		go s.ServerConn(conn)
	}
}

//方便用户使用
// lis, _ := net.Listen("tcp", ":9999")
//	grpc.Accept(lis)
func Accept(lis net.Listener) {
	DefaultService.Accept(lis)
}

func (s *Server) ServerConn(conn io.ReadWriteCloser)  {
	defer func() {
		_ = conn.Close()
	}()

	var option Option
	if err := json.NewDecoder(conn).Decode(&option); err != nil {
		log.Println("这是一个decode的错误",err)
		return
	}

	if option.MagicNumber != MagicNumber {
		log.Println("magicNumber 不一样",option.MagicNumber)
		return
	}

	codecFunc := codec.NewCodeFuncMap[option.CodecType]

	if codecFunc == nil {
		log.Println("codecFunc错误",option.CodecType)
		return
	}

	s.serverCodec(codecFunc(conn))
}

func (s *Server)serverCodec(cc codec.Codec)  {
	sending := new(sync.Mutex)
	group := new(sync.WaitGroup)
	for  {
		req, err :=s.readRequest(cc)
		if err != nil {
			if req == nil {
				break // 连接无法恢复，所以关闭
			}
		}
		req.h
	}
}


