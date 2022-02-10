package main

import (
	"fmt"
	"io"
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

func getOutputFileSizeBytes() int64 {
	fileInfo, err := os.Stat(outputFileName)
	if err != nil {
		logger.Printf("stat error: %v", err)
		return 0
	}
	return fileInfo.Size()
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

	maxBytesToWriteToOutputFile := maxFileSizeBytes - outputFileSizeBytes
	logger.Printf("before loop maxBytesToWriteToOutputFile = %v", maxBytesToWriteToOutputFile)

	for {
		bytesWritten, err := io.CopyN(outputFile, os.Stdin, maxBytesToWriteToOutputFile)

		if err == io.EOF {
			logger.Printf("io.CopyN returned EOF")
			os.Exit(0)
		} else if err != nil {
			logger.Fatalf("io.CopyN error: %v", err)
		}

		logger.Printf("after io.CopyN bytesWritten = %v", bytesWritten)

		err = outputFile.Close()
		if err != nil {
			logger.Fatalf("outputFileWriter.Close error: %v", err)
		}

		rotateOutputFiles()

		outputFile, err = os.OpenFile(outputFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Fatalf("error opening output file: %v", err)
		}

		maxBytesToWriteToOutputFile = maxFileSizeBytes
	}
}
