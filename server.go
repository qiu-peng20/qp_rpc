package grpc

import (
	"encoding/json"
	"fmt"
	"grpc/codec"
	"io"
	"log"
	"net"
	"reflect"
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

func (s *Server) ServerConn(conn io.ReadWriteCloser) {
	defer func() {
		_ = conn.Close()
	}()

	var option Option
	if err := json.NewDecoder(conn).Decode(&option); err != nil {
		log.Println("这是一个decode的错误", err)
		return
	}

	if option.MagicNumber != MagicNumber {
		log.Println("magicNumber 不一样", option.MagicNumber)
		return
	}

	codecFunc := codec.NewCodeFuncMap[option.CodecType]

	if codecFunc == nil {
		log.Println("codecFunc错误", option.CodecType)
		return
	}

	s.serverCodec(codecFunc(conn))
}

var invalidRequest = struct{}{}

func (s *Server) serverCodec(cc codec.Codec) {
	sending := new(sync.Mutex) //确保发送完整回复
	wg := new(sync.WaitGroup)  //等待所有请求处理完毕
	for {
		req, err := s.readRequest(cc)
		if err != nil {
			if req == nil {
				break // 连接无法恢复，所以关闭
			}
			req.h.Error = err.Error()
			s.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go s.handleRequest(cc, req, sending, wg)
	}
	//TODO: 为什么这里需要使用wait
	wg.Wait()
	_ = cc.Close()
}

type request struct {
	h      *codec.Header
	argv   reflect.Value
	replyv reflect.Value
}

func (s *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("read header error", err)
		}
		return nil, err
	}
	return &h, nil
}

func (s *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := s.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{
		h:    h,
		argv: reflect.New(reflect.TypeOf("")),
	}
	//TODO: 现在我们不知道请求类型
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("read argv err",err)
	}
	return req,nil
}

// 回复请求
func (s *Server)sendResponse(cc codec.Codec,h *codec.Header,body interface{},sending *sync.Mutex)  {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h,body); err != nil {
		log.Println("response err",err)
	}
}

//处理请求
func (s *Server)handleRequest(cc codec.Codec,req *request,sending *sync.Mutex,wg *sync.WaitGroup)  {
	//TODO : 应该调用注册的rpc方法来获取正确的回复
	defer wg.Done()
	log.Println(req.h,req.argv.Elem())
	req.replyv = reflect.ValueOf(fmt.Sprintf("resp %d",req.h.Seq))
	s.sendResponse(cc,req.h,req.replyv.Interface(),sending)
}
