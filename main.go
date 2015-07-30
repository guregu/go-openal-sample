package main

import (
	audio "azul3d.org/audio.v1"
	_ "azul3d.org/audio/wav.v1"
	al "azul3d.org/native/al.v1"
	"log"
	"math/rand"
	"os"
	"time"
	"unsafe"
)

const (
	freq = 44100 // 44.1 kHz
)

func init() {
	// al.SetErrorHandler(func(e error) {
	// 	panic(e)
	// })
}

func main() {
	var data []audio.PCM16
	var duration float64
	var config audio.Config
	switch len(os.Args) {
	case 1:
		log.Println("Loading white noise.")
		duration = 3
		data = genWhiteNoise(duration)
		config = audio.Config{SampleRate: freq, Channels: 1}
		break

	case 2:
		filename := os.Args[1]
		log.Printf("Loading %s\n", filename)
		data, duration, config = readFile(filename)
		data, duration, config = readFile(filename)
		break

	default:
		log.Panicf("Unexpected number of arguments: %d. Must be 1 or 2.\nUsage: go run main.go [file]\n", len(os.Args))
	}

	log.Println("Duration:", duration, "seconds")

	device, err := al.OpenDevice("", nil)
	if err != nil {
		log.Panic(err)
	}
	defer device.Close()

	var buffers uint32 = 0
	device.GenBuffers(1, &buffers)
	if config.Channels == 1 {
		device.BufferData(buffers, al.FORMAT_MONO16, unsafe.Pointer(&data[0]), int32(int(unsafe.Sizeof(data[0]))*len(data)), int32(config.SampleRate))
	} else {
		device.BufferData(buffers, al.FORMAT_STEREO16, unsafe.Pointer(&data[0]), int32(int(unsafe.Sizeof(data[0]))*len(data)), int32(config.SampleRate))
	}
	var sources uint32 = 0
	device.GenSources(1, &sources)
	device.Sourcei(sources, al.BUFFER, int32(buffers))
	device.SourcePlay(sources)

	time.Sleep(time.Duration(duration * float64(time.Second)))

	device.DeleteSources(1, &sources)
	device.DeleteBuffers(1, &buffers)
	log.Println("Done.")
}

func readFile(filename string) ([]audio.PCM16, float64, audio.Config) {
	file, err := os.Open(filename)
	if err != nil {
		log.Panic(err)
	}

	fi, err := file.Stat()
	if err != nil {
		log.Panic(err)
	}

	// Create a decoder for the audio source
	decoder, format, err := audio.NewDecoder(file)
	if err != nil {
		log.Panic("err: %T %v %#v", err, err, err)
	}
	config := decoder.Config()
	log.Printf("Decoding a %s file.\n", format)
	log.Println(config)

	time := float64(fi.Size()) / float64(config.SampleRate*config.Channels*16/8)

	// Create a buffer that can hold 3 second of audio samples
	bufSize := int(time * float64(config.SampleRate*config.Channels))
	// Most WAVs are PCM16
	buf := make(audio.PCM16Samples, bufSize)

	// Fill the buffer with as many audio samples as we can
	read, err := decoder.Read(buf)
	if err != nil && err != audio.EOS {
		log.Panic(err)
	}

	return []audio.PCM16(buf)[:read], time, config
}

func genWhiteNoise(duration float64) []audio.PCM16 {
	data := make([]audio.PCM16, int(float64(freq)*duration))
	for i := 0; i < len(data); i++ {
		data[i] = rnd(-32767, 32767)
	}
	return data
}

func rnd(min, max int) audio.PCM16 {
	return audio.PCM16(min + (rand.Intn(max - min)))
}
