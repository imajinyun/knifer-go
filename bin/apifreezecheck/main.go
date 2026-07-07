package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/imajinyun/knifer-go/bin/internal/govreport"
)

var expectedStatuses = setOf("recommended", "compatibility", "experimental", "deprecated")

type checker struct {
	aiContext map[string]any
	tools     map[string]any
	findings  []govreport.Finding
}

type apiFreezeReport struct {
	govreport.Envelope
	DeprecatedCount   int `json:"deprecated_count"`
	ExperimentalCount int `json:"experimental_count"`
}

func main() {
	contextPath := flag.String("ai-context", "ai-context.json", "ai-context.json path")
	toolsPath := flag.String("tools", "docs/api/tools.json", "tools catalog path")
	jsonFlag := flag.Bool("json", false, "emit machine-readable JSON output")
	flag.Parse()

	aiContext, err := loadJSON(*contextPath, "ai-context.json")
	if err != nil {
		writeAPIFreezeReport(*jsonFlag, apiFreezeReport{
			Envelope: govreport.Failed([]govreport.Finding{govreport.Error("API_FREEZE_INPUT_ERROR", *contextPath, err.Error())}),
		})
		os.Exit(1)
	}
	tools, err := loadJSON(*toolsPath, "tools catalog")
	if err != nil {
		writeAPIFreezeReport(*jsonFlag, apiFreezeReport{
			Envelope: govreport.Failed([]govreport.Finding{govreport.Error("API_FREEZE_INPUT_ERROR", *toolsPath, err.Error())}),
		})
		os.Exit(1)
	}
	c := &checker{aiContext: aiContext, tools: tools}
	c.run()
	deprecated, experimental := c.apiStatusCounts()
	report := apiFreezeReport{
		DeprecatedCount:   deprecated,
		ExperimentalCount: experimental,
	}
	if len(c.findings) > 0 {
		report.Envelope = govreport.Failed(c.findings)
		writeAPIFreezeReport(*jsonFlag, report)
		os.Exit(1)
	}
	report.Envelope = govreport.Passed()
	writeAPIFreezeReport(*jsonFlag, report)
}

