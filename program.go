package main

import (
	asr "ASR/functions"
	"flag"
	"fmt"
	"os"

	mgo "gopkg.in/mgo.v2"
)

func main() {
	args := os.Args
	filename := flag.String("i", args[2], "Input filename")
	flag.Parse()

	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	if args[1] == "analyze" {
		asr.Analyze(filename, session)
		fmt.Println("Analysis complete")
	}
	if args[1] == "lookup" {
		song := asr.LookUp(filename, session)
		fmt.Printf("Song is - %s", song)
	}
}
