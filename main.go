package main

import (
	"fmt"
	"io"
	"os"

	"github.com/gofrs/flock"

	"github.com/aaronriekenberg/go-simplerotate/constants"
	"github.com/aaronriekenberg/go-simplerotate/logging"
	"github.com/aaronriekenberg/go-simplerotate/rotation"
)

var logger = logging.GetLogger()

func acquireFlock() *flock.Flock {
	logger.Printf("begin acquireFlock lockFileName = %q", constants.LockFileName)
	flock := flock.New(constants.LockFileName)
	err := flock.Lock()
	if err != nil {
		logger.Fatalf("flock.Lock error: %v", err)
	}
	logger.Printf("end acquireFlock")
	return flock
}

// Similar to io.CopyN:
// Returns nil when max bytes have been written to output file and need to rotate.
// Returns io.EOF when EOF is read from input.
// Otherwise returns non-nil error.
func copyInputToOutputFile() error {
	logger.Printf("begin copyInputToOutputFile")

	outputFile, err := os.OpenFile(constants.OutputFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening output file: %w", err)
	}
	defer outputFile.Close()

	stats, err := outputFile.Stat()
	if err != nil {
		return fmt.Errorf("outputFile.Stat error: %w", err)
	}

	outputFileSizeBytes := stats.Size()
	logger.Printf("outputFileSizeBytes = %v", outputFileSizeBytes)
	if outputFileSizeBytes >= constants.MaxFileSizeBytes {
		logger.Printf("outputFileSizeBytes >= maxFileSizeBytes")
		return nil
	}

	maxBytesToWriteToOutputFile := constants.MaxFileSizeBytes - outputFileSizeBytes
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

	logger.Printf("begin main")

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

	for {
		err := copyInputToOutputFile()
		if err == io.EOF {
			logger.Printf("copyInputToOutputFile returned EOF")
			break
		} else if err != nil {
			logger.Fatalf("copyInputToOutputFile err: %v", err)
		}

		rotation.RotateOutputFiles()
	}

	logger.Printf("end main")
}
