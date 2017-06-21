package main

import (
	"log"
	"net"

	"github.com/puper/rpc/example/proto"

	"github.com/puper/rpc"

	"time"

	"github.com/puper/codec"
)

type Client struct {
	client *rpc.Client
	addr   string
}

func NewClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	codec_ := codec.NewClientCodec(conn)
	client := rpc.NewClientWithCodec(codec_)
	client.CallbackFunc = func(*rpc.Client, rpc.ClientCodec, rpc.Response) error {
		log.Println(11111111)
		return nil
	}
	client.CallbackPrefix = "Callback"
	return &Client{
		client: client,
		addr:   addr,
	}, nil
}

func (this *Client) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return this.client.Call(serviceMethod, args, reply)
}

func main() {
	c, err := NewClient(":8081")
	if err != nil {
		panic(err)
	}
	req := new(proto.AuthArgs)
	reply := new(proto.AuthReply)
	req.User = "puper1"
	err = c.Call("Front.Auth", req, reply)
	log.Println(err)
	log.Println(reply.Success)

	req1 := new(proto.ProtoArgs)
	req1.A = 5
	req1.B = 7
	reply1 := new(proto.ProtoReply)
	//err = c.Call("Front.Mul", req1, reply1)
	log.Println(err)
	log.Println(reply1.C)
	time.Sleep(5 * time.Second)

}
