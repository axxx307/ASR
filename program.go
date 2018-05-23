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
	data := wav.GetMonoData()

	s := &stft.STFT{
		FrameShift: int(float64(wav.SampleRate) / 100.0), // 0.01 sec,
		FrameLen:   2048,
		Window:     window.CreateHanning(2048),
	}

	spectrogram, _ := gossp.SplitSpectrogram(s.STFT(data))
	PrintMatrixAsGnuplotFormat(spectrogram)
}

//PrintMatrixAsGnuplotFormat s
func PrintMatrixAsGnuplotFormat(matrix [][]float64) {
	fmt.Println("#", len(matrix[0]), len(matrix)/2)
	for _, vec := range matrix {
		_, _, _, maxv := peakdetect.PeakDetect(vec[:], 1.0)
		fmt.Println(maxv)
		fmt.Println("")
	}
}
