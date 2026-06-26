package db

import (
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestDBErrorContract(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code knifer.ErrCode
	}{
		{
			name: "empty builder",
			err:  sqlErr(NewBuilder().SQL()),
			code: knifer.ErrCodeInvalidInput,
		},
		{
			name: "select without table",
			err:  sqlErr(Select("id").SQL()),
			code: knifer.ErrCodeInvalidInput,
		},
		{
			name: "insert without values",
			err:  sqlErr(Insert(NewEntity("users")).SQL()),
			code: knifer.ErrCodeInvalidInput,
		},
		{
			name: "update without values",
			err:  sqlErr(Update(NewEntity("users")).SQL()),
			code: knifer.ErrCodeInvalidInput,
		},
		{
			name: "delete without table",
			err:  sqlErr(Delete("").SQL()),
			code: knifer.ErrCodeInvalidInput,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertDBCode(t, tt.err, tt.code)
		})
	}
}
