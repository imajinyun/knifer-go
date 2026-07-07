package govreport

import (
	"encoding/json"
	"fmt"
	"io"
)

const (
	StatusPassed = "passed"
	StatusFailed = "failed"

	SeverityError = "error"
)

type Finding struct {
	RuleID   string `json:"rule_id"`
	Path     string `json:"path"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

type Envelope struct {
	Status   string    `json:"status"`
	Findings []Finding `json:"findings"`
}

func Error(ruleID, path, message string) Finding {
	return Finding{
		RuleID:   ruleID,
		Path:     path,
		Message:  message,
		Severity: SeverityError,
	}
}

func Passed() Envelope {
	return Envelope{
		Status:   StatusPassed,
		Findings: []Finding{},
	}
}

func Failed(findings []Finding) Envelope {
	return Envelope{
		Status:   StatusFailed,
		Findings: findings,
	}
}

func WriteJSON(w io.Writer, report any) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}
