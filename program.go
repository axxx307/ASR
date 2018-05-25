package main

import (
	asr "ASR/functions"
	"flag"
	"os"

	mgo "gopkg.in/mgo.v2"
)

func main() {
	args := os.Args

	filename := flag.String("i", args[1], "Input filename")
	flag.Parse()

	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	if args[0] == "analyze" {
		println("ss")
		asr.Analyze(filename, session)
	}
	if args[0] == "lookup" {
		asr.Analyze(filename, session)
	}
}

///CURRENTLY - EACH OF THE FRAME IS ROUGHLY A 10 MILISECONDS OF THE SONG
///EACH FRAME CONSISTS OF MULTIPLE PEAKS, THAT SHOULD BE 300 - 2000 HZ ONLY
///FINGERPRINT CONSISTS OF 256 SU-FINGERPRINT BLOCKS. SINGLE FINGERPRINT IS ENOUGHT TO
///IDENTIFY A SONG. WE CAN SEARCH DB BY ONLY ONE OF THE SUB-BLOCKS.
///EACH SUB-BLOCK SHOULD BE A 32BIT VALUE; WE HAVE A HASH FUNCTION INSTEAD
