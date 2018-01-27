package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"

	"./shared"
)

// A Chunk is the unit of reading/writing in DFS.
type Chunk [32]byte

// Represents a type of file access.
type FileMode int

const (
	// Read mode.
	READ FileMode = iota

	// Read/Write mode.
	WRITE

	// Disconnected read mode.
	DREAD
)

// Contains serverAddr
type DisconnectedError string

func (e DisconnectedError) Error() string {
	return fmt.Sprintf("DFS: Not connnected to server [%s]", string(e))
}

// Contains chunkNum that is unavailable
type ChunkUnavailableError uint8

func (e ChunkUnavailableError) Error() string {
	return fmt.Sprintf("DFS: Latest verson of chunk [%s] unavailable", string(e))
}

// Contains filename
type OpenWriteConflictError string

func (e OpenWriteConflictError) Error() string {
	return fmt.Sprintf("DFS: Filename [%s] is opened for writing by another client", string(e))
}

// Contains file mode that is bad.
type BadFileModeError FileMode

func (e BadFileModeError) Error() string {
	return fmt.Sprintf("DFS: Cannot perform this operation in current file mode [%s]", string(e))
}

// Contains filename.
type WriteModeTimeoutError string

func (e WriteModeTimeoutError) Error() string {
	return fmt.Sprintf("DFS: Write access to filename [%s] has timed out; reopen the file", string(e))
}

// Contains filename
type BadFilenameError string

func (e BadFilenameError) Error() string {
	return fmt.Sprintf("DFS: Filename [%s] includes illegal characters or has the wrong length", string(e))
}

// Contains filename
type FileUnavailableError string

func (e FileUnavailableError) Error() string {
	return fmt.Sprintf("DFS: Filename [%s] is unavailable", string(e))
}

// Contains local path
type LocalPathError string

func (e LocalPathError) Error() string {
	return fmt.Sprintf("DFS: Cannot access local path [%s]", string(e))
}

// Contains filename
type FileDoesNotExistError string

func (e FileDoesNotExistError) Error() string {
	return fmt.Sprintf("DFS: Cannot open file [%s] in D mode as it does not exist locally", string(e))
}

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
	LockedFileToClientID       map[string]int
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

func (s *ServerStruct) NotifyNewFile(args *shared.FileNameAndClientID, reply *shared.Reply) error {
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

func (s *ServerStruct) LockFile(args *shared.FileNameAndClientID, reply *shared.Reply) error {
	fmt.Println("locking file " + args.FileName)
	reply.Connected = true
	if _, ok := s.LockedFileToClientID[args.FileName]; ok {
		return OpenWriteConflictError(args.FileName)
	} else {
		s.LockedFileToClientID[args.FileName] = args.ClientID
		return nil
	}
}

func (s *ServerStruct) UnlockFileRPC(args *shared.FileNameAndClientID, reply *shared.Reply) error {
	delete(s.LockedFileToClientID, args.FileName)
	reply.Connected = true
	return nil
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
	dfsServer.LockedFileToClientID = map[string]int{}
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
