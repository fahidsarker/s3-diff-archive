package logger

import (
	"bufio"
	"bytes"
	"os"
)

// BufferedLogger encapsulates the logging state
type BufferedLogger struct {
	logChan chan string
	done    chan struct{}
}

// CreateLogger sets up a buffered logger writing to the specified file path.
func CreateLogger(path string) (*BufferedLogger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	logger := &BufferedLogger{
		logChan: make(chan string, 10000),
		done:    make(chan struct{}),
	}

	go func() {
		writer := bufio.NewWriterSize(file, 64*1024)
		defer func() {
			writer.Flush()
			file.Close()
			close(logger.done)
		}()

		for msg := range logger.logChan {
			writer.WriteString(msg + "\n")
		}
	}()

	return logger, nil
}

// Log sends a log message to the logger's buffer.
func Log(logger *BufferedLogger, message string) {
	select {
	case logger.logChan <- message:
	default:
		// drop log if buffer is full (or handle as needed)
	}
}

// CloseLogger gracefully shuts down the logger and flushes the buffer.
func CloseLogger(logger *BufferedLogger) {
	close(logger.logChan)
	<-logger.done // wait until flush and close are complete
}

func ReadLastLine(filePath string) (string, error) {
	const readBlockSize = 1024
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return "", err
	}
	fileSize := stat.Size()

	var (
		offset    int64
		foundLine bool
		lineBytes []byte
	)

	end := fileSize

	// Trim trailing newline(s)
	for end > 0 {
		b := make([]byte, 1)
		_, err := file.ReadAt(b, end-1)
		if err != nil {
			return "", err
		}
		if b[0] != '\n' && b[0] != '\r' {
			break
		}
		end--
	}

	// Start scanning backwards
	offset = end
	for offset > 0 && !foundLine {
		blockSize := readBlockSize
		if offset < int64(blockSize) {
			blockSize = int(offset)
			offset = 0
		} else {
			offset -= int64(blockSize)
		}

		buf := make([]byte, blockSize)
		_, err := file.ReadAt(buf, offset)
		if err != nil {
			return "", err
		}

		// Check for last newline
		if i := bytes.LastIndexByte(buf, '\n'); i >= 0 {
			// Line starts just after this newline
			start := offset + int64(i+1)
			length := end - start
			lineBuf := make([]byte, length)
			_, err := file.ReadAt(lineBuf, start)
			if err != nil {
				return "", err
			}
			lineBytes = bytes.TrimRight(lineBuf, "\r\n")
			foundLine = true
			break
		}
	}

	// No newline found: single-line file
	if !foundLine {
		lineBytes = make([]byte, end)
		_, err := file.ReadAt(lineBytes, 0)
		if err != nil {
			return "", err
		}
		lineBytes = bytes.TrimRight(lineBytes, "\r\n")
	}

	return string(lineBytes), nil
}
