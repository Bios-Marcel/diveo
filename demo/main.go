package main

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"image"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/kbinani/screenshot"
)

const (
	duration              = 10
	recordingFPS          = 60
	amountOfImages        = recordingFPS * duration
	durationBetweenImages = 1000 / recordingFPS * time.Millisecond
)

func main() {
	cmd := exec.Command("xrectsel")
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	rectAsString := strings.TrimSpace(string(out))
	size := strings.Split(rectAsString[:strings.IndexRune(rectAsString, '+')], "x")
	position := strings.Split(rectAsString[strings.IndexRune(rectAsString, '+')+1:], "+")

	width, _ := strconv.ParseInt(size[0], 10, 64)
	height, _ := strconv.ParseInt(size[1], 10, 64)
	x, _ := strconv.ParseInt(position[0], 10, 64)
	y, _ := strconv.ParseInt(position[1], 10, 64)
	region := image.Rectangle{
		Min: image.Point{int(x), int(y)},
		Max: image.Point{int(x) + int(width), int(y) + int(height)}}

	log.Printf("Recording %dx%d at X:%d and Y:%d\n ...", width, height, x, y)

	ticker := time.NewTicker(durationBetweenImages)
	var images []*image.RGBA
	for i := 0; i < amountOfImages; i++ {
		result, err := screenshot.CaptureRect(region)
		if err == nil {
			images = append(images, result)
		} else {
			panic(err)
		}
		<-ticker.C
	}
	ticker.Stop()

	log.Println("Composing video data ...")

	data, err := toDiveo(recordingFPS, images)
	if err == nil {
		log.Println("Size of video in KB:", len(data)/8/1000)
	}

	writeErr := ioutil.WriteFile("output.diveo", data, 0644)
	if writeErr != nil {
		panic(writeErr)
	}
}

func toDiveo(fps int, images []*image.RGBA) ([]byte, error) {
	if len(images) == 0 {
		return nil, errors.New("Can't create empty video")
	}

	metaData := make([]byte, 0, 72)

	//8 bit version
	metaData = append(metaData, 1)
	//8 bit fps
	metaData = append(metaData, byte(fps))
	//32 bit duration ms
	lengthAsBytes := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(lengthAsBytes, uint32(len(images)/fps*1000))
	metaData = append(metaData, lengthAsBytes...)
	//12 bit width and 12 bit height
	bounds := images[0].Rect.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	widthAndHeight := make([]byte, 3, 3)
	widthAndHeight[0] = byte((width >> 4) & 255)
	widthAndHeight[1] = byte(((width << 4) & 255) | ((height >> 8) & 15))
	widthAndHeight[2] = byte(height & 255)
	metaData = append(metaData, widthAndHeight...)

	//FIXME Is there a more optimal way of doing this?
	imageData := make([]byte, 0, width*height*3)

	//data of the first (origin) image
	firstImage := images[0]
	for index := 0; index < len(firstImage.Pix); index += 4 {
		imageData = append(imageData,
			byte(firstImage.Pix[index]),
			byte(firstImage.Pix[index+1]),
			byte(firstImage.Pix[index+2]))
	}

	//diffdata of the all other frames in comparison to their previous frame
	var pixelIndex int
	if len(images) > 1 {
		previousImage := firstImage
		var rN, gN, bN uint8
		var currentImage *image.RGBA
		for imageIndex := 1; imageIndex < len(images); imageIndex++ {
			currentImage = images[imageIndex]
			imageData = append(imageData,
				byte(0),
				byte(0),
				byte(0),
				byte(currentImage.Pix[0]),
				byte(currentImage.Pix[1]),
				byte(currentImage.Pix[2]))
			for index := 4; index < len(currentImage.Pix)-4; index += 4 {
				rN = currentImage.Pix[index]
				gN = currentImage.Pix[index+1]
				bN = currentImage.Pix[index+2]
				if rN != previousImage.Pix[index] ||
					gN != previousImage.Pix[index+1] ||
					bN != previousImage.Pix[index+2] {
					pixelIndex = index / 4
					imageData = append(imageData,
						byte((pixelIndex>>16)&255),
						byte((pixelIndex>>8)&255),
						byte(pixelIndex&255),
						byte(rN),
						byte(gN),
						byte(bN))
				}
			}
			pixelIndex = (len(currentImage.Pix) - 4) / 4

			imageData = append(imageData,
				byte((pixelIndex>>16)&255),
				byte((pixelIndex>>8)&255),
				byte(pixelIndex&255),
				byte(currentImage.Pix[len(currentImage.Pix)-4]),
				byte(currentImage.Pix[len(currentImage.Pix)-3]),
				byte(currentImage.Pix[len(currentImage.Pix)-2]))

			previousImage = currentImage
		}

		//Cleanup
		previousImage = nil
		currentImage = nil
	}

	compressed, err := gZipData(imageData)
	if err != nil {
		return nil, err
	}

	return append(metaData, compressed...), nil
}

func gZipData(data []byte) (compressedData []byte, err error) {
	var b bytes.Buffer
	gz, _ := gzip.NewWriterLevel(&b, gzip.BestCompression)

	_, err = gz.Write(data)
	if err != nil {
		return
	}

	if err = gz.Flush(); err != nil {
		return
	}

	if err = gz.Close(); err != nil {
		return
	}

	compressedData = b.Bytes()

	return
}
