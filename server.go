package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strconv"

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
	//Client Info Mapped To Client ID
	ClientInfoToClientID map[shared.ClientInfo]int

	//Client Id Mapped To *rpc.Client
	ClientIDToClientConnection map[int]*rpc.Client

	//Client Id Mapped To a map of File Name To Chunk Versions
	ClientIDToFileNameToChunkVersions map[int]map[string][]int

	//Global Files mapped to array of array of Client IDs
	// example key value: {"helloworld" => [[1, 2], [1,3]]}
	// This means that for the file "helloworld", for chunk 0 is last updated by client 2
	// chunk 1 is last updated by client 3
	GlobalFileToChunksToClientIDs map[string][][]int

	// Locked file to client id of the client that has the lock
	LockedFileToClientID map[string]int
}

func (s *ServerStruct) CallClient(args *shared.ClientInfo, reply *shared.Reply) error {
	clientAddr, err := net.ResolveTCPAddr("tcp", args.ClientAddr)
	if err != nil {
		fmt.Println(err.Error())
	}
	client, err := rpc.Dial("tcp", clientAddr.String())
	if err != nil {
		fmt.Println(err.Error())
	}
	newReply := shared.Reply{Connected: false}
	for clientInfo, id := range s.ClientInfoToClientID {
		if clientInfo.ClientLocalPath == args.ClientLocalPath && clientInfo.ClientIP == args.ClientIP {

			fmt.Println("adding existing clientID " + strconv.Itoa(id))
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
	fmt.Println("adding clientID " + strconv.Itoa(newID))
	*reply = shared.Reply{true}
	return nil
}

func (s *ServerStruct) CloseClient(args *shared.ClientID, reply *shared.Reply) error {
	if client, ok := s.ClientIDToClientConnection[args.ClientID]; ok {
		err := client.Close()
		delete(s.ClientIDToClientConnection, args.ClientID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ServerStruct) NotifyNewFile(args *shared.FileNameAndClientID, reply *shared.Reply) error {
	fmt.Println("notify new file called")
	if _, ok := s.GlobalFileToChunksToClientIDs[args.FileName]; !ok {
		fmt.Println("adding new global file " + args.FileName)
		chunksToClientIDs := make([][]int, 256)
		for i := 0; i < 256; i++ {
			var ids []int
			chunksToClientIDs[i] = ids
		}
		s.GlobalFileToChunksToClientIDs[args.FileName] = chunksToClientIDs

	} else {

		fmt.Println(args.FileName + " did not add")
		reply.Connected = true
	}
	chunkVersions := make([]int, 256)
	fileNameToChunkVersions := map[string][]int{}
	fileNameToChunkVersions[args.FileName] = chunkVersions
	s.ClientIDToFileNameToChunkVersions[args.ClientID] = fileNameToChunkVersions
	reply.Connected = true
	return nil
}

func (s *ServerStruct) NotifyChunkVersionUpdate(args *shared.FileNameAndChunkNumberAndClientID, reply *shared.Reply) error {
	fmt.Println("updating file " + args.FileName + " chunk " + strconv.Itoa(args.ChunkNumber))
	clientIDs := s.GlobalFileToChunksToClientIDs[args.FileName][args.ChunkNumber]
	fmt.Println(strconv.Itoa(args.ClientID))
	s.GlobalFileToChunksToClientIDs[args.FileName][args.ChunkNumber] = append(clientIDs, args.ClientID)
	fmt.Println(s.GlobalFileToChunksToClientIDs[args.FileName][args.ChunkNumber])
	fmt.Println(len(s.GlobalFileToChunksToClientIDs[args.FileName][args.ChunkNumber]))
	s.ClientIDToFileNameToChunkVersions[args.ClientID][args.FileName][args.ChunkNumber] = len(s.GlobalFileToChunksToClientIDs[args.FileName][args.ChunkNumber])
	reply.Connected = true
	return nil
}

func (s *ServerStruct) GetLatestFileRPC(args *shared.FileNameAndClientID, reply *shared.FileData) error {
	fmt.Println("get latest file " + args.FileName + " for client " + strconv.Itoa(args.ClientID))
	result, versions, err := s.GetLatestFile(args.FileName, args.ClientID)
	if err != nil {
		return err
	} else {
		copy(reply.ChunkVersions[:], versions[0:256])
		copy(reply.FileData[0:8192], result[:])
		return nil
	}
}

func (s *ServerStruct) GetLatestChunkRPC(args *shared.FileNameAndChunkNumberAndClientID, reply *shared.ChunkData) error {
	fmt.Println("get latest chunk " + strconv.Itoa(args.ChunkNumber) + " for file " + args.FileName)
	result, version, err := s.GetLatestChunk(args.FileName, args.ClientID, args.ChunkNumber)
	if err != nil {
		return err
	} else {
		copy(reply.ChunkData[0:32], result[:])
		s.ClientIDToFileNameToChunkVersions[args.ClientID][args.FileName][args.ChunkNumber] = version
		return nil
	}
}

func (s *ServerStruct) GetLatestFile(fname string, clientID int) ([8192]byte, [256]int, error) {
	result := [8192]byte{}
	versions := [256]int{}
	for i := 0; i < 256; i++ {
		data, version, err := s.GetLatestChunk(fname, clientID, i)
		if err != nil {
			return [8192]byte{}, versions, FileUnavailableError(fname)
		}
		copy(result[i*32:(i+1)*32], data[:])
		versions[i] = version
	}

	return [8192]byte{}, versions, nil
}

func (s *ServerStruct) GetLatestChunk(fname string, clientID int, chunkNumber int) ([32]byte, int, error) {

	data := [32]byte{}
	chunksToClientIDs := s.GlobalFileToChunksToClientIDs[fname]
	latestClientIDs := chunksToClientIDs[chunkNumber]
	if len(latestClientIDs) == 0 {
		return data, 0, nil
	}
	currentVersion := s.ClientIDToFileNameToChunkVersions[clientID][fname][chunkNumber]
	println("current version " + fname + strconv.Itoa(chunkNumber) + ": " + strconv.Itoa(currentVersion))
	if currentVersion >= len(latestClientIDs) {
		if conn, ok := s.ClientIDToClientConnection[clientID]; ok {
			args := shared.FileNameAndChunkNumberAndClientID{fname, chunkNumber, clientID}
			reply := shared.ChunkData{ChunkData: data}
			err := conn.Call("ClientStruct.ReadChunk", args, &reply)
			if err != nil {
				fmt.Println(err)
				return [32]byte{}, 0, ChunkUnavailableError(chunkNumber)
			} else {
				return reply.ChunkData, len(s.GlobalFileToChunksToClientIDs[fname][chunkNumber]), nil
			}
		}
	}
	for j := len(latestClientIDs) - 1; j >= currentVersion; j-- {
		// fmt.Println("latest updated client for file " + fname + " chunk " + strconv.Itoa(chunkNumber) + " is " + strconv.Itoa(latestClientIDs[j]))
		if conn, ok := s.ClientIDToClientConnection[latestClientIDs[j]]; ok {
			args := shared.FileNameAndChunkNumberAndClientID{fname, chunkNumber, latestClientIDs[j]}
			reply := shared.ChunkData{ChunkData: data}
			err := conn.Call("ClientStruct.ReadChunk", args, &reply)
			if err != nil {
				fmt.Println(err)
				return [32]byte{}, 0, ChunkUnavailableError(chunkNumber)
			} else {
				return reply.ChunkData, len(s.GlobalFileToChunksToClientIDs[fname][chunkNumber]), nil
			}
		}
	}

	fmt.Println("cant get latest chunk!")
	return data, 0, ChunkUnavailableError(chunkNumber)
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
	if _, ok := s.LockedFileToClientID[args.FileName]; ok {
		delete(s.LockedFileToClientID, args.FileName)
	}
	fmt.Println("unlocked file " + args.FileName)
	reply.Connected = true
	return nil
}

func (s *ServerStruct) GlobalFileExists(args *shared.FileName, reply *shared.FileExists) error {
	fmt.Println("GlobalFileExists called")
	if _, ok := s.GlobalFileToChunksToClientIDs[args.FileName]; ok {
		reply.FileExists = true
		return nil
	}
	return nil
}

func main() {
	serverIPAndPort, err := net.ResolveTCPAddr("tcp", os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	dfsServer := new(ServerStruct)
	dfsServer.ClientInfoToClientID = map[shared.ClientInfo]int{}
	dfsServer.ClientIDToClientConnection = map[int]*rpc.Client{}
	dfsServer.ClientIDToFileNameToChunkVersions = map[int]map[string][]int{}
	dfsServer.GlobalFileToChunksToClientIDs = map[string][][]int{}
	dfsServer.LockedFileToClientID = map[string]int{}
	rpc.RegisterName("ServerStruct", dfsServer)

	fmt.Println("Starting server " + serverIPAndPort.String())

	l, e := net.Listen("tcp", serverIPAndPort.String())
	if e != nil {
		log.Fatal("listen error:", e)
	}
	for {
		conn, _ := l.Accept()
		go rpc.ServeConn(conn)
	}
}
