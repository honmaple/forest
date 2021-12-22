package forest

import (
	"log"
	"os"
)

type (
	Logger interface {
		Info(...interface{})
		Warn(...interface{})
		Print(...interface{})
		Error(...interface{})
		Fatal(...interface{})
		Panic(...interface{})
		Infoln(...interface{})
		Warnln(...interface{})
		Println(...interface{})
		Errorln(...interface{})
		Fatalln(...interface{})
		Panicln(...interface{})
		Infof(string, ...interface{})
		Warnf(string, ...interface{})
		Printf(string, ...interface{})
		Errorf(string, ...interface{})
		Fatalf(string, ...interface{})
		Panicf(string, ...interface{})
	}
	logger struct {
		*log.Logger
	}
)

func newLogger() *logger {
	std := log.New(os.Stderr, "", log.LstdFlags)
	std.SetFlags(0)
	return &logger{std}
}

func (s *logger) Info(args ...interface{}) {
	s.Logger.Print(args...)
}

func (s *logger) Infof(format string, args ...interface{}) {
	s.Logger.Printf(format, args...)
}

func (s *logger) Infoln(args ...interface{}) {
	s.Logger.Println(args...)
}

func (s *logger) Warn(args ...interface{}) {
	s.Logger.Print(args...)
}

func (s *logger) Warnf(format string, args ...interface{}) {
	s.Logger.Printf(format, args...)
}

func (s *logger) Warnln(args ...interface{}) {
	s.Logger.Println(args...)
}

func (s *logger) Error(args ...interface{}) {
	s.Logger.Print(args...)
}

func (s *logger) Errorf(format string, args ...interface{}) {
	s.Logger.Printf(format, args...)
}

func (s *logger) Errorln(args ...interface{}) {
	s.Logger.Println(args...)
}
