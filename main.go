package main

import (
	audio "azul3d.org/audio.v1"
	_ "azul3d.org/audio/wav.v1"
	al "azul3d.org/native/al.v1"
	"log"
	"os"
	"time"
	"unsafe"
)

func main() {
	var duration float64
	var files []string
	switch len(os.Args) {
	case 1:
		log.Panicf("wav player: No files specified.\nUsage: go run main.go file1 file2...\n", len(os.Args))
	default:
		files = os.Args[1:]
	}

	device, err := al.OpenDevice("", nil)
	if err != nil {
		log.Panic(err)
	}
	defer device.Close()

	var buffers = make([]uint32, len(files))
	device.GenBuffers(int32(len(buffers)), &buffers[0])
	var sources = make([]uint32, len(files))
	device.GenSources(int32(len(sources)), &sources[0])

	for i, file := range files {
		data, dur, config := readFile(file)

		if dur > duration {
			duration = dur
		}

		if config.Channels == 1 {
			device.BufferData(buffers[i], al.FORMAT_MONO16, unsafe.Pointer(&data[0]), int32(int(unsafe.Sizeof(data[0]))*len(data)), int32(config.SampleRate))
		} else {
			device.BufferData(buffers[i], al.FORMAT_STEREO16, unsafe.Pointer(&data[0]), int32(int(unsafe.Sizeof(data[0]))*len(data)), int32(config.SampleRate))
		}

		device.Sourcei(sources[i], al.BUFFER, int32(buffers[i]))
	}

	log.Println("Duration:", duration, "seconds")

	device.SourcePlayv(sources)

	time.Sleep(time.Duration(duration * float64(time.Second)))

	device.DeleteSources(int32(len(sources)), &sources[0])
	device.DeleteBuffers(int32(len(buffers)), &buffers[0])
	log.Println("Done.")
}

func readFile(filename string) (data []audio.PCM16, duration float64, config audio.Config) {
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
	config = decoder.Config()
	log.Printf("Decoding a %s file.\n", format)
	log.Println(config)

	duration = float64(fi.Size()) / float64(config.SampleRate*config.Channels*16/8)

	// Create a buffer that can hold 3 second of audio samples
	bufSize := int(duration * float64(config.SampleRate*config.Channels)) // this might be too big
	// Most WAVs are PCM16
	buf := make(audio.PCM16Samples, bufSize)

	// Fill the buffer with as many audio samples as we can
	read, err := decoder.Read(buf)
	if err != nil && err != audio.EOS {
		log.Panic(err)
	}

	return []audio.PCM16(buf)[:read], duration, config
}
