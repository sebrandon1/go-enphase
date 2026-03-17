package main

import (
	"fmt"
	"os"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
)

var useColor = os.Getenv("NO_COLOR") == ""

func logf(level, color, format string, args ...any) {
	ts := time.Now().Format("2006-01-02T15:04:05")
	msg := fmt.Sprintf(format, args...)
	if useColor {
		fmt.Fprintf(os.Stderr, "%s[%s]%s %s %s\n", color, level, colorReset, ts, msg)
	} else {
		fmt.Fprintf(os.Stderr, "[%s] %s %s\n", level, ts, msg)
	}
}

// Info logs an informational message to stderr.
func Info(format string, args ...any) {
	logf("INFO", colorGreen, format, args...)
}

// Warn logs a warning message to stderr.
func Warn(format string, args ...any) {
	logf("WARN", colorYellow, format, args...)
}

// Error logs an error message to stderr.
func Error(format string, args ...any) {
	logf("ERROR", colorRed, format, args...)
}
