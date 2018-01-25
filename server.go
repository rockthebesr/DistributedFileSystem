package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"

	"./shared"
)

type Server interface {
	CallClient(ci shared.ClientInfo) shared.Reply
}
type ServerStruct struct {
	ClientInfoToClientID      map[shared.ClientInfo]int
	Client                    *rpc.Client
	GlobalFilesToClients      []string
	GlobalFileToChunkVersions map[string][]int
}

func (s *ServerStruct) CallClient(args *shared.ClientInfo, reply *shared.Reply) error {
	client, err := rpc.Dial("tcp", args.ClientIP)
	if err != nil {
		fmt.Println(err.Error())
	}
	s.Client = client
	newReply := shared.Reply{Connected: false}
	for clientInfo, id := range s.ClientInfoToClientID {
		if clientInfo.ClientLocalPath == args.ClientLocalPath && clientInfo.ClientIP == args.ClientIP {

			// fmt.Println("args")
			// fmt.Println(args.ClientIP)
			// fmt.Println(args.ClientLocalPath)
			// fmt.Println("clientInfo")
			// fmt.Println(clientInfo.ClientIP)
			// fmt.Println(clientInfo.ClientLocalPath)
			client.Call("ClientStruct.PrintClientID", id, &newReply)
			// fmt.Println("Exisiting ID: ")
			// fmt.Println(id)
			*reply = shared.Reply{true}
			return nil
		}
	}
	newID := rand.Int()
	s.ClientInfoToClientID[*args] = newID
	client.Call("ClientStruct.PrintClientID", newID, &newReply)
	*reply = shared.Reply{true}
	return nil
}

func main() {
	dfsServer := new(ServerStruct)
	dfsServer.ClientInfoToClientID = map[shared.ClientInfo]int{}
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
