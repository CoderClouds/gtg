/**
  create by yy on 2019-10-21
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	currentPath, _ := os.Getwd()
	//flag.Usage = Usage
	message := flag.String("s", "default s message", "it's user send message[help message]")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1{
		fmt.Println("no flag")
	}else {
		fmt.Println("args: ", args)
	}

	log.Println("message:", *message, "current path", currentPath)
}
