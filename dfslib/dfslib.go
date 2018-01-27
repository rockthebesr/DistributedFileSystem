/*

This package specifies the application's interface to the distributed
file system (DFS) system to be used in assignment 2 of UBC CS 416
2017W2.

*/

package dfslib

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"unicode"

	"../shared"
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

////////////////////////////////////////////////////////////////////////////////////////////
// <ERROR DEFINITIONS>

// These type definitions allow the application to explicitly check
// for the kind of error that occurred. Each API call below lists the
// errors that it is allowed to raise.
//
// Also see:
// https://blog.golang.org/error-handling-and-go
// https://blog.golang.org/errors-are-values

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

// </ERROR DEFINITIONS>
////////////////////////////////////////////////////////////////////////////////////////////

// Represents a file in the DFS system.
type DFSFile interface {
	// Reads chunk number chunkNum into storage pointed to by
	// chunk. Returns a non-nil error if the read was unsuccessful.
	//
	// Can return the following errors:
	// - DisconnectedError (in READ,WRITE modes)
	// - ChunkUnavailableError (in READ,WRITE modes)
	Read(chunkNum uint8, chunk *Chunk) (err error)

	// Writes chunk number chunkNum from storage pointed to by
	// chunk. Returns a non-nil error if the write was unsuccessful.
	//
	// Can return the following errors:
	// - BadFileModeError (in READ,DREAD modes)
	// - DisconnectedError (in WRITE mode)
	// - WriteModeTimeoutError (in WRITE mode)
	Write(chunkNum uint8, chunk *Chunk) (err error)

	// Closes the file/cleans up. Can return the following errors:
	// - DisconnectedError
	Close() (err error)
}

type DFSFileStruct struct {
	FileName string
	File     *os.File
	DFS      DFSInstance
	FileMode FileMode
}

func (file DFSFileStruct) Read(chunkNum uint8, chunk *Chunk) (err error) {
	result := make([]byte, 32)
	if file.FileMode == DREAD {
		_, err = file.File.ReadAt(result, int64(chunkNum*32))
		if err != nil {
			return err
		}
		for _, element := range result {
			chunk[0] = element
		}
		return nil
	} else {
		//TODO: Server check get latest version
		return nil
	}
}

func (file DFSFileStruct) Write(chunkNum uint8, chunk *Chunk) (err error) {
	if file.FileMode != WRITE {
		return BadFileModeError(file.FileMode)
	}
	result := make([]byte, 32)
	for _, element := range chunk {
		result[0] = element
	}
	n, err := file.File.WriteAt(result, int64(chunkNum*32))
	if err != nil {
		return err
	}
	if n != 0 {
		//TODO notify chunk version change
	}
	return nil
}

func (file DFSFileStruct) Close() (err error) {
	err = file.File.Close()
	if err != nil {
		return err
	}
	if file.FileMode == WRITE {
		err = file.DFS.UnlockFile(file.FileName)
		if err != nil {
			return err
		}
	}
	return nil
}

// Represents a connection to the DFS system.
type DFS interface {
	// Check if a file with filename fname exists locally (i.e.,
	// available for DREAD reads).
	//
	// Can return the following errors:
	// - BadFilenameError (if filename contains non alpha-numeric chars or is not 1-16 chars long)
	LocalFileExists(fname string) (exists bool, err error)

	// Check if a file with filename fname exists globally.
	//
	// Can return the following errors:
	// - BadFilenameError (if filename contains non alpha-numeric chars or is not 1-16 chars long)
	// - DisconnectedError
	GlobalFileExists(fname string) (exists bool, err error)

	// Opens a filename with name fname using mode. Creates the file
	// in READ/WRITE modes if it does not exist. Returns a handle to
	// the file through which other operations on this file can be
	// made.
	//
	// Can return the following errors:
	// - OpenWriteConflictError (in WRITE mode)
	// - DisconnectedError (in READ,WRITE modes)
	// - FileUnavailableError (in READ,WRITE modes)
	// - FileDoesNotExistError (in DREAD mode)
	// - BadFilenameError (if filename contains non alpha-numeric chars or is not 1-16 chars long)
	Open(fname string, mode FileMode) (f DFSFile, err error)

	// Disconnects from the server. Can return the following errors:
	// - DisconnectedError
	UMountDFS() (err error)
}

