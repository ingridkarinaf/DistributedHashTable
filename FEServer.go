package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"fmt"
	// "sync"
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
	replicaManagers 		map[string]hashtable.HashTableClient
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
		fmt.Printf("FEServer failed to listen on port %s: %v", address, err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	grpcServer := grpc.NewServer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := &FEServer{
		port:          os.Args[1],
		primaryServer: nil,
		ctx:           ctx,
		replicaManagers: make(map[string]hashtable.HashTableClient),
	}
	hashtable.RegisterHashTableServer(grpcServer, server) 
	fmt.Printf("FEServer %s: Server on port %s: Listening at %v\n", server.port, port, list.Addr())

	
	go func() {
		fmt.Printf("FEServer _attempting_ to listen on port %s \n", server.port)
		if err := grpcServer.Serve(list); err != nil {
			// fmt.Println("error when trying to listen")
			// fmt.Printf("failed to serve %v", err)
			log.Fatalf("failed to serve %v", err)
		}

		fmt.Printf("FEServer %s successfully listening for requests.", server.port)
	}()

	//Dialing to ports 5000, 5001 and 5002
	for i := 0; i < 3; i++ {
		port := 5000 + i 
		address := ":" + strconv.Itoa(port)
		fmt.Printf("port: %v, address: %s", port, address)
		conn := server.DialToServer(address)
		defer conn.Close()
	}

	for {}
}

func (FE *FEServer) DialToServer(port string) (*grpc.ClientConn) {
	fmt.Println("port inside dial to server: ", port)
	//Add connection to map
	// var conn *grpc.ClientConn
	log.Printf("RMServer %v: Trying to dial: %v\n", port)
	conn, err := grpc.Dial(port, grpc.WithInsecure(), grpc.WithBlock()) //This is going to wait until it receives the connection
	if err != nil { // //Reconsider error handling
		log.Fatalf("FEServer could not connect: %s", err)
	}
	
	c := hashtable.NewHashTableClient(conn)
	FE.replicaManagers[port] = c
	return conn
}

//Waits only for two success responses, chucks out the last one (for performance, only a bonus if the last one is successful)
func (FE *FEServer) Put(ctx context.Context, hashUpt *hashtable.PutRequest) (*hashtable.PutResponse, error){

	//var wg sync.WaitGroup
	resultChannel := make(chan bool, 2)
	for _, RMconnection := range FE.replicaManagers  {
		//wg.Add(1)
		fmt.Println("rm connection: ", RMconnection)
		
		go func(connection hashtable.HashTableClient) {
			//Reconsider if something should be done with return value 
			_, err := connection.Put(context.Background(), hashUpt) //does context.background make it async?
			if err != nil {
				fmt.Printf("Map update failed: %s", err) //identify which replica server?
			} else {
				resultChannel <- true
			}
		}(RMconnection)
	}
	//wg.Wait()
	resp1, resp2 := <-resultChannel, <-resultChannel // Should wait until received two values
	fmt.Println("Responses: ", resp1, resp2)

	hashtableUpdateOutcome := &hashtable.PutResponse{
		Success: true,
	}
	return hashtableUpdateOutcome, nil
}

func (FE *FEServer) Get(ctx context.Context, getRsqt *hashtable.GetRequest) (*hashtable.GetResponse, error) {

	responseChannel := make(chan *hashtable.GetResponse, 2)
	for _, RMconnection := range FE.replicaManagers  {
		fmt.Println("RM connection: ", RMconnection)
		go func(connection hashtable.HashTableClient) {
			fmt.Println("RM connection inside thread: ", connection)
			result, err := connection.Get(context.Background(), getRsqt) 
			if err != nil {
				fmt.Printf("Map update failed: %s", err)
			} else {
				fmt.Println("result inside get before passing to channel: ", result )
				responseChannel <- result
			}
		}(RMconnection)
	}

	resp1, resp2 := <-responseChannel, <-responseChannel
	fmt.Println("responses inside get of FE server: ", resp1, resp2)
	if resp1 != resp2 {
		resp3 := <-responseChannel
		fmt.Println("responses inside get of FE server: if statement where resp1 and 2 arent equal ", resp1, resp2, resp3)
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