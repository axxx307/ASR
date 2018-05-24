package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/mpiannucci/peakdetect"
	"github.com/r9y9/gossp"
	"github.com/r9y9/gossp/io"
	"github.com/r9y9/gossp/stft"
	"github.com/r9y9/gossp/window"
)

//MinAmpLimit minimum threshold for a frequency to be registered
const MinAmpLimit = 300

//MaxAmpLimit minimum threshold for a frequency to be registered
const MaxAmpLimit = 2000

//Analyze sf
func Analyze(file *string) {
	spectorgram := createSpectrogram(file)
	peaks := processPeaks(spectorgram)

	//remove frequencies below threshold
	for _, peak := range peaks {
		k := 0
		for _, value := range peak {
			if value >= MinAmpLimit && value <= MaxAmpLimit {
				peak[k] = value
				k++
			}
			peak = peak[:]
		}
	}
	for _, peak := range peaks {
		hash := generateHashes(&peak)
		println(hash)
	}
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
