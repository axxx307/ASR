package main

import (
	asr "ASR/functions"
	"flag"
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
	}
	if args[1] == "lookup" {
		asr.LookUp(filename, session)
	}
}
