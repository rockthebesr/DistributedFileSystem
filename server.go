package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	"./shared"
)

type Server interface {
	CallClient(ci shared.ClientInfo) shared.Reply
}
type ServerStruct struct {
	Client                    *rpc.Client
	GlobalFilesToClients      []string
	GlobalFileToChunkVersions map[string][]int
}

func (s *ServerStruct) CallClient(args *shared.ClientInfo, reply *shared.Reply) error {
	fmt.Println("calling client" + args.ClientIP)
	client, err := rpc.Dial("tcp", args.ClientIP)
	if err != nil {
		fmt.Println(err.Error())
	}
	s.Client = client
	newReply := shared.Reply{Connected: false}
	err = client.Call("ClientStruct.PrintClientID", args, &newReply)
	fmt.Println(newReply.Connected)
	*reply = shared.Reply{true}
	return nil
}

func main() {
	dfsServer := new(ServerStruct)
	// server := rpc.NewServer()
	rpc.RegisterName("ServerStruct", dfsServer)
	l, e := net.Listen("tcp", ":8080")

	fmt.Println("Starting server" + l.Addr().String())
	if e != nil {
		log.Fatal("listen error:", e)
	}
	for {
		conn, _ := l.Accept()
		rpc.ServeConn(conn)
	}
}