func (c *checker) run() {
	apiFreeze, ok := c.aiContext["api_freeze"].(map[string]any)
	if !ok {
		c.addError("API_FREEZE_SCHEMA_INVALID", "ai-context.json api_freeze must be an object")
		apiFreeze = map[string]any{}
	}
	if boolValue(apiFreeze["decision_card_required"]) != true {
		c.addError("API_FREEZE_DECISION_CARD_REQUIRED", "api_freeze.decision_card_required must be true")
	}
	if boolValue(apiFreeze["replacement_required_for_deprecation"]) != true {
		c.addError("API_FREEZE_DEPRECATION_REPLACEMENT_REQUIRED", "api_freeze.replacement_required_for_deprecation must be true")
	}
	if !sameSet(stringList(apiFreeze["allowed_statuses"]), expectedStatuses) {
		c.addError("API_FREEZE_ALLOWED_STATUS_INVALID", "api_freeze.allowed_statuses must match recommended, compatibility, experimental, deprecated")
	}

	publicFacades := c.publicFacades()
	decisionCards := list(apiFreeze["decision_cards"])
	if _, ok := apiFreeze["decision_cards"].([]any); !ok {
		c.addError("API_FREEZE_DECISION_CARD_SCHEMA", "api_freeze.decision_cards must be a list")
		decisionCards = nil
	}
	if boolValue(apiFreeze["decision_card_required"]) && len(decisionCards) < 5 {
		c.addError("API_FREEZE_DECISION_CARD_REQUIRED", "api_freeze.decision_cards must contain at least five v1 decision cards")
	}

	expectedCardIDs := setOf(
		"v1-public-api-entry-budget",
		"v1-dynamic-contract-matrix",
		"v1-heavy-dependency-isolation",
		"v1-error-taxonomy",
		"v1-security-threat-model",
	)
	seenCardIDs := map[string]bool{}
	cardsByID := map[string]map[string]any{}
	for index, item := range decisionCards {
		card, ok := item.(map[string]any)
		if !ok {
			c.addError("API_FREEZE_DECISION_CARD_SCHEMA", fmt.Sprintf("api_freeze.decision_cards[%d] must be an object", index))
			continue
		}
		cardID := stringValue(card["id"])
		if cardID == "" {
			c.addError("API_FREEZE_DECISION_CARD_SCHEMA", fmt.Sprintf("api_freeze.decision_cards[%d].id must be non-empty", index))
		} else if seenCardIDs[cardID] {
			c.addError("API_FREEZE_DECISION_CARD_DUPLICATE", "api_freeze.decision_cards duplicate id "+cardID)
		} else {
			seenCardIDs[cardID] = true
			cardsByID[cardID] = card
		}
		if !expectedStatuses[stringValue(card["status"])] {
			c.addError("API_FREEZE_DECISION_CARD_STATUS_INVALID", fmt.Sprintf("api_freeze.decision_cards[%d].status must be an allowed API status", index))
		}
		for _, field := range []string{"decision", "rationale"} {
			if strings.TrimSpace(stringValue(card[field])) == "" {
				c.addError("API_FREEZE_DECISION_CARD_SCHEMA", fmt.Sprintf("api_freeze.decision_cards[%d].%s must be non-empty", index, field))
			}
		}
		packages := stringList(card["packages"])
		if len(packages) == 0 {
			c.addError("API_FREEZE_DECISION_CARD_SCHEMA", fmt.Sprintf("api_freeze.decision_cards[%d].packages must be non-empty", index))
		} else {
			var unknown []string
			for _, pkg := range packages {
				if pkg != "all" && !publicFacades[pkg] {
					unknown = append(unknown, pkg)
				}
			}
			sort.Strings(unknown)
			if len(unknown) > 0 {
				c.addError("API_FREEZE_DECISION_CARD_UNKNOWN_PACKAGE", fmt.Sprintf("api_freeze.decision_cards[%d].packages contains unknown facade(s): %s", index, strings.Join(unknown, ", ")))
			}
		}
		if len(stringList(card["validation"])) < 2 {
			c.addError("API_FREEZE_DECISION_CARD_SCHEMA", fmt.Sprintf("api_freeze.decision_cards[%d].validation must contain at least two validation entries", index))
		}
	}
	if missing := missingKeys(expectedCardIDs, seenCardIDs); len(missing) > 0 {
		c.addError("API_FREEZE_DECISION_CARD_REQUIRED", "api_freeze.decision_cards missing required v1 decision card(s): "+strings.Join(missing, ", "))
	}

	apiStatusDecisionCards, ok := apiFreeze["api_status_decision_cards"].(map[string]any)
	if !ok {
		c.addError("API_FREEZE_STATUS_CARD_SCHEMA", "api_freeze.api_status_decision_cards must be an object")
		apiStatusDecisionCards = map[string]any{}
	}
	if !sameSet(mapKeys(apiStatusDecisionCards), expectedStatuses) {
		c.addError("API_FREEZE_STATUS_CARD_SCHEMA", "api_freeze.api_status_decision_cards must map recommended, compatibility, experimental, deprecated")
	}
	for _, status := range sortedSet(expectedStatuses) {
		cardIDs := stringList(apiStatusDecisionCards[status])
		if len(cardIDs) == 0 {
			c.addError("API_FREEZE_STATUS_CARD_SCHEMA", fmt.Sprintf("api_freeze.api_status_decision_cards.%s must contain at least one decision card id", status))
			continue
		}
		seenStatusCardIDs := map[string]bool{}
		for index, cardID := range cardIDs {
			if cardID == "" {
				c.addError("API_FREEZE_STATUS_CARD_SCHEMA", fmt.Sprintf("api_freeze.api_status_decision_cards.%s[%d] must be a non-empty string", status, index))
				continue
			}
			if seenStatusCardIDs[cardID] {
				c.addError("API_FREEZE_STATUS_CARD_DUPLICATE", fmt.Sprintf("api_freeze.api_status_decision_cards.%s duplicates decision card %s", status, cardID))
			}
			seenStatusCardIDs[cardID] = true
			card := cardsByID[cardID]
			if card == nil {
				c.addError("API_FREEZE_STATUS_CARD_UNKNOWN", fmt.Sprintf("api_freeze.api_status_decision_cards.%s references unknown decision card %s", status, cardID))
				continue
			}
			if stringValue(card["status"]) != status {
				c.addError("API_FREEZE_STATUS_CARD_STATUS_MISMATCH", fmt.Sprintf("api_freeze.api_status_decision_cards.%s references %s with status '%s'", status, cardID, stringValue(card["status"])))
			}
		}
	}

	c.validateFreezeChecks(apiFreeze)
	deprecatedFunctions, experimentalFunctions, mustFunctions := c.validateTools(cardsByID, apiStatusDecisionCards)
	if boolValue(apiFreeze["v1_candidate"]) && len(experimentalFunctions) > 0 {
		c.addError("API_FREEZE_EXPERIMENTAL_DENIED", "api_freeze.v1_candidate forbids experimental APIs: "+strings.Join(experimentalFunctions, ", "))
	}
	c.validateDeprecations(apiFreeze, deprecatedFunctions)
	c.validateMustInventory(apiFreeze, mustFunctions)
}

