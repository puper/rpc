package main

import (
	"log"
	"net"
	"test/proto"

	"github.com/puper/rpc"

	"github.com/puper/codec"
)

func main() {
	conn, err := net.Dial("tcp", ":8081")
	if err != nil {
		panic(err)
	}
	codec_ := codec.NewClientCodec(conn)
	c := rpc.NewClientWithCodec(codec_)
	req := new(proto.ProtoArgs)
	reply := new(proto.ProtoReply)
	req.A = 5
	req.B = 7
	for i := 0; i < 100; i++ {
		err = c.Call("Arith.Mul", req, reply)
		log.Println(err)
		log.Println(reply)
	}

}
