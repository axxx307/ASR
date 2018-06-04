package main

import (
	music "ASR/fingerprint"
	"flag"
	"fmt"
	"os"

	mgo "gopkg.in/mgo.v2"
)

/// need to work on creating db from scratch with indexed hash songs
/// then see performance difference between faulty lookups vs all good ones
///

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
		music.Init(music.Analyze)
		if song := music.SearchExistingSong(filename, session); song != nil {
			fmt.Printf("Song %s already exists \n", song.Name)
			return
		}
		music.AnalyzeInput(filename, session)
		fmt.Println("Analysis complete")
	case "lookup":
		music.Init(music.Lookup)
		song := music.LookUp(filename, session)
		fmt.Printf("Song is - %s \n", song)
	case "lookup-mic":
		music.Init(music.LookupMic)
		song := music.LookUp(filename, session)
		fmt.Printf("Song is - %s \n", song)
	}
}