func (c *checker) validateFreezeChecks(apiFreeze map[string]any) {
	freezeChecks := stringList(apiFreeze["freeze_checks"])
	if len(freezeChecks) < 4 {
		c.addError("API_FREEZE_FREEZE_CHECKS_INCOMPLETE", "api_freeze.freeze_checks must document at least four freeze checks")
		return
	}
	freezeText := strings.ToLower(strings.Join(freezeChecks, " "))
	for _, term := range []string{"decision card", "replacement", "snapshot", "tools catalog"} {
		if !strings.Contains(freezeText, term) {
			c.addError("API_FREEZE_FREEZE_CHECKS_INCOMPLETE", fmt.Sprintf("api_freeze.freeze_checks must mention %q", term))
		}
	}
}

func (c *checker) validateTools(cardsByID map[string]map[string]any, apiStatusDecisionCards map[string]any) ([]string, []string, []string) {
	var deprecatedFunctions, experimentalFunctions, mustFunctions []string
	for _, packageValue := range list(c.tools["packages"]) {
		pkg := mapValue(packageValue)
		packageName := stringValue(pkg["name"])
		for _, fnValue := range list(pkg["functions"]) {
			fn := mapValue(fnValue)
			status := stringValue(fn["status"])
			name := packageName + "." + stringValue(fn["name"])
			if !expectedStatuses[status] {
				c.addError("API_FREEZE_TOOL_STATUS_UNKNOWN", fmt.Sprintf("%s has unknown API status %q", name, status))
				continue
			}
			var coveringCards []string
			for _, cardID := range stringList(apiStatusDecisionCards[status]) {
				card := cardsByID[cardID]
				if card == nil {
					continue
				}
				packages := stringList(card["packages"])
				if contains(packages, "all") || contains(packages, packageName) {
					coveringCards = append(coveringCards, cardID)
				}
			}
			if len(coveringCards) == 0 {
				c.addError("API_FREEZE_TOOL_STATUS_UNCOVERED", fmt.Sprintf("%s status %q is not covered by api_freeze.api_status_decision_cards", name, status))
			}
			if status == "deprecated" {
				deprecatedFunctions = append(deprecatedFunctions, name)
				synopsis := stringValue(fn["synopsis"])
				if !strings.Contains(synopsis, "Deprecated:") || !strings.Contains(synopsis, "Use ") {
					c.addError("API_FREEZE_DEPRECATED_SYNOPSIS_INVALID", fmt.Sprintf("%s is deprecated but synopsis must include 'Deprecated:' and a replacement using 'Use '", name))
				}
			}
			if status == "experimental" {
				experimentalFunctions = append(experimentalFunctions, name)
			}
			if strings.HasPrefix(stringValue(fn["name"]), "Must") {
				mustFunctions = append(mustFunctions, name)
			}
		}
	}
	sort.Strings(deprecatedFunctions)
	sort.Strings(experimentalFunctions)
	sort.Strings(mustFunctions)
	return deprecatedFunctions, experimentalFunctions, mustFunctions
}

