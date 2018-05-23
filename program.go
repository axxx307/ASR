package main

import (
	"flag"
	"fmt"
	"log"

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
	fingerprints := findPeaksInFrame(spectrogram)
	fmt.Println(fingerprints)
}

//PrintMatrixAsGnuplotFormat s
func findPeaksInFrame(matrix [][]float64) [][]float64 {
	fmt.Println(len(matrix))
	frameToFreq := make([][]float64, len(matrix))
	for frame, vec := range matrix {
		_, _, _, maxv := peakdetect.PeakDetect(vec[:], 1.0)
		frameToFreq[frame] = maxv
	}
	return frameToFreq
}
