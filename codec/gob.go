package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer //为了防止阻塞而创建的带缓冲的Writer
	dec  *gob.Decoder  //解码器 ~ 从远程读取数据信息
	end  *gob.Encoder  //编码器 ~ 管理数据信息发送到远程
}

var _ Codec = (*GobCodec)(nil)

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		//函数返回一个从conn读取数据的*Decoder，如果conn不满足io.ByteReader接口，则会包装r为bufio.Reader。
		dec:  gob.NewDecoder(conn),
		//NewEncoder返回一个将编码后数据写入w的*Encoder。
		end:  gob.NewEncoder(buf),
	}
}

func (g *GobCodec) ReadHeader(header *Header) (err error) {
	err = g.dec.Decode(header)
	return
}

func (g *GobCodec) ReadBody(body interface{}) (err error) {
	//Decode从输入流读取下一个之并将该值存入e。如果e是nil，将丢弃该值；
	// 否则e必须是可接收该值的类型的指针。如果输入结束，方法会返回io.EOF并且不修改e（指向的值）。
	err = g.dec.Decode(body)
	return
}

func (g *GobCodec) Write(header *Header, body interface{}) (err error) {
	defer func() {
		_ = g.buf.Flush()
		if err != nil {
			_ = g.Close()
		}
	}()
	if err = g.end.Encode(header); err != nil {
		log.Println("这是Encode错误header", err)
		return
	}
	if err = g.end.Encode(body); err != nil {
		log.Println("这是Encode错误body", err)
		return
	}
	return
}

func (g *GobCodec) Close() error {
	return g.conn.Close()
}
