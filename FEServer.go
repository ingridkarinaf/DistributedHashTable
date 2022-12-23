package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
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
	hashtable.UnimplementedHashTableServer
	port                    string 
	ctx                     context.Context
	replicaManagers 		map[string]hashtable.HashTableClient
}

var serverToDial int

func main() {
	f := setLogFEServer()
	defer f.Close()

	port := os.Args[1] 
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
		ctx:           ctx,
		replicaManagers: make(map[string]hashtable.HashTableClient),
	}
	hashtable.RegisterHashTableServer(grpcServer, server) 
	log.Printf("FEServer %s: Server on port %s: Listening at %v\n", server.port, port, list.Addr())

	
	go func() {
		log.Printf("FEServer _attempting_ to listen on port %s \n", server.port)
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("Failed to serve on port %s: %v", server.port, err)
		}

		log.Printf("FEServer %s successfully listening for requests.", server.port)
	}()

	//Dialing to ports 5000, 5001 and 5002
	for i := 0; i < 3; i++ {
		port := 5000 + i 
		address := ":" + strconv.Itoa(port)
		conn := server.DialToServer(address)
		defer conn.Close()
	}

	for {}
}

func (FE *FEServer) DialToServer(port string) (*grpc.ClientConn) {
	log.Printf("FE server %v: Trying to dial RM server with port: %v\n", FE.port, port)
	conn, err := grpc.Dial(port, grpc.WithInsecure(), grpc.WithBlock()) //This is going to wait until it receives the connection
	if err != nil { //Reconsider error handling
		log.Fatalf("FEServer %s could not connect to RM server with port: %s", FE.port, port, err)
	}
	
	c := hashtable.NewHashTableClient(conn)
	FE.replicaManagers[port] = c
	return conn
}

//Waits only for two success responses, chucks out the last one (for performance, only a bonus if the last one is successful)
func (FE *FEServer) Put(ctx context.Context, hashUpt *hashtable.PutRequest) (*hashtable.PutResponse, error){
	
	resultChannel := make(chan bool, 2)
	for port, RMconnection := range FE.replicaManagers  {
		
		//Sending update request to all replica managers simultaneously
		go func(rmPort string, connection hashtable.HashTableClient) {
			_, err := connection.Put(context.Background(), hashUpt) //does context.background make it async?
			if err != nil {
				log.Printf("FE Server: Hash table to RM server update failed for FE server %s: %s", rmPort, FE.port, err) //identify which replica server?
				delete(FE.replicaManagers, rmPort)
			} else {
				resultChannel <- true
			}
		}(port, RMconnection)
	}
	
	// Should wait until received two values
	<-resultChannel
	<-resultChannel 

	hashtableUpdateOutcome := &hashtable.PutResponse{
		Success: true,
	}
	return hashtableUpdateOutcome, nil
}

func (FE *FEServer) Get(ctx context.Context, getRsqt *hashtable.GetRequest) (*hashtable.GetResponse, error) {
	responseChannel := make(chan *hashtable.GetResponse, 2)
	for port, RMconnection := range FE.replicaManagers  {
		go func(rmPort string, connection hashtable.HashTableClient) {
			result, err := connection.Get(context.Background(), getRsqt) 
			if err != nil {
				log.Printf("Hash table update to RM server %s failed in FE server %s: %s", rmPort, FE.port, err)
				delete(FE.replicaManagers, rmPort)
			} else {
				responseChannel <- result
			}
		}(port, RMconnection)
	}

	resp1 := <-responseChannel
	resp2 :=  <-responseChannel

	if resp1.Value != resp2.Value {
		resp3 := <-responseChannel
		return resp3, nil
	}
	return resp1, nil
}

func setLogFEServer() *os.File {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}