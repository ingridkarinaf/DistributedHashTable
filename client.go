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
)

func main() {

	//Creating log file
	f := setLogClient()
	defer f.Close()

	//Creating 
	port := ":" + os.Args[1] 
	connection, err := grpc.Dial(port, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to connect: %v", err)
	}

	server := hashtable.NewHashTableClient(connection) //creates a new client
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
			// hashtableUpdate := &hashtable.putRequest{
			// 	key:   int32(key),
			// 	value: int32(value),
			// }

			fmt.Println("hashtableupdate: ", hashtable)
			fmt.Println("server: ", server, key, value)
		}
	}()

	for {

	}
}


func setLogClient() *os.File {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}