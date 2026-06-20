package tokenize

import (
	"errors"
	"testing"
)

func TestTokenizeRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request TokenizeRequest
		wantErr error
	}{
		{name: "valid precise", request: TokenizeRequest{Text: "南京市长江大桥", Mode: ModePrecise}},
		{name: "valid search", request: TokenizeRequest{Text: "南京市长江大桥", Mode: ModeSearch, MaxInputRunes: 7, MaxTokens: 8}},
		{name: "valid full", request: TokenizeRequest{Text: "南京市长江大桥", Mode: ModeFull}},
		{name: "missing text", request: TokenizeRequest{}, wantErr: ErrInvalidTokenizeRequest},
		{name: "blank text", request: TokenizeRequest{Text: " \t\n"}, wantErr: ErrInvalidTokenizeRequest},
		{name: "nul text", request: TokenizeRequest{Text: "南京\x00"}, wantErr: ErrInvalidTokenizeRequest},
		{name: "negative input limit", request: TokenizeRequest{Text: "南京", MaxInputRunes: -1}, wantErr: ErrInvalidTokenizeRequest},
		{name: "input limit exceeded", request: TokenizeRequest{Text: "南京", MaxInputRunes: 1}, wantErr: ErrInputLimitExceeded},
		{name: "negative token limit", request: TokenizeRequest{Text: "南京", MaxTokens: -1}, wantErr: ErrInvalidTokenizeRequest},
		{name: "invalid mode", request: TokenizeRequest{Text: "南京", Mode: Mode("bad")}, wantErr: ErrInvalidTokenizeRequest},
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

func TestKeywordsRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request KeywordsRequest
		wantErr error
	}{
		{name: "valid text", request: KeywordsRequest{Text: "南京市长江大桥", Limit: 2, MaxInputRunes: 7}},
		{name: "missing text", request: KeywordsRequest{}, wantErr: ErrInvalidKeywordsRequest},
		{name: "nul text", request: KeywordsRequest{Text: "南京\x00"}, wantErr: ErrInvalidKeywordsRequest},
		{name: "negative input limit", request: KeywordsRequest{Text: "南京", MaxInputRunes: -1}, wantErr: ErrInvalidKeywordsRequest},
		{name: "input limit exceeded", request: KeywordsRequest{Text: "南京", MaxInputRunes: 1}, wantErr: ErrInputLimitExceeded},
		{name: "negative keyword limit", request: KeywordsRequest{Text: "南京", Limit: -1}, wantErr: ErrInvalidKeywordsRequest},
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
	tokenizeReq := TokenizeRequest{Text: "南京", Metadata: map[string]string{"trace": "one"}}
	tokenizeReqClone := tokenizeReq.Clone()
	tokenizeReq.Metadata["trace"] = "changed"
	if tokenizeReqClone.Metadata["trace"] != "one" {
		t.Fatalf("tokenize request clone was mutated: %+v", tokenizeReqClone)
	}

	tokenizeResp := TokenizeResponse{
		Text:     "南京",
		Tokens:   []Token{{Text: "南京", Start: 0, End: 2, Position: 0, Weight: 1.5, Metadata: map[string]string{"pos": "ns"}}},
		Metadata: map[string]string{"trace": "one"},
	}
	tokenizeRespClone := tokenizeResp.Clone()
	tokenizeResp.Tokens[0].Text = "changed"
	tokenizeResp.Tokens[0].Metadata["pos"] = "changed"
	tokenizeResp.Metadata["trace"] = "changed"
	if tokenizeRespClone.Tokens[0].Text != "南京" || tokenizeRespClone.Tokens[0].Metadata["pos"] != "ns" || tokenizeRespClone.Metadata["trace"] != "one" {
		t.Fatalf("tokenize response clone was mutated: %+v", tokenizeRespClone)
	}

	keywordsReq := KeywordsRequest{Text: "南京", Metadata: map[string]string{"trace": "one"}}
	keywordsReqClone := keywordsReq.Clone()
	keywordsReq.Metadata["trace"] = "changed"
	if keywordsReqClone.Metadata["trace"] != "one" {
		t.Fatalf("keywords request clone was mutated: %+v", keywordsReqClone)
	}

	keywordsResp := KeywordsResponse{Keywords: []Keyword{{Text: "南京", Score: 0.9, Metadata: map[string]string{"source": "tfidf"}}}, Metadata: map[string]string{"trace": "one"}}
	keywordsRespClone := keywordsResp.Clone()
	keywordsResp.Keywords[0].Text = "changed"
	keywordsResp.Keywords[0].Metadata["source"] = "changed"
	keywordsResp.Metadata["trace"] = "changed"
	if keywordsRespClone.Keywords[0].Text != "南京" || keywordsRespClone.Keywords[0].Metadata["source"] != "tfidf" || keywordsRespClone.Metadata["trace"] != "one" {
		t.Fatalf("keywords response clone was mutated: %+v", keywordsRespClone)
	}
}