//DFSInstance is an instance of the DFS lib
type DFSInstance struct {
	ClientID                 int
	IsConnected              bool
	Server                   *rpc.Client
	LocalPath                string
	LocalIP                  string
	LocalFiles               []string
	LocalFileToChunkVersions map[string][]int
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func GoodFileName(name string) bool {
	if len(name) < 1 || len(name) > 16 {
		return false
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

func (dfs DFSInstance) LocalFileExists(fname string) (exists bool, err error) {
	if !GoodFileName(fname) {
		return false, BadFilenameError(fname)
	}
	if Exists(dfs.LocalPath + fname + ".dfs") {
		return true, nil
	}
	return false, FileDoesNotExistError(fname)
}

func (dfs DFSInstance) GlobalFileExists(fname string) (exists bool, err error) {
	if !GoodFileName(fname) {
		return false, BadFilenameError(fname)
	}
	args := shared.FileName{fname}
	reply := shared.FileExists{false}
	dfs.Server.Call("ServerStruct.GlobalFileExists", args, &reply)
	if reply.FileExists {
		return true, nil
	}
	return false, FileDoesNotExistError(fname)
}

func (dfs DFSInstance) LockFile(fname string) error {
	args := shared.FileNameAndClientID{fname, dfs.ClientID}
	reply := shared.Reply{false}
	err := dfs.Server.Call("ServerStruct.LockFile", args, &reply)
	return err
}

func (dfs DFSInstance) UnlockFile(fname string) error {
	args := shared.FileNameAndClientID{fname, dfs.ClientID}
	reply := shared.Reply{false}
	err := dfs.Server.Call("ServerStruct.UnlockFileRPC", args, &reply)
	return err
}

// Opens a filename with name fname using mode. Creates the file
// in READ/WRITE modes if it does not exist. Returns a handle to
// the file through which other operations on this file can be
// made.
//
// Can return the following errors:
// - OpenWriteConflictError (in WRITE mode)
// - DisconnectedError (in READ,WRITE modes)
// - FileUnavailableError (in READ,WRITE modes)
// - FileDoesNotExistError (in DREAD mode)
// - BadFilenameError (if filename contains non alpha-numeric chars or is not 1-16 chars long)

func (dfs DFSInstance) Open(fname string, mode FileMode) (f DFSFile, err error) {
	if !GoodFileName(fname) {
		return nil, BadFilenameError(fname)
	}
	switch mode {
	case READ:
		fmt.Println("open read file called")
		//check if file exists globally
		fileExists, _ := dfs.GlobalFileExists(fname)
		filePath := dfs.LocalPath + fname + ".dfs"
		if !fileExists {
			file, err := os.Create(filePath)
			if err != nil {
				fmt.Println(err)
			}
			bytes := make([]byte, 256*32)
			file.Write(bytes)
			dfs.NotifyNewFile(fname)
			return DFSFileStruct{File: file, DFS: dfs, FileMode: READ, FileName: fname}, nil
		} else {
			//TODO
			//need to check if file is available
			//call serverstruct.getFile
			file, _ := os.Open(filePath)
			return DFSFileStruct{File: file, DFS: dfs, FileMode: READ, FileName: fname}, nil
		}
	case WRITE:
		fmt.Println("open write file called")
		//check if file exists globally
		fileExists, _ := dfs.GlobalFileExists(fname)
		filePath := dfs.LocalPath + fname + ".dfs"
		if !fileExists {
			file, err := os.Create(filePath)
			if err != nil {
				fmt.Println(err)
			}
			bytes := make([]byte, 256*32)
			file.Write(bytes)
			dfs.NotifyNewFile(fname)
			err = dfs.LockFile(fname)
			if err != nil {
				return nil, err
			}
			return DFSFileStruct{File: file, DFS: dfs, FileMode: WRITE, FileName: fname}, nil
		} else {
			//TODO
			//need to check if file is available
			// call serverstruct.getFile(fname)
			//call serverstruct.LockFile(fname)
			file, _ := os.Open(filePath)
			return DFSFileStruct{File: file, DFS: dfs, FileMode: WRITE, FileName: fname}, nil
		}
	case DREAD:
		if b, _ := dfs.LocalFileExists(fname); !b {
			return nil, FileDoesNotExistError(fname)
		}
		filePath := dfs.LocalPath + fname + ".dfs"
		file, _ := os.Open(filePath)
		return DFSFileStruct{File: file, DFS: dfs, FileMode: DREAD, FileName: fname}, nil
	}

	return nil, nil
}

func (dfs DFSInstance) UMountDFS() (err error) {
	return nil
}

func (dfs DFSInstance) NotifyNewFile(fileName string) {
	args := shared.FileNameAndClientID{FileName: fileName, ClientID: dfs.ClientID}
	reply := shared.Reply{false}
	dfs.Server.Call("ServerStruct.NotifyNewFile", args, &reply)
}

var existDFSInstance *DFSInstance

type Listener int

type Client interface {
	PrintClientID(ci shared.ClientInfo) shared.Reply
}

type ClientStruct struct {
	ClientID int
}

func (s *ClientStruct) PrintClientID(id int, reply *shared.Reply) error {
	s.ClientID = id
	fmt.Println(s.ClientID)
	reply.Connected = true
	return nil
}

// The constructor for a new DFS object instance. Takes the server's
// IP:port address string as parameter, the localIP to use to
// establish the connection to the server, and a localPath path on the
// local filesystem where the client has allocated storage (and
// possibly existing state) for this DFS.
//
// The returned dfs instance is singleton: an application is expected
// to interact with just one dfs at a time.
//
// This call should succeed regardless of whether the server is
// reachable. Otherwise, applications cannot access (local) files
// while disconnected.
//
// Can return the following errors:
// - LocalPathError
// - Networking errors related to localIP or serverAddr

func MountDFS(serverAddr string, localIP string, localPath string) (dfs DFS, err error) {
	// TODO
	// For now return LocalPathError
	// if dfs instance already exists

	if existDFSInstance != nil {
		return existDFSInstance, nil
	}
	if !Exists(localPath) {
		return nil, LocalPathError(localPath)
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", localIP+":1111")
	conn, err := net.Listen("tcp", tcpAddr.String())
	client := new(ClientStruct)
	rpc.RegisterName("ClientStruct", client)
	go rpc.Accept(conn)
	server, err := rpc.Dial("tcp", serverAddr)

	var rep shared.Reply
	server.Call("ServerStruct.CallClient", shared.ClientInfo{ClientLocalPath: localPath, ClientIP: tcpAddr.String()}, &rep)
	fmt.Println("called server now")
	isConnected := err != nil
	localDFSInstance := DFSInstance{IsConnected: isConnected, Server: server, LocalIP: localIP, LocalPath: localPath}

	return localDFSInstance, nil

}
