package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

const (
	outputFileName   = "output"
	maxFileSizeBytes = 1 * 1024 * 1024
	maxOutputFiles   = 10
)

type rotationAction struct {
	fromFileName string
	toFileName   string
}

func (rotationAction rotationAction) rotate() {
	log.Printf("rotate from %v to %v", rotationAction.fromFileName, rotationAction.toFileName)
	os.Rename(rotationAction.fromFileName, rotationAction.toFileName)
}

func buildRotationActions() []rotationAction {
	if maxOutputFiles <= 1 {
		return []rotationAction{}
	}

	buildFileName := func(fileIndex int) string {
		switch fileIndex {
		case 0:
			return outputFileName
		default:
			return fmt.Sprintf("%v.%v", outputFileName, fileIndex)
		}
	}

	rotationActions := make([]rotationAction, 0, maxOutputFiles-1)
	for i := maxOutputFiles - 1; i > 0; i-- {
		fromFileName := buildFileName(i - 1)
		toFileName := buildFileName(i)
		rotationActions = append(rotationActions,
			rotationAction{
				fromFileName: fromFileName,
				toFileName:   toFileName,
			},
		)
	}

	return rotationActions
}

var rotationActions = buildRotationActions()

func rotateOutputFiles() {
	for _, action := range rotationActions {
		action.rotate()
	}
}

func getOutputFileSizeBytes() uint64 {
	fileInfo, err := os.Stat(outputFileName)
	if err != nil {
		log.Printf("stat error: %v", err)
		return 0
	}
	return uint64(fileInfo.Size())
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	log.Printf("begin main rotationActions = %+v", rotationActions)

	outputFileSizeBytes := getOutputFileSizeBytes()
	log.Printf("outputFileSizeBytes = %v", outputFileSizeBytes)

	outputFile, err := os.OpenFile(outputFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening output file: %v", err)
	}
	outputFileWriter := bufio.NewWriter(outputFile)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		log.Printf("got bytes = %v", bytes)

		bytesWritten, err := outputFileWriter.Write(bytes)
		if err != nil {
			log.Fatalf("outputFileWriter.Write error; %v", err)
		}
		outputFileSizeBytes += uint64(bytesWritten)

		bytesWritten, err = outputFileWriter.WriteRune('\n')
		if err != nil {
			log.Fatalf("outputFileWriter.WriteRune error; %v", err)
		}
		outputFileSizeBytes += uint64(bytesWritten)

		err = outputFileWriter.Flush()
		if err != nil {
			log.Fatalf("outputFileWriter.Flush error; %v", err)
		}

		if bytesWritten > maxFileSizeBytes {
			err = outputFile.Close()
			if err != nil {
				log.Fatalf("outputFileWriter.Close error; %v", err)
			}
			outputFileSizeBytes = 0

			rotateOutputFiles()

			outputFile, err = os.OpenFile(outputFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatalf("error opening output file: %v", err)
			}
			outputFileWriter = bufio.NewWriter(outputFile)
		}
	}

	if scanner.Err() != nil {
		log.Printf("scanner error %v", scanner.Err())
	}

	log.Printf("end main")
}
