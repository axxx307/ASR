package asr

import (
	"log"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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

func WriteSubFingerprint(subprint *SubFingerprint, session *mgo.Session) error {
	c := session.DB("ASR").C("fingerprints")
	return c.Insert(subprint)
}

func WriteSong(song *Song, session *mgo.Session) error {
	c := session.DB("ASR").C("songs")
	return c.Insert(song)
}

func SearchSong(hash *string, session *mgo.Session) string {
	c := session.DB("ASR").C("fingerprints")
	result := &SubFingerprint{Hash: ""}
	err := c.Find(bson.M{"hash": hash}).One(result)
	if err != nil {
		log.Fatal(err)
	}
	if result.Hash == "" {
		return ""
	}
	return result.FingerPrintID
}
