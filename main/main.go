package main

import (
	"encoding/json"
	"fmt"
	"grpc"
	"grpc/codec"
	"log"
	"net"
	"time"
)

func startService(addr chan string)  {
	l, err := net.Listen("tcp", ":7878")
	if err != nil {
		log.Fatal("netWork error",err)
	}
	log.Println("server on",l.Addr())
	addr <- l.Addr().String()
	grpc.Accept(l)
}

func main()  {
	addr := make(chan string)
	go startService(addr)

	conn , _ := net.Dial("tcp",<-addr)
	defer func() {
		_ = conn.Close()
	}()

	time.Sleep(time.Second)
	//NewEncoder创建一个将数据写入w的*Encoder。
	_ = json.NewEncoder(conn).Encode(grpc.DefaultOption)
	cc := codec.NewGobCodec(conn)

	for i:=0;i<5 ;i++  {
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(1),
		}
		_ = cc.Write(h, fmt.Sprintf("req %d", h.Seq))
		_ = cc.ReadHeader(h)
		var reply string
		_ = cc.ReadBody(&reply)
		log.Println("reply:",reply)
	}
}
