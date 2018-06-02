# ASR

ASR - short for Automatic Speech Recognition. In a current state supports adding/retreiving song from mongodb by using wave parsing and spectrogram analysis. Works with providing song file in .wav or .mp3 formats or via microphone.

## Getting Started

Clone the project and run ``` go run program <mode> <filename>``` where mode: 1) analysis - for writing songs fingerprints into db 2) lookup - to find song by using file

## Future work
* [ ] Improve finding speed 
* [ ] Add auto-tests
* [ ] Add support for voice recognition
* [ ] Add support for real-time recognition

## Acknowledgments

The materials that were used to implement this idea are:

* Will Drevo and his great material at http://willdrevo.com/fingerprinting-and-audio-recognition-with-python/
* Jaap Haitsma and Ton Kalker: A Highly Robust Audio Fingerprinting System
