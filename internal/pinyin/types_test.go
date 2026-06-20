package pinyin

import (
	"errors"
	"testing"
)

func TestConvertRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request ConvertRequest
		wantErr error
	}{
		{name: "valid plain", request: ConvertRequest{Text: "中国", ToneStyle: ToneStylePlain}},
		{name: "valid number", request: ConvertRequest{Text: "中国", ToneStyle: ToneStyleNumber, MaxInputRunes: 2}},
		{name: "valid mark", request: ConvertRequest{Text: "中国", ToneStyle: ToneStyleMark}},
		{name: "missing text", request: ConvertRequest{}, wantErr: ErrInvalidConvertRequest},
		{name: "blank text", request: ConvertRequest{Text: " \t\n"}, wantErr: ErrInvalidConvertRequest},
		{name: "nul text", request: ConvertRequest{Text: "中\x00国"}, wantErr: ErrInvalidConvertRequest},
		{name: "negative input limit", request: ConvertRequest{Text: "中国", MaxInputRunes: -1}, wantErr: ErrInvalidConvertRequest},
		{name: "input limit exceeded", request: ConvertRequest{Text: "中国", MaxInputRunes: 1}, wantErr: ErrInputLimitExceeded},
		{name: "invalid tone style", request: ConvertRequest{Text: "中国", ToneStyle: ToneStyle("bad")}, wantErr: ErrInvalidConvertRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestInitialsRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request InitialsRequest
		wantErr error
	}{
		{name: "valid text", request: InitialsRequest{Text: "中国", MaxInputRunes: 2}},
		{name: "missing text", request: InitialsRequest{}, wantErr: ErrInvalidInitialsRequest},
		{name: "nul text", request: InitialsRequest{Text: "中\x00国"}, wantErr: ErrInvalidInitialsRequest},
		{name: "negative input limit", request: InitialsRequest{Text: "中国", MaxInputRunes: -1}, wantErr: ErrInvalidInitialsRequest},
		{name: "input limit exceeded", request: InitialsRequest{Text: "中国", MaxInputRunes: 1}, wantErr: ErrInputLimitExceeded},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCloneCopiesMutableFields(t *testing.T) {
	convert := ConvertRequest{Text: "中国", Metadata: map[string]string{"trace": "one"}}
	convertClone := convert.Clone()
	convert.Metadata["trace"] = "changed"
	if convertClone.Metadata["trace"] != "one" {
		t.Fatalf("convert request clone was mutated: %+v", convertClone)
	}

	convertResp := ConvertResponse{
		Text:     "中国",
		Output:   "zhong guo",
		Tokens:   []Token{{Text: "中", Syllables: []string{"zhong"}, Metadata: map[string]string{"index": "0"}}},
		Metadata: map[string]string{"trace": "one"},
	}
	convertRespClone := convertResp.Clone()
	convertResp.Tokens[0].Text = "changed"
	convertResp.Tokens[0].Syllables[0] = "changed"
	convertResp.Tokens[0].Metadata["index"] = "changed"
	convertResp.Metadata["trace"] = "changed"
	if convertRespClone.Tokens[0].Text != "中" || convertRespClone.Tokens[0].Syllables[0] != "zhong" || convertRespClone.Tokens[0].Metadata["index"] != "0" || convertRespClone.Metadata["trace"] != "one" {
		t.Fatalf("convert response clone was mutated: %+v", convertRespClone)
	}

	initials := InitialsRequest{Text: "中国", Metadata: map[string]string{"trace": "one"}}
	initialsClone := initials.Clone()
	initials.Metadata["trace"] = "changed"
	if initialsClone.Metadata["trace"] != "one" {
		t.Fatalf("initials request clone was mutated: %+v", initialsClone)
	}

	initialsResp := InitialsResponse{Initials: []string{"z", "g"}, Metadata: map[string]string{"trace": "one"}}
	initialsRespClone := initialsResp.Clone()
	initialsResp.Initials[0] = "x"
	initialsResp.Metadata["trace"] = "changed"
	if initialsRespClone.Initials[0] != "z" || initialsRespClone.Metadata["trace"] != "one" {
		t.Fatalf("initials response clone was mutated: %+v", initialsRespClone)
	}
}
