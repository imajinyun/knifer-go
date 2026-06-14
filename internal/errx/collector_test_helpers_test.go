package errx

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
)

func silenceLogrus(t *testing.T) {
	t.Helper()
	logger := logrus.StandardLogger()
	oldOut := logger.Out
	oldFormatter := logger.Formatter
	oldLevel := logger.Level
	logger.SetOutput(io.Discard)
	logger.SetFormatter(EmptyFormatter)
	logger.SetLevel(logrus.TraceLevel)
	t.Cleanup(func() {
		logger.SetOutput(oldOut)
		logger.SetFormatter(oldFormatter)
		logger.SetLevel(oldLevel)
	})
}
