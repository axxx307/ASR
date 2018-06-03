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
	BlockPosition    uint16
}

//Song ...
type Song struct {
	Name           string
	Duration       string
	FingerprintIDs *[]string
}

//WriteSubFingerprint - add subfingerprint block into db
func WriteSubFingerprint(subprint *SubFingerprint, session *mgo.Session) error {
	c := session.DB("ASR").C("fingerprints")
	return c.Insert(subprint)
}

//WriteSong - add song into db
func WriteSong(song *Song, session *mgo.Session) error {
	c := session.DB("ASR").C("songs")
	return c.Insert(song)
}

//SearchSongBySubFingerprint - search all fingerprint blocks by subfingerprint hash
func SearchSongBySubFingerprint(hash *string, session *mgo.Session) string {
	c := session.DB("ASR").C("fingerprints")
	result := &SubFingerprint{Hash: ""}
	c.Find(bson.M{"hash": hash}).One(result)
	if result.Hash == "" {
		return ""
	}
	return result.FingerPrintID
}

//SearchSongByFingerprint - search song by fingerprint block
func SearchSongByFingerprint(hash *string, session *mgo.Session) *Song {
	c := session.DB("ASR").C("songs")
	result := &Song{}
	err := c.Find(bson.M{"fingerprintids": hash}).One(result)
	if err != nil {
		log.Fatal(err)
	}
	if result == nil {
		return nil
	}
	return result
}

//SearchExistingSongInDb - search song by name in case we try to run analysis on it again
func SearchExistingSongInDb(name *string, session *mgo.Session) *Song {
	c := session.DB("ASR").C("songs")
	result := &Song{}
	err := c.Find(bson.M{"name": name}).One(result)
	if err != nil {
		return nil
	}
	return result
}
