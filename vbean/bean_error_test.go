package vbean_test

import (
	"errors"
	"strconv"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vbean"
)

func TestFacadeBeanErrorContract(t *testing.T) {
	_, err := vbean.ToMap(nil)
	assertFacadeBeanCode(t, err, knifer.ErrCodeInvalidInput)

	var dst userModel
	err = vbean.CopyProperties(map[string]any{"age": "not-a-number"}, &dst)
	assertFacadeBeanCode(t, err, knifer.ErrCodeInvalidInput)
	var numErr *strconv.NumError
	if !errors.As(err, &numErr) {
		t.Fatalf("CopyProperties should preserve strconv.NumError cause: %v", err)
	}
}