func (c *checker) validateDeprecations(apiFreeze map[string]any, deprecatedFunctions []string) {
	declared := list(apiFreeze["deprecations"])
	if _, ok := apiFreeze["deprecations"].([]any); !ok {
		c.addError("API_FREEZE_DEPRECATION_SCHEMA", "api_freeze.deprecations must be a list")
		declared = nil
	}
	declaredNames := map[string]bool{}
	for index, item := range declared {
		entry, ok := item.(map[string]any)
		if !ok {
			c.addError("API_FREEZE_DEPRECATION_SCHEMA", fmt.Sprintf("api_freeze.deprecations[%d] must be an object", index))
			continue
		}
		name := stringValue(entry["name"])
		if name == "" {
			c.addError("API_FREEZE_DEPRECATION_SCHEMA", fmt.Sprintf("api_freeze.deprecations[%d].name must be a non-empty string", index))
			continue
		}
		declaredNames[name] = true
		if stringValue(entry["replacement"]) == "" {
			c.addError("API_FREEZE_DEPRECATION_SCHEMA", fmt.Sprintf("api_freeze.deprecations[%d].replacement must be a non-empty string", index))
		}
		if stringValue(entry["rationale"]) == "" {
			c.addError("API_FREEZE_DEPRECATION_SCHEMA", fmt.Sprintf("api_freeze.deprecations[%d].rationale must be a non-empty string", index))
		}
	}
	if missing := missingFromList(deprecatedFunctions, declaredNames); len(missing) > 0 {
		c.addError("API_FREEZE_DEPRECATION_INVENTORY_MISSING", "api_freeze.deprecations missing deprecated function(s): "+strings.Join(missing, ", "))
	}
}

func (c *checker) validateMustInventory(apiFreeze map[string]any, mustFunctions []string) {
	var inventory []any
	if value, exists := apiFreeze["must_api_inventory"]; exists {
		var ok bool
		inventory, ok = value.([]any)
		if !ok {
			c.addError("API_FREEZE_MUST_INVENTORY_SCHEMA", "api_freeze.must_api_inventory must be a list")
			inventory = nil
		}
	}
	inventoryNames := map[string]bool{}
	for index, item := range inventory {
		entry, ok := item.(map[string]any)
		if !ok {
			c.addError("API_FREEZE_MUST_INVENTORY_SCHEMA", fmt.Sprintf("api_freeze.must_api_inventory[%d] must be an object", index))
			continue
		}
		name := stringValue(entry["name"])
		if name == "" {
			c.addError("API_FREEZE_MUST_INVENTORY_SCHEMA", fmt.Sprintf("api_freeze.must_api_inventory[%d].name must be a non-empty string", index))
			continue
		}
		inventoryNames[name] = true
		replacement := stringValue(entry["replacement"])
		rationale := stringValue(entry["rationale"])
		docPath := stringValue(entry["doc_path"])
		for _, field := range []struct {
			name  string
			value string
		}{{"replacement", replacement}, {"rationale", rationale}, {"doc_path", docPath}} {
			if strings.TrimSpace(field.value) == "" {
				c.addError("API_FREEZE_MUST_INVENTORY_SCHEMA", fmt.Sprintf("api_freeze.must_api_inventory[%d].%s must be a non-empty string", index, field.name))
			}
		}
		if docPath == "" {
			continue
		}
		docBytes, err := os.ReadFile(docPath)
		if err != nil {
			if os.IsNotExist(err) {
				c.addError("API_FREEZE_MUST_INVENTORY_DOC_INVALID", fmt.Sprintf("api_freeze.must_api_inventory[%d].doc_path does not exist: %s", index, docPath))
			} else {
				c.addError("API_FREEZE_MUST_INVENTORY_DOC_INVALID", fmt.Sprintf("api_freeze.must_api_inventory[%d].doc_path cannot be read: %s", index, docPath))
			}
			continue
		}
		docText := string(docBytes)
		functionName := name
		if idx := strings.LastIndex(name, "."); idx >= 0 {
			functionName = name[idx+1:]
		}
		if !strings.Contains(docText, functionName) {
			c.addError("API_FREEZE_MUST_INVENTORY_DOC_INVALID", fmt.Sprintf("%s must document %s", docPath, name))
		}
		if replacement != "" {
			if !containsAnyReplacementToken(docText, replacement) {
				c.addError("API_FREEZE_MUST_INVENTORY_DOC_INVALID", fmt.Sprintf("%s must mention replacement guidance for %s: %s", docPath, name, replacement))
			}
		}
	}
	missing := missingFromList(mustFunctions, inventoryNames)
	if len(missing) > 0 {
		c.addError("API_FREEZE_MUST_INVENTORY_MISSING", "api_freeze.must_api_inventory missing Must API(s): "+strings.Join(missing, ", "))
	}
	mustSet := map[string]bool{}
	for _, name := range mustFunctions {
		mustSet[name] = true
	}
	var stale []string
	for name := range inventoryNames {
		if !mustSet[name] {
			stale = append(stale, name)
		}
	}
	sort.Strings(stale)
	if len(stale) > 0 {
		c.addError("API_FREEZE_MUST_INVENTORY_STALE", "api_freeze.must_api_inventory includes non-Must or missing API(s): "+strings.Join(stale, ", "))
	}
}

