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

//Analyze sf
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
func LookUp(file *string, session *mgo.Session) Song {
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
	for _, hash := range hashes {
		if fingerprintID := SearchSong(&hash, session); fingerprintID != "" {
			println(fingerprintID)
			break
		}
	}
	return Song{}
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
	fmt.Println(len(matrix))
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
