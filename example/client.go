package main

import (
	"log"
	"net"

	"github.com/puper/rpc/example/proto"

	"github.com/puper/rpc"

	"github.com/golang/protobuf/ptypes/empty"
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
	client.CallbackFunc = func(client *rpc.Client, codec_ rpc.ClientCodec, response rpc.Response) error {
		reply := &empty.Empty{}
		codec_.ReadResponseBody(reply)
		log.Println(response)
		return nil
	}
	client.CallbackPrefix = "Callback"
	return &Client{
		client: client,
		addr:   addr,
	}, nil
}
func (this *Client) Rigester() {

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
	req.User = "puper"
	err = c.Call("Front.Auth", req, reply)
	if err != nil {
		log.Println(err)
	}
	log.Println(reply.Success)

	req1 := new(proto.ProtoArgs)
	req1.A = 5
	req1.B = 7
	reply1 := new(proto.ProtoReply)
	err = c.Call("Front.Mul", req1, reply1)
	if err != nil {
		log.Println(err)
	}
	//log.Println(reply1.C)

}
