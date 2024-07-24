package main

import (
	"flag" // helps us to get value from the command line
	"log"
)

func init() {
	log.SetPrefix("Blockchain: ")
}

func main() {
	/* Defining the command line flag 
	port(name of the options we will use) - this command-line option we can set when we run our program 
	5000 - defaul value to use if no value is provided
	short description of what this option does - TCP port number for blockchain server
	*/
	port := flag.Uint("port", 5000, "TCP port number for blockchain server")
	flag.Parse()
	// tells the program to look at the command line and finc any options we've set 
	app := NewBlockChainServer(uint16(*port))
	app.Run()
}