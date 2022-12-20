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
	"reflect"
)

/* 
Responsible for:
	1. Making connection to FE, has to redial if they lose connection
	2. 

Limitations:
	1. Can only dial to pre-determined front-ends or by incrementing 
	(then there is no guarantee that there is an FE with that port number)
	2. Assumes a failed request is due to a crashed server, redials immediately
*/

var server hashtable.HashTableClient
var connection *grpc.ClientConn 

func main() {

	//Creating log file
	f := setLogClient()
	defer f.Close()

	FEport := ":" + os.Args[1] //dial 4000 or 4001 (available ports on the FEServers)
	conn, err := grpc.Dial(FEport, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to connect: %v", err)
	}
	connection = conn

	server = hashtable.NewHashTableClient(connection) //creates a connection with an FE server
	defer connection.Close()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		
		for {
			println("Enter 'update' to update hash table, 'get' to get value of a given key. (without quotation marks)")
			scanner.Scan()
			textChoice := scanner.Text()
			if (textChoice == "update") {
				println("Enter key and value, separated by a space (integers only!)")
				scanner.Scan()
				text := scanner.Text()
				inputArray := strings.Fields(text)
				
				key, err := strconv.Atoi(inputArray[0])
				if err != nil {
					fmt.Println("Couldn't convert key to int: ", err)
					//log.Fatalf("Couldn't convert key to int: ", err)
				}
				value, err := strconv.Atoi(inputArray[1])
				if err != nil {
					fmt.Println("Couldn't convert key to int: ", err)
					//log.Fatalf("Couldn't convert value to int: ", err)
				}
	
				hashtableUpdate := &hashtable.PutRequest{
					Key:   int32(key),
					Value: int32(value),
				}

				result := Put(hashtableUpdate)
				if result.Success == true {
					fmt.Printf("Hashtable successfully updated to %v for key %v.\n", key, value)
				} else {
					fmt.Println("Update unsuccessful, please try again.")
				}
				
			} else if (textChoice == "get") {
				println("Enter the key of the value you would like to retireve (integers only!): ")
				scanner.Scan()
				text := scanner.Text()
				key, err := strconv.Atoi(text)
				if err != nil {
					fmt.Println("Could not convert key to integer: ", err)
					//log.Fatalf("Could not convert key to integer: ", err)
				}
				getReq := &hashtable.GetRequest{
					Key:   int32(key),
				}

				result := Get(getReq) //result := Get(getReq, connection, server)
				fmt.Printf("Value of key %s: %v \n",text, int(result))
			} else {
				fmt.Println("Sorry, didn't catch that. ")
			}
		}
	}()

	for {}
}

//If function returns an error, redial to other front-end and try again
func Put(hashUpt *hashtable.PutRequest) (*hashtable.PutResponse) {
	result, err := server.Put(context.Background(), hashUpt) //What does the context.background do?
	if err != nil {
		fmt.Printf("Client %s hashUpdate failed:%s. \n Redialing and retrying. \n", connection.Target(), err)
		Redial()
		return Put(hashUpt)
	}
	return result
}

func Get(getRsqt *hashtable.GetRequest) (int32) {
	result, err := server.Get(context.Background(), getRsqt)
	if err != nil {
		fmt.Printf("Client %s get request failed: %s", connection.Target(), err)
		Redial()
		return Get(getRsqt)
	}

	if reflect.ValueOf(result.Value).Kind() != reflect.ValueOf(int32(5)).Kind() {
		return 0
	}
	return result.Value
}

//In the case of losing connection - alternates between predefined front-
func Redial() {
	var port string
	if connection.Target()[len(connection.Target())-1:] == "1" {
		port =  ":4000"
	} else {
		port = ":4001"
	}

	conn, err := grpc.Dial(port, grpc.WithInsecure())
	if err != nil {
		fmt.Println("Unable to connect to port %s: %v", port, err)
	}

	connection = conn
	server = hashtable.NewHashTableClient(connection) //creates a connection with an FE server
}


func setLogClient() *os.File {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}