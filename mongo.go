package main

import (
	mgo "gopkg.in/mgo.v2"
)

func setConntedtion() {
	session, err := mgo.Dial("server1.example.com,server2.example.com")
	if err != nil {
		panic(err)
	}
	defer session.Close()
}

//SubFingerprint is a par of the 256 blocks that create a single fingerprint
//to identify a song
type SubFingerprint struct {
	FingerPrintID    string
	SubFingerPrintID string
	Hash             string
}

//Song ...
type Song struct {
	Name           string
	Duration       string
	FingerprintIDs *[]string
}

func writeSubFingerprint(subprint *SubFingerprint, session *mgo.Session) error {
	c := session.DB("ASR").C("fingerprints")
	return c.Insert(subprint)
}

func writeSong(song *Song, session *mgo.Session) error {
	c := session.DB("ASR").C("songs")
	return c.Insert(song)
}

func searchSong(song *Song, session *mgo.Session) error {
	c := session.DB("ASR").C("songs")
	return c.Insert(song)
}
