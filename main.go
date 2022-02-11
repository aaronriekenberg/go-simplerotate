package main

import (
	"fmt"
	"io"
	"os"

	"github.com/gofrs/flock"

	loggerPackage "github.com/aaronriekenberg/go-simplerotate/logger"
)

const (
	outputFileName   = "output"
	lockFileName     = "lock"
	maxFileSizeBytes = 1 * 1024 * 1024
	maxOutputFiles   = 10
)

var logger = loggerPackage.GetLogger()

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
		return nil
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

func acquireFlock() *flock.Flock {
	logger.Printf("begin acquireFlock")
	flock := flock.New(lockFileName)
	err := flock.Lock()
	if err != nil {
		logger.Fatalf("flock.Lock error: %v", err)
	}
	logger.Printf("end acquireFlock")
	return flock
}

func getOutputFileSizeBytes() int64 {
	fileInfo, err := os.Stat(outputFileName)
	if err != nil {
		logger.Printf("stat error: %v", err)
		return 0
	}
	return fileInfo.Size()
}

func copyInputToOutputFile() error {
	outputFile, err := os.OpenFile(outputFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening output file: %w", err)
	}
	defer outputFile.Close()

	stats, err := outputFile.Stat()
	if err != nil {
		return fmt.Errorf("outputFile.Stat error: %w", err)
	}

	outputFileSizeBytes := stats.Size()
	maxBytesToWriteToOutputFile := maxFileSizeBytes - outputFileSizeBytes
	if maxBytesToWriteToOutputFile < 0 {
		maxBytesToWriteToOutputFile = 0
	}
	logger.Printf("maxBytesToWriteToOutputFile = %v", maxBytesToWriteToOutputFile)

	bytesWritten, err := io.CopyN(outputFile, os.Stdin, maxBytesToWriteToOutputFile)
	if err == io.EOF {
		logger.Printf("io.CopyN returned EOF")
		return io.EOF
	} else if err != nil {
		return fmt.Errorf("io.CopyN error: %w", err)
	}

	logger.Printf("end copyInputToOutputFile bytesWritten = %v", bytesWritten)
	return nil
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

	flock := acquireFlock()
	defer flock.Unlock()

	outputFileSizeBytes := getOutputFileSizeBytes()
	logger.Printf("initial outputFileSizeBytes = %v", outputFileSizeBytes)

	if outputFileSizeBytes >= maxFileSizeBytes {
		logger.Printf("initial outputFileSizeBytes at max, calling rotateOutputFiles")
		rotateOutputFiles()
	}

	for {
		err := copyInputToOutputFile()
		if err == io.EOF {
			logger.Printf("copyInputToOutputFile returned EOF")
			break
		} else if err != nil {
			logger.Fatalf("copyInputToOutputFile err: %v", err)
		}

		rotateOutputFiles()
	}

	logger.Printf("end main")
}
