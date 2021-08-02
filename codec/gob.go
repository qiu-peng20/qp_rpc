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
		dec:  gob.NewDecoder(conn),
		end:  gob.NewEncoder(buf),
	}
}

func (g *GobCodec) ReadHeader(header *Header) (err error) {
	err = g.dec.Decode(header)
	return
}

func (g *GobCodec) ReadBody(body interface{}) (err error) {
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
