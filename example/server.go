package main

import (
	"errors"
	"log"
	"net"
	"test/proto"

	"github.com/puper/rpc"

	"io"

	"sync"

	"github.com/puper/codec"
)

type Args struct {
	A int
	B int
}

type Reply struct {
	C int
}

type Arith int

func (t *Arith) Mul(args *proto.ProtoArgs, reply *proto.ProtoReply) error {
	reply.C = args.A * args.B
	return nil
}

var invalidRequest = struct{}{}

type Server struct {
	addr     string
	server   *rpc.Server
	listener net.Listener
}

func NewServer(addr string) *Server {
	return &Server{
		addr:   addr,
		server: rpc.NewServer(),
	}
}

func (this *Server) Start() {
	var (
		err error
	)
	this.listener, err = net.Listen("tcp", this.addr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := this.listener.Accept()
		if err != nil {
			log.Print("rpc.Serve: accept:", err.Error())
			return
		}
		go this.ServeConn(conn)
	}
}

func (this *Server) ServeConn(conn io.ReadWriteCloser) {
	srv := codec.NewServerCodec(conn)
	this.ServeCodec(srv)
}

func (this *Server) ServeCodec(codec rpc.ServerCodec) {
	var (
		err error
	)
	if err = this.Auth(codec); err != nil {
		log.Println(err)
	}
	sending := new(sync.Mutex)
	for {
		service, mtype, req, argv, replyv, keepReading, err := this.server.ReadRequest(codec)
		if err != nil {
			if !keepReading {
				break
			}
			// send a response if we actually managed to read a header.
			if req != nil {
				this.server.SendResponse(sending, req, invalidRequest, codec, err.Error())
				this.server.FreeRequest(req)
			}
			continue
		}
		go service.Call(this.server, sending, mtype, req, argv, replyv, codec)
	}
	codec.Close()
}

func (this *Server) Auth(codec rpc.ServerCodec) error {
	sending := new(sync.Mutex)
	service, mtype, req, argv, replyv, keepReading, err := this.server.ReadRequest(codec)
	if err != nil {
		if !keepReading {
			return err
		}
		// send a response if we actually managed to read a header.
		if req != nil {
			this.server.SendResponse(sending, req, invalidRequest, codec, err.Error())
			this.server.FreeRequest(req)
		}
		return err
	}
	if req.ServiceMethod == "Arith.Mul" {
		this.server.SendResponse(sending, req, invalidRequest, codec, err.Error())
		this.server.FreeRequest(req)
		return errors.New("not auth service")
	}
	return nil
}

func (this *Server) RegisterName(name string, rcvr interface{}) error {
	return this.server.RegisterName(name, rcvr)
}

func main() {
	server := NewServer(":8081")
	server.RegisterName("Arith", new(Arith))
	server.Start()
	log.Println(1111)
}
