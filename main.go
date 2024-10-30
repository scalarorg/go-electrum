package main

import (
	"github/scalar.org/go-electrum/config"
	"log"
	"net/rpc"
)

type Args struct {
}

func main() {
	// Load environment variables into viper
	if err := config.LoadEnv(); err != nil {
		panic("Failed to load environment variables: " + err.Error())
	}

	var reply int64
	args := Args{}
	client, err := rpc.DialHTTP("tcp", "localhost"+":60001")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	err = client.Call("TimeServer.GiveServerTime", args, &reply)
	if err != nil {
		log.Fatal("arith error:", err)
	}
	log.Printf("%d", reply)
}
