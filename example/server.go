package main

import (
	"errors"
	"log"
	"net"

	"github.com/puper/rpc/example/proto"

	"github.com/puper/rpc"

	"io"

	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/puper/codec"
)

type Front struct {
}

func (t *Front) Auth(args *proto.AuthArgs, reply *proto.AuthReply) error {
	if args.User == "puper" {
		reply.Success = true
	} else {
		reply.Success = false
	}
	return nil
}

func (t *Front) Mul(args *proto.ProtoArgs, reply *proto.ProtoReply) error {
	reply.C = args.A * args.B
	return nil
}

var invalidRequest = &empty.Empty{}

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
		codec.Close()
		log.Println(err)
		return
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
	if req.ServiceMethod != "Front.Auth" {
		this.server.SendResponse(sending, req, invalidRequest, codec, "")
		this.server.FreeRequest(req)
		return errors.New("not auth service")
	}
	service.Call(this.server, sending, mtype, req, argv, replyv, codec)
	reply := replyv.Interface().(*proto.AuthReply)
	if !reply.Success {
		err = codec.WriteResponse(&rpc.Response{
			ServiceMethod: "Callback.Test",
		}, invalidRequest)
		log.Println(err)
		return errors.New("auth failed")
	}
	return nil
}

func (this *Server) RegisterName(name string, rcvr interface{}) error {
	return this.server.RegisterName(name, rcvr)
}

func main() {
	server := NewServer(":8081")
	server.RegisterName("Front", new(Front))
	server.Start()
	log.Println(1111)
}
