package main

import (
	"fmt"
	"net/rpc"
	"net/http"
)

type clientInfo struct {
	clientId int
	clientIp string
}

type callClientReply struct {
	connected bool
}

type Server struct {
	globalFilesToClients      []string
	globalFileToChunkVersions map[string][]int
}

func callClient(ci clientInfo) callClientReply {
	fmt.Println("calling client")
	client := rpc.Dial("tcp", clientInfo.clientIp)
}

func main() {
	fmt.Println("Starting server")
	server = rpc.NewServer()
	dfsServer := new(Server)
	server.RegisterName("server", dfsServer)
	l, e:= net.Listen("tcp", ":8080")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go server.Accept(l)
}
