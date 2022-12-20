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
*/

var i32 interface{} = int32(5)

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
		
		for {
			println("Enter 'update' to update hash table, 'get' to get value of a given key. (without quotation marks)")
			scanner.Scan()
			textChoice := scanner.Text()
			if (textChoice == "update") {
				println("Enter key and value, separated by a space (integers only!)")
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
				fmt.Println("Put result: ", result)
			} else if (textChoice == "get") {
				println("Enter the key of the value you would like to retireve (integers only!): ")
				scanner.Scan()
				text := scanner.Text()
				val, err := strconv.Atoi(text)
				if err != nil {
					fmt.Println("Could not convert key to integer: ", err)
				}
				getReq := &hashtable.GetRequest{
					Key:   int32(val),
				}

				result := Get(getReq, connection, server)
				fmt.Println("Get result: ", result)
			} else {
				fmt.Println("Sorry, didn't catch that. ")
			}
			
		}
	}()

	for {

	}
}

//If function returns an error, redial to other front-end and try again
func Put(hashUpt *hashtable.PutRequest, connection *grpc.ClientConn, server hashtable.HashTableClient) (*hashtable.PutResponse) {
	result, err := server.Put(context.Background(), hashUpt) //What does the context.background do?
	if err != nil {
		fmt.Printf("Client %s hashUpdate failed:%s", connection.Target(), err)
	}
	return result
}

func Get(getRsqt *hashtable.GetRequest, connection *grpc.ClientConn, server hashtable.HashTableClient) (int32) {
	result, err := server.Get(context.Background(), getRsqt)
	if err != nil {
		fmt.Printf("Client %s get request failed: %s", connection.Target(), err)
		Redial()
	}

	if reflect.ValueOf(result.Value).Kind() !=  reflect.ValueOf(int32(5)).Kind() {
		fmt.Println(reflect.ValueOf(int32(5)).Kind())
		return 0
	}
	fmt.Println("result inside get of client: ", result, reflect.TypeOf(result))
	return result.Value
}

//In the case of losing connection - alternates between predefined front-
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