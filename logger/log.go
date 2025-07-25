package lg

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"s3-diff-archive/constants"
	"s3-diff-archive/utils"
)

// BufferedLogger encapsulates the logging state
type BufferedLogger struct {
	logChan        chan string
	done           chan struct{}
	printToConsole bool
	printToFile    bool
}

// CreateLogger sets up a buffered logger writing to the specified file path.
func CreateLogger(path string, printToConsole bool, printToFile bool) (*BufferedLogger, error) {

	logger := &BufferedLogger{
		logChan:        make(chan string, 10000),
		done:           make(chan struct{}),
		printToConsole: printToConsole,
		printToFile:    printToFile,
	}

	if printToFile {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
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
	}

	return logger, nil
}

// Log sends a log message to the logger's buffer.
func Log(logger *BufferedLogger, message string) {
	if !logger.printToFile {
		return
	}
	select {
	case logger.logChan <- message:
	default:
		// drop log if buffer is full (or handle as needed)
	}
}

func FormatedLog(logger *BufferedLogger, level string, message string) {
	toLog := fmt.Sprintf("%s | %s\t| %s", utils.NowTime(), level, message)
	if logger.printToConsole {
		prefix := ""
		switch level {
		case "INFO":
			prefix = constants.Green
		case "ERROR":
			prefix = constants.Red
		case "WARN":
			prefix = constants.Yellow
		default:
			prefix = ""
		}

		fmt.Printf("\r%s%s%s\n", prefix, toLog, constants.Reset)
	}
	Log(logger, toLog)
}

func (logger *BufferedLogger) Info(message string, args ...any) {
	FormatedLog(logger, "INFO", fmt.Sprintf(message, args...))
}

func (logger *BufferedLogger) Error(message string, args ...any) {
	FormatedLog(logger, "ERROR", fmt.Sprintf(message, args...))
}
func (logger *BufferedLogger) ElogR(message string, args ...any) error {
	msg := fmt.Sprintf(message, args...)
	FormatedLog(logger, "ERROR", msg)
	return fmt.Errorf(message, args...)
}
func (logger *BufferedLogger) Fatal(message string, args ...any) {
	FormatedLog(logger, "ERROR", fmt.Sprintf(message, args...))
	os.Exit(1)
}

func (logger *BufferedLogger) Warn(message string, args ...any) {
	FormatedLog(logger, "WARN", fmt.Sprintf(message, args...))
}

func (logger *BufferedLogger) Close() {
	// Log(logger, message)
	CloseLogger(logger)
}

// CloseLogger gracefully shuts down the logger and flushes the buffer.
func CloseLogger(logger *BufferedLogger) {
	if logger.printToFile {
		close(logger.logChan)
		<-logger.done // wait until flush and close are complete
	}
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
