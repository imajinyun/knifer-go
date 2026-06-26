package csvx

import knifer "github.com/imajinyun/knifer-go"

// CSVError represents an error produced by CSV helpers.
type CSVError struct {
	Code  knifer.ErrCode
	Msg   string
	Cause error
}

// Error returns the CSV error message.
func (e *CSVError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}
	return e.Msg
}

// ErrorCode returns the knifer-go error code.
func (e *CSVError) ErrorCode() knifer.ErrCode {
	if e == nil {
		return ""
	}
	return e.Code
}

// Unwrap returns the underlying cause.
func (e *CSVError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is supports errors.Is(err, knifer.ErrCodeXxx) matching by error code.
func (e *CSVError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if code, ok := target.(knifer.ErrCode); ok {
		return e.Code == code
	}
	if other, ok := target.(*CSVError); ok {
		return e.Code == other.Code
	}
	return false
}

func invalidInput(msg string) *CSVError {
	return &CSVError{Code: knifer.ErrCodeInvalidInput, Msg: msg}
}

func wrapCSVError(code knifer.ErrCode, msg string, cause error) error {
	if cause == nil {
		return nil
	}
	return &CSVError{Code: code, Msg: msg, Cause: cause}
}
