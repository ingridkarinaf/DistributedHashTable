package main

import (
	"context"
	"log"
	"net"
	"os"
	// "strconv"
	"fmt"

	hashtable "github.com/ingridkarinaf/DistributedHashTable/interface"
	grpc "google.golang.org/grpc"
)

/* 
Handles majority of the logic, i.e. 
	1. Replicating to servers
	2. Determining system model for node behaviour; in this case crash stop
	3. Determining failure handling and resiliance relating to servers; in this case resiliant to 1 node crash

Dials to pre-defined replica manager servers, i.e. 5000, 5001 and 5002
*/

type FEServer struct {
	hashtable.UnimplementedHashTableServer        // You need this line if you have are(?) a server
	port                    string // Not required but useful if your server needs to know what port it's listening to
	primaryServer           hashtable.HashTableClient
	ctx                     context.Context
	replicaManagers 		map[int32]hashtable.HashTableClient
}

var serverToDial int

func main() {
	f := setLogFEServer()
	defer f.Close()

	//Creating front-end server
	port := os.Args[1] //Port for the FEServer to listen on
	address := ":" + port
	list, err := net.Listen("tcp", address)
	if err != nil {
		log.Printf("FEServer failed to listen on port %s: %v", address, err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	grpcServer := grpc.NewServer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := &FEServer{
		port:          os.Args[1],
		primaryServer: nil,
		ctx:           ctx,
	}
	hashtable.RegisterHashTableServer(grpcServer, server) 
	fmt.Printf("FEServer %s: Server on port %s: Listening at %v\n", server.port, port, list.Addr())

	
	go func() {
		log.Printf("FEServer _attempting_ listening on port %s:", server.port)
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to serve %v", err)
		}

		log.Printf("FEServer %s successfully listening for requests.", server.port)
	}()

	for i := 0; i < 3; i++ {
		port := 5000 + i 
		address := ":" + port
		conn := server.DialToServer(serverToDial)
		defer conn.Close()

	}

	for {}

}

func (FE *FEServer) Put(ctx context.Context, hashUpt *hashtable.PutRequest) (*hashTable.PutResponse, error){
	fmt.Println("Inside of put method of FE server: ", FE)

	//Loop through list of servers 
	for RM in FE.replicaManagers {
		fmt.Println("RM: ", RM)
	}

	hashtableUpdateOutcome := &hashtable.PutResponse{
		Success: false
	}

	var err error 
	return hashtableUpdateOutcome, err
}

// sets the logger to use a log.txt file instead of the console
func setLogFEServer() *os.File {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}