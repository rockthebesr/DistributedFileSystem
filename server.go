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
	GlobalFileExists(args shared.FileName) shared.FileExists
}
type ServerStruct struct {
	ClientInfoToClientID       map[shared.ClientInfo]int
	ClientIDToClientConnection map[int]*rpc.Client
	Client                     *rpc.Client
	GlobalFilesToClientIDs     map[string][]int
	GlobalFileToChunkVersions  map[string][]int
}

func (s *ServerStruct) CallClient(args *shared.ClientInfo, reply *shared.Reply) error {
	client, err := rpc.Dial("tcp", args.ClientIP)
	if err != nil {
		fmt.Println(err.Error())
	}
	newReply := shared.Reply{Connected: false}
	for clientInfo, id := range s.ClientInfoToClientID {
		if clientInfo.ClientLocalPath == args.ClientLocalPath && clientInfo.ClientIP == args.ClientIP {
			client.Call("ClientStruct.PrintClientID", id, &newReply)
			*reply = shared.Reply{true}
			s.ClientIDToClientConnection[id] = client
			return nil
		}
	}
	newID := rand.Int()
	s.ClientInfoToClientID[*args] = newID
	s.ClientIDToClientConnection[newID] = client
	client.Call("ClientStruct.PrintClientID", newID, &newReply)
	*reply = shared.Reply{true}
	return nil
}

func (s *ServerStruct) NotifyNewFile(args *shared.NotifyNewFile, reply *shared.Reply) error {
	fmt.Println("notify new file called")
	if val, ok := s.GlobalFilesToClientIDs[args.FileName]; ok {
		if !shared.Contains(val, args.ClientID) {
			val = append(val, args.ClientID)
		}
		fmt.Println(val)
		reply.Connected = true
		return nil
	} else {
		s.GlobalFilesToClientIDs[args.FileName] = []int{args.ClientID}
		fmt.Println(s.GlobalFilesToClientIDs[args.FileName])
		reply.Connected = true
		return nil
	}
}

func (s *ServerStruct) GlobalFileExists(args *shared.FileName, reply *shared.FileExists) error {
	fmt.Println("GlobalFileExists called")
	for fileName, _ := range s.GlobalFilesToClientIDs {
		if fileName == args.FileName {
			reply.FileExists = true
			return nil
		}
	}
	return nil
}

func main() {
	dfsServer := new(ServerStruct)
	dfsServer.ClientInfoToClientID = map[shared.ClientInfo]int{}
	dfsServer.ClientIDToClientConnection = map[int]*rpc.Client{}
	dfsServer.GlobalFilesToClientIDs = map[string][]int{}
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
