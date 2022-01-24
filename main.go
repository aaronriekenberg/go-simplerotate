package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/gofrs/flock"
)

const (
	outputFileName   = "output"
	lockFileName     = "lock"
	maxFileSizeBytes = 1 * 1024 * 1024
	maxOutputFiles   = 10
)

type rotationAction struct {
	fromFileName string
	toFileName   string
}

func (rotationAction rotationAction) rotate() {
	logger.Printf("rotate from %v to %v", rotationAction.fromFileName, rotationAction.toFileName)
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
		logger.Printf("stat error: %v", err)
		return 0
	}
	return uint64(fileInfo.Size())
}

func main() {

	logger.Printf("begin main rotationActions = %+v", rotationActions)

	if len(os.Args) > 1 {
		logDirectory := os.Args[1]
		logger.Printf("logDirectory = %q", logDirectory)
		err := os.Chdir(logDirectory)
		if err != nil {
			logger.Fatalf("os.Chdir error: %v", err)
		}
	}

	logger.Printf("before flock")
	flock := flock.New(lockFileName)
	err := flock.Lock()
	if err != nil {
		logger.Fatalf("flock.Lock error: %v", err)
	}
	logger.Printf("after flock")

	outputFileSizeBytes := getOutputFileSizeBytes()
	logger.Printf("outputFileSizeBytes = %v", outputFileSizeBytes)

	outputFile, err := os.OpenFile(outputFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Fatalf("error opening output file: %v", err)
	}
	outputFileWriter := bufio.NewWriter(outputFile)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		logger.Printf("got bytes = %v", bytes)

		bytesWritten, err := outputFileWriter.Write(bytes)
		if err != nil {
			logger.Fatalf("outputFileWriter.Write error: %v", err)
		}
		outputFileSizeBytes += uint64(bytesWritten)

		bytesWritten, err = outputFileWriter.WriteRune('\n')
		if err != nil {
			logger.Fatalf("outputFileWriter.WriteRune error: %v", err)
		}
		outputFileSizeBytes += uint64(bytesWritten)

		err = outputFileWriter.Flush()
		if err != nil {
			logger.Fatalf("outputFileWriter.Flush error: %v", err)
		}

		logger.Printf("outputFileSizeBytes = %v maxFileSizeBytes = %v", outputFileSizeBytes, maxFileSizeBytes)

		if outputFileSizeBytes > maxFileSizeBytes {
			err = outputFile.Close()
			if err != nil {
				logger.Fatalf("outputFileWriter.Close error: %v", err)
			}
			outputFileSizeBytes = 0

			rotateOutputFiles()

			outputFile, err = os.OpenFile(outputFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				logger.Fatalf("error opening output file: %v", err)
			}
			outputFileWriter = bufio.NewWriter(outputFile)
		}
	}

	if scanner.Err() != nil {
		logger.Printf("scanner error %v", scanner.Err())
	}

	logger.Printf("end main")
}