func (c *checker) publicFacades() map[string]bool {
	facades := map[string]bool{}
	for _, item := range list(c.aiContext["public_facades"]) {
		entry := mapValue(item)
		if pkg := stringValue(entry["package"]); pkg != "" {
			facades[pkg] = true
		}
	}
	return facades
}

func (c *checker) apiStatusCounts() (deprecated, experimental int) {
	for _, packageValue := range list(c.tools["packages"]) {
		for _, fnValue := range list(mapValue(packageValue)["functions"]) {
			switch stringValue(mapValue(fnValue)["status"]) {
			case "deprecated":
				deprecated++
			case "experimental":
				experimental++
			}
		}
	}
	return deprecated, experimental
}

func loadJSON(path, label string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("missing %s: %s", label, path)
		}
		return nil, fmt.Errorf("cannot read %s: %w", label, err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", label, err)
	}
	return out, nil
}

func (c *checker) addError(ruleID, message string) {
	c.findings = append(c.findings, govreport.Error(ruleID, "ai-context.json", message))
}

func writeAPIFreezeReport(jsonOutput bool, report apiFreezeReport) {
	if jsonOutput {
		if err := govreport.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "api-freeze check error: [API_FREEZE_INPUT_ERROR] cannot encode JSON output: %v\n", err)
		}
		return
	}
	if report.Status == govreport.StatusFailed {
		for _, finding := range report.Findings {
			fmt.Fprintf(os.Stderr, "api-freeze check error: [%s] %s\n", finding.RuleID, finding.Message)
		}
		return
	}
	fmt.Printf("api freeze metadata is valid (%d deprecated, %d experimental APIs)\n", report.DeprecatedCount, report.ExperimentalCount)
}

func mapValue(value any) map[string]any {
	mapping, _ := value.(map[string]any)
	if mapping == nil {
		return map[string]any{}
	}
	return mapping
}

func list(value any) []any {
	values, _ := value.([]any)
	return values
}

func stringValue(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

func stringList(value any) []string {
	values, ok := value.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, item := range values {
		if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
			out = append(out, strings.TrimSpace(text))
		}
	}
	return out
}

func boolValue(value any) bool {
	result, _ := value.(bool)
	return result
}

func setOf(values ...string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}

func sameSet(values []string, expected map[string]bool) bool {
	if len(values) != len(expected) {
		return false
	}
	for _, value := range values {
		if !expected[value] {
			return false
		}
	}
	return true
}

func mapKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedSet(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func missingKeys(expected map[string]bool, seen map[string]bool) []string {
	var missing []string
	for key := range expected {
		if !seen[key] {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	return missing
}

func missingFromList(values []string, seen map[string]bool) []string {
	var missing []string
	for _, value := range values {
		if !seen[value] {
			missing = append(missing, value)
		}
	}
	sort.Strings(missing)
	return missing
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsAnyReplacementToken(docText, replacement string) bool {
	replacer := strings.NewReplacer(" or ", " ", ",", " ")
	for _, raw := range strings.Fields(replacer.Replace(replacement)) {
		token := strings.Trim(raw, "` ,.")
		if token != "" && strings.Contains(docText, token) {
			return true
		}
	}
	return false
}
