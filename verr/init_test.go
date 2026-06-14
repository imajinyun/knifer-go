package verr_test

import (
	"strings"
	"testing"

	"github.com/imajinyun/go-knifer/verr"
)

func TestInitWithOptionsFacade(t *testing.T) {
	var b strings.Builder
	verr.InitWithOptions(verr.WithLogOutput(&b), verr.WithReportCaller(false))
}
