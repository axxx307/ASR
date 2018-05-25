package main

import (
	"flag"

	mgo "gopkg.in/mgo.v2"
)

func main() {
	filename := flag.String("i", "sample.wav", "Input filename")
	flag.Parse()

	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	Analyze(filename, session)
}

///CURRENTLY - EACH OF THE FRAME IS ROUGHLY A 10 MILISECONDS OF THE SONG
///EACH FRAME CONSISTS OF MULTIPLE PEAKS, THAT SHOULD BE 300 - 2000 HZ ONLY
///FINGERPRINT CONSISTS OF 256 SU-FINGERPRINT BLOCKS. SINGLE FINGERPRINT IS ENOUGHT TO
///IDENTIFY A SONG. WE CAN SEARCH DB BY ONLY ONE OF THE SUB-BLOCKS.
///EACH SUB-BLOCK SHOULD BE A 32BIT VALUE; WE HAVE A HASH FUNCTION INSTEAD
