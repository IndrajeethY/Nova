package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

type TelegramSender interface {
	SendLog(msg string) error
}

type Logger struct {
	mu           sync.RWMutex
	tgSender     TelegramSender
	logToChannel bool
	minLevel     LogLevel
}

var instance *Logger
var once sync.Once
var logFileWriter *lumberjack.Logger

func GetInstance() *Logger {
	once.Do(func() {
		instance = &Logger{
			logToChannel: true,
			minLevel:     LevelInfo,
		}
	})
	return instance
}

func Init() error {
	os.Remove("bot_logs.log")

	logFileWriter = &lumberjack.Logger{
		Filename:   "bot_logs.log",
		MaxSize:    20,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	log.SetOutput(io.MultiWriter(os.Stdout, logFileWriter))
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "15:04:05",
		FullTimestamp:   true,
		ForceColors:     true,
	})
	log.SetLevel(log.InfoLevel)

	return nil
}

func GetLogWriter() io.Writer {
	if logFileWriter == nil {
		return os.Stdout
	}
	return io.MultiWriter(os.Stdout, logFileWriter)
}

func SetTelegramSender(sender TelegramSender) {
	l := GetInstance()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.tgSender = sender
}

func SetLogToChannel(enabled bool) {
	l := GetInstance()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logToChannel = enabled
}

func SetMinLevel(level LogLevel) {
	l := GetInstance()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

func (l *Logger) sendToChannel(level, msg string) {
	l.mu.RLock()
	sender := l.tgSender
	shouldLog := l.logToChannel
	l.mu.RUnlock()

	if !shouldLog || sender == nil {
		return
	}

	go func() {
		formatted := fmt.Sprintf("<b>[%s]</b> <code>%s</code>\n%s",
			level, time.Now().Format("15:04:05"), msg)
		sender.SendLog(formatted)
	}()
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Warn(args ...interface{}) {
	msg := fmt.Sprint(args...)
	log.Warn(msg)
	GetInstance().sendToChannel("⚠️ WARN", msg)
}

func Warnf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Warn(msg)
	GetInstance().sendToChannel("⚠️ WARN", msg)
}

func Error(args ...interface{}) {
	msg := fmt.Sprint(args...)
	log.Error(msg)
	GetInstance().sendToChannel("❌ ERROR", msg)
}

func Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Error(msg)
	GetInstance().sendToChannel("❌ ERROR", msg)
}

func Fatal(args ...interface{}) {
	msg := fmt.Sprint(args...)
	GetInstance().sendToChannel("💀 FATAL", msg)
	log.Fatal(msg)
}

func Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	GetInstance().sendToChannel("💀 FATAL", msg)
	log.Fatal(msg)
}

func Event(event string, args ...interface{}) {
	msg := fmt.Sprint(args...)
	log.WithField("event", event).Info(msg)
	GetInstance().sendToChannel("📋 "+event, msg)
}

func Eventf(event, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.WithField("event", event).Info(msg)
	GetInstance().sendToChannel(fmt.Sprintf("📋 %s", event), msg)
}

func Command(user, command string) {
	log.WithFields(log.Fields{
		"user":    user,
		"command": command,
	}).Info("Command executed")
}

func Startup(msg string) {
	log.Info(msg)
	GetInstance().sendToChannel("🚀 STARTUP", msg)
}

func Shutdown(msg string) {
	log.Info(msg)
	GetInstance().sendToChannel("🛑 SHUTDOWN", msg)
}
