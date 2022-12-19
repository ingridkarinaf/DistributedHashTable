package main

import (
	hashtable "github.com/ingridkarinaf/DistributedHashTable/interface"
	grpc "google.golang.org/grpc"
	"strings"
	"strconv"
	"log"
	"os"
	"bufio"
	"fmt"
	"context"
)

/* 
Responsible for:
	1. Making connection to FE, has to redial if they lose connection
	2. 

Limitations:
	1. Can only dial to pre-determined front-ends or by incrementing 
	(then there is no guarantee that there is an FE with that port number)
*/

func main() {

	//Creating log file
	f := setLogClient()
	defer f.Close()

	//clientPort := ":" + os.Args[2]
	FEport := ":" + os.Args[2] //dial 4000 or 4001 (available ports on the FEServers)
	connection, err := grpc.Dial(FEport, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to connect: %v", err)
	}

	server := hashtable.NewHashTableClient(connection) //creates a connection with an FE server
	defer connection.Close()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		println("Enter key and value, separated by a space")

		for {
			scanner.Scan()
			text := scanner.Text()
			inputArray := strings.Fields(text)
			key, err := strconv.Atoi(inputArray[0])
			if err != nil {
				log.Fatalf("Couldn't convert key to int: ", err)
			}
			value, err := strconv.Atoi(inputArray[1])
			if err != nil {
				log.Fatalf("Couldn't convert value to int: ", err)
			}

			//Send request to update hashtable
			hashtableUpdate := &hashtable.PutRequest{
				Key:   int32(key),
				Value: int32(value),
			}

			fmt.Println(hashtableUpdate)

			result := Put(hashtableUpdate, connection, server)
			fmt.Println("result: ", result)
			// log.Printf("Client %s: Bid response: ", port, ack.GetAcknowledgement())
			// println("Bid response: ", ack.GetAcknowledgement())
			
		}
	}()

	for {

	}
}

func Put(hashUpt *hashtable.PutRequest, connection *grpc.ClientConn, server hashtable.HashTableClient) (*hashtable.PutResponse) {
	result, err := server.Put(context.Background(), hashUpt) //What does the context.background do?
	if err != nil {
		fmt.Printf("Client %s: update failed:%s", connection.Target(), err)
	}
	return result
}

//In the case of losing connection
func Redial() {

}


func setLogClient() *os.File {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}