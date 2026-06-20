package vhan_test

import (
	"context"
	"errors"
	"testing"

	"github.com/imajinyun/go-knifer/vhan"
)

type providerFunc struct {
	convert  func(context.Context, vhan.ConvertRequest) (vhan.ConvertResponse, error)
	initials func(context.Context, vhan.InitialsRequest) (vhan.InitialsResponse, error)
}

func (p providerFunc) Convert(ctx context.Context, request vhan.ConvertRequest) (vhan.ConvertResponse, error) {
	return p.convert(ctx, request)
}

func (p providerFunc) Initials(ctx context.Context, request vhan.InitialsRequest) (vhan.InitialsResponse, error) {
	return p.initials(ctx, request)
}

func TestConvertFacade(t *testing.T) {
	provider := providerFunc{
		convert: func(ctx context.Context, request vhan.ConvertRequest) (vhan.ConvertResponse, error) {
			return vhan.ConvertResponse{Text: request.Text, Output: "zhong guo"}, nil
		},
	}

	response, err := vhan.Convert(context.Background(), provider, vhan.ConvertRequest{Text: "中国"})
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}
	if response.Output != "zhong guo" {
		t.Fatalf("Convert response = %+v", response)
	}
}

func TestInitialsFacade(t *testing.T) {
	provider := providerFunc{
		initials: func(ctx context.Context, request vhan.InitialsRequest) (vhan.InitialsResponse, error) {
			return vhan.InitialsResponse{Text: request.Text, Output: "zg", Initials: []string{"z", "g"}}, nil
		},
	}

	response, err := vhan.Initials(context.Background(), provider, vhan.InitialsRequest{Text: "中国"})
	if err != nil {
		t.Fatalf("Initials returned error: %v", err)
	}
	if response.Output != "zg" {
		t.Fatalf("Initials response = %+v", response)
	}
}

func TestFacadeExposesErrors(t *testing.T) {
	_, err := vhan.Convert(context.Background(), nil, vhan.ConvertRequest{Text: "中国"})
	if !errors.Is(err, vhan.ErrMissingProvider) {
		t.Fatalf("Convert error = %v, want ErrMissingProvider", err)
	}

	_, err = vhan.Convert(context.Background(), providerFunc{}, vhan.ConvertRequest{Text: "中国", ToneStyle: vhan.ToneStyle("bad")})
	if !errors.Is(err, vhan.ErrInvalidConvertRequest) {
		t.Fatalf("Convert error = %v, want ErrInvalidConvertRequest", err)
	}

	_, err = vhan.Initials(context.Background(), providerFunc{}, vhan.InitialsRequest{Text: "中国", MaxInputRunes: 1})
	if !errors.Is(err, vhan.ErrInputLimitExceeded) {
		t.Fatalf("Initials error = %v, want ErrInputLimitExceeded", err)
	}
}
