package main

import (
	"fmt"
	hashtable "github.com/ingridkarinaf/DistributedHashTable/interface"
	grpc "google.golang.org/grpc"
	"strconv"
	"os"
	"context"
	"net"
	"log"
	"time"
)

/*
	- Responsible for maintaining copies of data for FEServer. 
	- Is subject to crashing.
*/

type RMServer struct {
	hashtable.UnimplementedHashTableServer
	id              int32 //portnumber, between 5000 and 5002
	ctx             context.Context
	hashTableCopy   map[int32]int32
	lockChannel 	chan bool
}



func main() {
	//log to file instead of console
	f := setLogRMServer()
	defer f.Close()

	portInput, _ := strconv.ParseInt(os.Args[1], 10, 32) //Takes arguments 5000, 5001 and 5002
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rmServer := &RMServer{
		id:              int32(portInput),
		hashTableCopy:   make(map[int32]int32),
		ctx:             ctx,
		lockChannel: 	make(chan bool, 1),
	}

	//Unlock channel
	rmServer.lockChannel <- true

	list, err := net.Listen("tcp", fmt.Sprintf(":%v", portInput))
	if err != nil {
		log.Fatalf("RM Server %v: Failed to listen on port: %v", rmServer.id, err)
	}

	grpcServer := grpc.NewServer()
	hashtable.RegisterHashTableServer(grpcServer, rmServer)
	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("RM server %v failed to serve %v", rmServer.id, err)
		}
	}()

	for {}
}

func (RM *RMServer) Put(ctx context.Context, hashUpt *hashtable.PutRequest) (*hashtable.PutResponse, error){
	<- RM.lockChannel 
	time.Sleep(5 * time.Second)
	RM.hashTableCopy[hashUpt.Key] = hashUpt.Value
	hashtableUpdateOutcome := &hashtable.PutResponse{
		Success: true,
	}
	RM.lockChannel <- true
	return hashtableUpdateOutcome, nil
}


func (RM *RMServer) Get(ctx context.Context, getRqst *hashtable.GetRequest) (*hashtable.GetResponse, error) {
	hashValue := RM.hashTableCopy[getRqst.Key]
	getResp := &hashtable.GetResponse{
		Value:  hashValue,
	}
	return getResp, nil
}

func setLogRMServer() *os.File {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}