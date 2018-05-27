package asr

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/satori/go.uuid"

	"github.com/mpiannucci/peakdetect"
	"github.com/r9y9/gossp"
	"github.com/r9y9/gossp/io"
	"github.com/r9y9/gossp/stft"
	"github.com/r9y9/gossp/window"
	mgo "gopkg.in/mgo.v2"
)

//MinAmpLimit minimum threshold for a frequency to be registered
const MinAmpLimit = 300

//MaxAmpLimit minimum threshold for a frequency to be registered
const MaxAmpLimit = 2000

//SortPair is used sorting songs appearence by value
type SortPair struct {
	key   string
	value int
}

//PairList A slice of pairs that implements sort.Interface to sort by values
type PairList []SortPair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool { return p[i].value < p[j].value }

//Analyze song into multiple blocks of subfingerprints
func Analyze(file *string, session *mgo.Session) {
	spectorgram := createSpectrogram(file)
	peaks := processPeaks(spectorgram)

	//remove frequencies below threshold
	for index, peak := range peaks {
		k := 0
		subFingerprint := make([]float64, len(peak))
		for _, value := range peak {
			if value/100 >= MinAmpLimit && value/100 <= MaxAmpLimit {
				subFingerprint[k] = value
				k++
			}
		}
		peaks[index] = subFingerprint

	}

	hashes := make([]string, len(peaks))
	for index, peak := range peaks {
		hashes[index] = generateHashes(&peak)
	}
	fingerprintIDs := writeFingerPrintsToDB(&hashes, session)
	song := &Song{Name: *file, Duration: "0", FingerprintIDs: &fingerprintIDs}
	if error := WriteSong(song, session); error != nil {
		log.Fatal(error)
	}
}

//LookUp searches song by generated hashes
func LookUp(file *string, session *mgo.Session) string {
	spectorgram := createSpectrogram(file)
	peaks := processPeaks(spectorgram)

	//remove frequencies below threshold
	for index, peak := range peaks {
		k := 0
		subFingerprint := make([]float64, len(peak))
		for _, value := range peak {
			if value/100 >= MinAmpLimit && value/100 <= MaxAmpLimit {
				subFingerprint[k] = value
				k++
			}
		}
		peaks[index] = subFingerprint

	}

	hashes := make([]string, len(peaks))
	for index, peak := range peaks {
		hashes[index] = generateHashes(&peak)
	}

	//find fingerprint blocks where at least one of the subfingerprints match in database
	hashMap := make(map[string]bool)
	for _, hash := range hashes {
		if fingerprintID := SearchSongBySubFingerprint(&hash, session); fingerprintID != "" {
			if _, exists := hashMap[fingerprintID]; !exists {
				hashMap[fingerprintID] = true
			}
		}
	}

	//retreive all songs and number of times tey were found by fingerprint
	songs := make(map[string]int)
	for fingerprint := range hashMap {
		song := SearchSongByFingerprint(&fingerprint, session)
		if _, exists := songs[song.Name]; !exists && song.Name != "" {
			songs[song.Name] = 1
		} else {
			songs[song.Name]++
		}
	}

	//sort songs in descending order and return one with more of block hit
	index := 0
	result := make(PairList, len(songs))
	for key, value := range songs {
		result[index] = SortPair{key, value}
		index++
	}

	sort.Sort(sort.Reverse(result))
	return result[0].key
}

//SearchExistingSong - search song by name in case we try to run analysis on it again
func SearchExistingSong(name *string, session *mgo.Session) *Song {
	return SearchExistingSongInDb(name, session)
}

//Create spectrogram for wav file
func createSpectrogram(file *string) [][]float64 {
	wav, werr := io.ReadWav(*file)
	if werr != nil {
		log.Fatal(werr)
	}

	monoData := wav.GetMonoData()

	spgramConfig := &stft.STFT{
		FrameShift: int(float64(wav.SampleRate) / 100.0), // 0.01 sec,
		FrameLen:   2048,
		Window:     window.CreateHanning(2048),
	}

	spectrogram, _ := gossp.SplitSpectrogram(spgramConfig.STFT(monoData))
	return spectrogram
}

func processPeaks(matrix [][]float64) [][]float64 {
	peaks := make([][]float64, len(matrix))
	for frame, vec := range matrix {
		_, _, _, maxv := peakdetect.PeakDetect(vec[:], 1.0)
		sort.Float64s(maxv)
		peaks[frame] = maxv
	}
	return peaks
}

func generateHashes(localMax *[]float64) string {
	hash := sha1.New()
	hashStr := ""
	hashStr += strings.Trim(strings.Join(strings.Fields(fmt.Sprint(*localMax)), "|"), "[]")
	hash.Write([]byte(hashStr))
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

func writeFingerPrintsToDB(hashes *[]string, session *mgo.Session) []string {
	fingerpintGUID, _ := uuid.NewV4()
	fingerprintIDs := make([]string, len(*hashes)/256+1)
	fingerprintIDs[0] = fingerpintGUID.String()
	for index, hash := range *hashes {
		guid, _ := uuid.NewV4()
		fingerprint := &SubFingerprint{SubFingerPrintID: guid.String()}
		fingerprint.FingerPrintID = fingerpintGUID.String()
		//set new fingerprint block
		if index%256 == 0 && index != 0 {
			fingerpintGUID, _ = uuid.NewV4()
			fingerprintIDs = append(fingerprintIDs, fingerpintGUID.String())
		}
		fingerprint.Hash = hash
		if error := WriteSubFingerprint(fingerprint, session); error != nil {
			log.Fatal(error)
		}
	}
	return fingerprintIDs
}
