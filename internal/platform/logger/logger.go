package logger

// #include <stdlib.h>
//
// void clear() {
//  system("clear");
// }
import "C" //nolint:typecheck
import (
	_ "embed"
	"log"
	"os"
	"time"
)

// Logger is the standard logger interface.
type Logger interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Flush()
	PrintHeader()
}

type lg struct {
	logger *log.Logger
}

// header logo of company
//
//go:embed header.txt
var header []byte

// New initializes a new logger.
func New() Logger {
	const flag = 5
	return &lg{logger: log.New(os.Stdout, "", flag)}
}

func (l *lg) Error(args ...interface{}) {
	l.logger.Println(args...)
}

func (l *lg) Errorf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

func (l *lg) Info(args ...interface{}) {
	l.logger.Println(args...)
}

func (l *lg) Infof(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

func (l *lg) PrintHeader() {
	l.Info(string(header))
	time.Sleep(time.Second)
}

func (l *lg) Flush() {
	C.clear()
	l.Info(string(header))
	time.Sleep(time.Second)
}
