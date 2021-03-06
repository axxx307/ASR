package music

import (
	mongo "ASR/mongodb"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/wav"
	"github.com/gordonklaus/portaudio"

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

//MaxFrameLengthThreshold max frames in a single song/audio. roughly a ~9 seconds
const MaxFrameLengthThreshold = 866

const (
	//Lookup mode
	Lookup ProgramMode = "lookup"
	//Analyze mode
	Analyze ProgramMode = "analyze"
	//LookupMic mode
	LookupMic ProgramMode = "lookup-mic"
)

//CurrentMode is a global value for current mode of the program
var CurrentMode ProgramMode

//ProgramMode is a type for an enum of programs mode: lookup, analyze ...
type ProgramMode string

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

//Init constructor
func Init(mode ProgramMode) {
	CurrentMode = mode
}

//AnalyzeInput song into multiple blocks of subfingerprints
func AnalyzeInput(file *string, session *mgo.Session) {
	monoData, sampleRate := readWavMonoData(file)
	spectorgram := createSpectrogram(&monoData, &sampleRate)
	peaks := processPeaks(spectorgram)

	//remove frequencies below threshold
	for index, peak := range peaks {
		subFingerprint := make(map[int][]float64)
		for _, value := range peak {
			if value/100 >= MinAmpLimit && value/100 <= MaxAmpLimit {
				subFingerprint[index] = append(subFingerprint[index], value)
			}
		}
		if len(subFingerprint) == 0 {
			delete(peaks, index)
			continue
		}
		peaks[index] = subFingerprint[index]

	}

	i := 0
	k := 5
	hashes := []string{}
	for _, peak := range peaks {
		var subPeak []float64
		for len(peak)-1 >= k+5 {
			subPeak = peak[i:k]
			hashes = append(hashes, generateHashes(&subPeak))
			i += 5
			k += 5
		}
		subPeak = peak[i:]
		hashes = append(hashes, generateHashes(&subPeak))
		i = 0
		k = 5
	}
	songID, _ := uuid.NewV4()
	song := &mongo.Song{Name: *file, Duration: "0", ID: songID.String()}
	fingerprintIDs := writeFingerPrintsToDB(&hashes, songID.String(), session)
	song.FingerprintIDs = &fingerprintIDs
	if error := mongo.WriteSong(song, session); error != nil {
		log.Fatal(error)
	}

	mongo.CreateIndex(session)
}

//LookUp searches song by generated hashes
func LookUp(file *string, session *mgo.Session) string {
	var monoData []float64
	var sampleRate uint32
	if CurrentMode == LookupMic {
		monoData, sampleRate = microphoneInput()
	} else {
		monoData, sampleRate = readWavMonoData(file)
	}

	spectorgram := createSpectrogram(&monoData, &sampleRate)
	peaks := processPeaks(spectorgram)

	//remove frequencies below threshold
	for index, peak := range peaks {
		subFingerprint := make(map[int][]float64)
		for _, value := range peak {
			if value/100 >= MinAmpLimit && value/100 <= MaxAmpLimit {
				subFingerprint[index] = append(subFingerprint[index], value)
			}
		}
		peaks[index] = subFingerprint[index]
	}

	i := 0
	k := 5
	hashes := []string{}
	for _, peak := range peaks {
		var subPeak []float64
		for len(peak)-1 >= k+5 {
			subPeak = peak[i:k]
			hashes = append(hashes, generateHashes(&subPeak))
			i += 5
			k += 5
		}
		subPeak = peak[i:]
		hashes = append(hashes, generateHashes(&subPeak))
		i = 0
		k = 5
	}

	//find fingerprint blocks where at least one of the subfingerprints match in database
	hashMap := make(map[string]bool)
	sem := make(chan struct{})
	var wait sync.WaitGroup
	wait.Add(len(hashes))
	for _, hash := range hashes {
		select {
		case sem <- struct{}{}:
			//pass waitgroup by reference to avoid deadlock
			go func(wt *sync.WaitGroup) {
				searchSongRoutine(&hash, &hashMap, session)
				<-sem
				wt.Done()
			}(&wait)
		default:
			searchSongRoutine(&hash, &hashMap, session)
			wait.Done()
		}
	}
	wait.Wait()

	if len(hashMap) == 0 {
		return "unknown"
	}
	//retreive all songs and number of times tey were found by fingerprint
	songs := make(map[string]int)
	for fingerprint := range hashMap {
		song := mongo.SearchSongByFingerprint(&fingerprint, session)
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
func SearchExistingSong(name *string, session *mgo.Session) *mongo.Song {
	return mongo.SearchExistingSongInDb(name, session)
}

func searchSongRoutine(hash *string, hashMap *map[string]bool, session *mgo.Session) {
	sessionCopy := session.Copy()
	defer sessionCopy.Close()
	fingerprintID := mongo.SearchSongBySubFingerprint(hash, sessionCopy)
	if fingerprintID != "" {
		(*hashMap)[fingerprintID] = true
	}
}

//MicrophoneInput read microphone input
func microphoneInput() ([]float64, uint32) {
	portaudio.Initialize()
	defer portaudio.Terminate()
	in := make([]int32, 512)
	nSamples := 0
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), in)
	chk(err)
	defer stream.Close()

	data := make([][]int32, MaxFrameLengthThreshold)
	fmt.Println("In 3 seconds the recording will start")
	time.Sleep(3 * time.Second)
	chk(stream.Start())
	fmt.Println("Recording started")
	for index := 0; index < MaxFrameLengthThreshold; index++ {
		chk(stream.Read())
		nSamples += len(in)
		data[index] = in
	}
	chk(stream.Stop())
	fmt.Println("Recording stopped")

	flData := []float64{}
	for index := 0; index < MaxFrameLengthThreshold; index++ {
		for _, value := range data[index] {
			flData = append(flData, float64(value))
		}
	}
	return flData, 44100
}

//Create spectrogram for wav file
func createSpectrogram(monoData *[]float64, sampleRate *uint32) [][]float64 {
	start := time.Now()
	spgramConfig := &stft.STFT{
		FrameShift: int(float64(*sampleRate) / 100.0), // 0.01 sec,
		FrameLen:   2048,
		Window:     window.CreateHanning(2048),
	}
	elapsed := time.Since(start)
	log.Printf("Hanning window search took %s", elapsed)

	//get short ft value and limit number of frames to MaxFrameLengthThreshold

	ft := spgramConfig.STFT(*monoData)
	if CurrentMode == Lookup && len(ft) > MaxFrameLengthThreshold {
		ft = ft[:MaxFrameLengthThreshold]
	}

	spectrogram, _ := gossp.SplitSpectrogram(ft)
	return spectrogram
}

func readWavMonoData(fileName *string) ([]float64, uint32) {
	if strings.Contains(*fileName, ".mp3") {
		fmt.Println("File is in mp3 format; converting to wav...")
		fileName = mp3ToWavConverter(fileName)
		fmt.Println("Finished conveting")
	}

	wav, werr := io.ReadWav(*fileName)
	if werr != nil {
		log.Fatal(werr)
	}

	return wav.GetMonoData(), wav.SampleRate

}

func processPeaks(matrix [][]float64) map[int][]float64 {
	k := 0
	peaks := make(map[int][]float64)
	for _, vec := range matrix {
		_, _, _, maxp := peakdetect.PeakDetect(vec[:], 1.0)
		if len(maxp) == 0 {
			continue
		}
		sort.Float64s(maxp)
		peaks[k] = maxp
		k++
	}
	return peaks
}

func generateHashes(localMax *[]float64) string {
	hash := sha1.New()
	hashStr := strings.Trim(strings.Replace(fmt.Sprint(*localMax), " ", "|", -1), "[]")
	hash.Write([]byte(hashStr))
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

func writeFingerPrintsToDB(hashes *[]string, songID string, session *mgo.Session) []string {
	fingerpintGUID, _ := uuid.NewV4()
	fingerprintIDs := []string{}
	fingerprintIDs = append(fingerprintIDs, fingerpintGUID.String())
	k := 0
	for _, hash := range *hashes {
		guid, _ := uuid.NewV4()
		fingerprint := &mongo.SubFingerprint{SubFingerPrintID: guid.String(), SongID: songID, Hash: hash}
		fingerprint.FingerPrintID = fingerpintGUID.String()
		//set new fingerprint block
		if k/256 == 1 {
			fingerpintGUID, _ = uuid.NewV4()
			fingerprintIDs = append(fingerprintIDs, fingerpintGUID.String())
			k = 0
		}
		if error := mongo.WriteSubFingerprint(fingerprint, session); error != nil {
			log.Fatal(error)
		}
		k++
	}
	return fingerprintIDs
}

func mp3ToWavConverter(fileName *string) *string {
	file, _ := os.Open(*fileName)

	wave, format, err := mp3.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	name := strings.Replace(*fileName, ".mp3", ".wav", 1)
	output, _ := os.Create(name)
	encodeErr := wav.Encode(output, wave, format)
	if encodeErr != nil {
		log.Fatal(encodeErr)
	}
	output.Close()
	return &name
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
