package errx

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestWrapperExecReturnsFunctionError(t *testing.T) {
	silenceLogrus(t)

	want := errors.New("wrapped failure")
	got := Wrap(func() error { return want }).WithErrorf("failed").Exec(context.Background())
	if !ErrorIs(got, want) {
		t.Fatalf("Exec() error = %v, want %v", got, want)
	}
}

func TestWrapperExecConvertsPanic(t *testing.T) {
	silenceLogrus(t)

	got := Wrap(func() error {
		panic("panic from wrapper")
	}).WithWarnf("panic").Exec(context.TODO())
	if got == nil || !strings.Contains(got.Error(), "panic from wrapper") {
		t.Fatalf("Exec() panic error = %v, want panic value", got)
	}
}

func TestWrapperExecNilFunction(t *testing.T) {
	if err := Wrap(nil).Exec(context.Background()); err != nil {
		t.Fatalf("Exec(nil) = %v, want nil", err)
	}
}

func TestRecoverHelpers(t *testing.T) {
	silenceLogrus(t)

	want := errors.New("recover failure")
	if got := Recover(func() error { return want }, "recover"); !ErrorIs(got, want) {
		t.Fatalf("Recover() = %v, want %v", got, want)
	}
	got := RecoverWithoutError(func() { panic("recover without error") }, "recover without error")
	if got == nil || !strings.Contains(got.Error(), "recover without error") {
		t.Fatalf("RecoverWithoutError() = %v, want panic value", got)
	}
	if got := RecoverWithoutError(nil, "nil function"); got != nil {
		t.Fatalf("RecoverWithoutError(nil) = %v, want nil", got)
	}
}

func TestEmptyFormatterSuppressesOutput(t *testing.T) {
	data, err := EmptyFormatter.Format(logrus.NewEntry(logrus.New()))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Fatalf("EmptyFormatter output length = %d, want 0", len(data))
	}
}
