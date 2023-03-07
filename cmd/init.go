package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

func init() {
	// SetReportCaller sets whether the standard logger will include the calling
	// method as a field.
	logrus.SetReportCaller(false)

	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.JSONFormatter{})
	// logrus.SetFormatter(&logrus.TextFormatter{
	// 	ForceColors: true,
	// })

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.InfoLevel)
}
