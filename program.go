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

	switch args[1] {
	case "analyze":
		// search existing song
		asr.Init(asr.Analyze)
		if song := asr.SearchExistingSong(filename, session); song != nil {
			fmt.Printf("Song %s already exists \n", song.Name)
			return
		}
		asr.AnalyzeInput(filename, session)
		fmt.Println("Analysis complete")
	case "lookup":
		asr.Init(asr.Lookup)
		song := asr.LookUp(filename, session)
		fmt.Printf("Song is - %s \n", song)
	case "read-lookup":
		asr.Init(asr.Lookup)
		song := asr.LookUp(filename, session)
		fmt.Printf("Song is - %s \n", song)
	}
}
