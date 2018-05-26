# ASR

ASR - short for Automatic Speech Recognition. In a current state supports adding/retreiving song from mongodb by using wave parsing and spectrogram analysis. In a future work recognition from micropthone input will be added with support for recognizing human speech on-the-fly. 

## Getting Started

Clone the project and run ``` go run program <mode> <filename>``` where mode: 1) analysis - for writing songs fingerprints into db 2) lookup - to find song by using file

## Acknowledgments

The materials that were used to implement this idea are:

* Will Drevo and his great material at http://willdrevo.com/fingerprinting-and-audio-recognition-with-python/
* Jaap Haitsma and Ton Kalker: A Highly Robust Audio Fingerprinting System
