package main

import (
	"crypto/sha1"
	"encoding/base64"
	"flag"
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

func main() {
	filename := flag.String("i", "sample.wav", "Input filename")
	flag.Parse()

	wav, werr := io.ReadWav(*filename)
	if werr != nil {
		log.Fatal(werr)
	}
	//file bits
	monoData := wav.GetMonoData()

	spgramConfig := &stft.STFT{
		FrameShift: int(float64(wav.SampleRate) / 100.0), // 0.01 sec,
		FrameLen:   2048,
		Window:     window.CreateHanning(2048),
	}

	spectrogram, _ := gossp.SplitSpectrogram(spgramConfig.STFT(monoData))
	fingerprint := generateFingerPrints(spectrogram)
	for _, v := range fingerprint {
		println(v)
	}
}

///CURRENTLY - EACH OF THE FRAME IS ROUGLY A 10 MILISECONDS OF THE SONG
///EACH FRAME CONSISTS OF MULTIPLE PEAKS, THAT SHOULD BE 300 - 2000 HZ ONLY
///FINGERPRINT CONSISTS OF 256 SU-FINGERPRINT BLOCKS. SINGLE FINGERPRINT IS ENOUGHT TO
///IDENTIFY A SONG. WE CAN SEARCH DB BY ONLY ONE OF THE SUB-BLOCKS.
///EACH SUB-BLOCK SHOULD BE A 32BIT VALUE; WE HAVE A HASH FUNCTION INSTEAD

//PrintMatrixAsGnuplotFormat s
func generateFingerPrints(matrix [][]float64) []string {
	fmt.Println(len(matrix))
	frameToFreq := make([]string, len(matrix))
	for frame, vec := range matrix {
		_, _, _, maxv := peakdetect.PeakDetect(vec[:], 1.0)
		sort.Float64s(maxv)
		frameToFreq[frame] = generateHashes(&frame, &maxv)
	}
	return frameToFreq
}
func generateHashes(frame *int, localMax *[]float64) string {
	hash := sha1.New()
	hashStr := fmt.Sprintf("%v|", *frame)
	hashStr += strings.Trim(strings.Join(strings.Fields(fmt.Sprint(*localMax)), "|"), "[]")
	hash.Write([]byte(hashStr))
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}
