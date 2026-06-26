package imgx

import (
	"fmt"
	"strconv"
	"strings"

	randutil "github.com/imajinyun/knifer-go/internal/rand"
)

// CodeGenerator mirrors the CodeGenerator interface from captcha.
//
//	Gen returns the raw captcha text that is rendered into the image.
//	Verify   checks whether user input matches the raw text. Implementations may
//	         use different semantics; RandomGenerator compares directly, while
//	         MathGenerator evaluates the expression.
type CodeGenerator interface {
	Gen() string
	Verify(code, userInput string) bool
}

// ---------------------------------------------------------------------------
// RandomGenerator mirrors RandomGenerator.
// ---------------------------------------------------------------------------

// RandomGenerator generates random character captchas.
type RandomGenerator struct {
	// BaseStr is the character set; defaults to digits plus upper/lowercase letters.
	BaseStr string
	// Length is the captcha length.
	Length int
}

// NewRandomGenerator creates a generator with the default character set.
func NewRandomGenerator(length int) *RandomGenerator {
	return &RandomGenerator{BaseStr: randutil.BaseCharNumberUC, Length: length}
}

// NewRandomGeneratorWithBase creates a generator with a custom base string and length.
func NewRandomGeneratorWithBase(base string, length int) *RandomGenerator {
	return &RandomGenerator{BaseStr: base, Length: length}
}

// Gen returns a random string.
func (g *RandomGenerator) Gen() string {
	return g.GenWithOptions()
}

// GenWithOptions returns a random string with per-call options.
func (g *RandomGenerator) GenWithOptions(opts ...GeneratorOption) string {
	base := g.BaseStr
	if base == "" {
		base = randutil.BaseCharNumberUC
	}
	n := g.Length
	if n <= 0 {
		n = 4
	}
	runes := []rune(base)
	out := make([]rune, n)
	cfg := applyGeneratorOptions(opts)
	for i := 0; i < n; i++ {
		out[i] = runes[normalizeRandomIndex(cfg.randomInt(len(runes)), len(runes))]
	}
	return string(out)
}

// Verify compares case-insensitively and rejects blank input.
func (g *RandomGenerator) Verify(code, userInput string) bool {
	if strings.TrimSpace(userInput) == "" {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(code), strings.TrimSpace(userInput))
}

// ---------------------------------------------------------------------------
// MathGenerator mirrors the utility toolkit MathGenerator.
// ---------------------------------------------------------------------------
const mathOperators = "+-*"

type generatorConfig struct {
	randomInt func(max int) int
	parseInt  func(string) (int, error)
}

// GeneratorOption customizes captcha code generation per call.
type GeneratorOption func(*generatorConfig)

// WithGeneratorRandomInt sets the random integer function used by Gen*WithOptions helpers.
func WithGeneratorRandomInt(randomInt func(max int) int) GeneratorOption {
	return func(c *generatorConfig) {
		if randomInt != nil {
			c.randomInt = randomInt
		}
	}
}

// WithGeneratorIntParser sets the integer parser used by math captcha verification.
func WithGeneratorIntParser(parser func(string) (int, error)) GeneratorOption {
	return func(c *generatorConfig) {
		if parser != nil {
			c.parseInt = parser
		}
	}
}

func applyGeneratorOptions(opts []GeneratorOption) generatorConfig {
	cfg := generatorConfig{randomInt: randutil.RandomInt, parseInt: strconv.Atoi}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.randomInt == nil {
		cfg.randomInt = randutil.RandomInt
	}
	if cfg.parseInt == nil {
		cfg.parseInt = strconv.Atoi
	}
	return cfg
}

// MathGenerator generates expression captchas such as "12+3 =" and verifies
// user input by evaluating the expression.
type MathGenerator struct {
	// NumberLength is the maximum digit count of operands; the utility toolkit defaults to 2.
	NumberLength int
	// ResultHasNegativeNumber controls whether negative results are allowed.
	ResultHasNegativeNumber bool
}

// NewMathGenerator creates a generator with numberLength=2 and negative results enabled.
func NewMathGenerator() *MathGenerator {
	return &MathGenerator{NumberLength: 2, ResultHasNegativeNumber: true}
}

// NewMathGeneratorWith creates a generator with custom options.
func NewMathGeneratorWith(numberLength int, resultHasNegativeNumber bool) *MathGenerator {
	if numberLength <= 0 {
		numberLength = 2
	}
	return &MathGenerator{NumberLength: numberLength, ResultHasNegativeNumber: resultHasNegativeNumber}
}

// Length returns the rendered captcha length: numberLength*2 + 2.
func (g *MathGenerator) Length() int { return g.NumberLength*2 + 2 }

// Gen returns an "a op b=" expression padded with spaces on the right.
func (g *MathGenerator) Gen() string {
	return g.GenWithOptions()
}

// GenWithOptions returns an "a op b=" expression padded with spaces on the right with per-call options.
func (g *MathGenerator) GenWithOptions(opts ...GeneratorOption) string {
	limit := g.limit()
	cfg := applyGeneratorOptions(opts)
	op := mathOperators[normalizeRandomIndex(cfg.randomInt(len(mathOperators)), len(mathOperators))]
	a := normalizeRandomIndex(cfg.randomInt(limit), limit)
	var b int
	if !g.ResultHasNegativeNumber && op == '-' {
		if a == 0 {
			b = 0
		} else {
			b = normalizeRandomIndex(cfg.randomInt(a), a)
		}
	} else {
		b = normalizeRandomIndex(cfg.randomInt(limit), limit)
	}
	n1 := padRight(strconv.Itoa(a), g.NumberLength, ' ')
	n2 := padRight(strconv.Itoa(b), g.NumberLength, ' ')
	return fmt.Sprintf("%s%c%s=", n1, op, n2)
}

// Verify evaluates code and compares it with user input.
func (g *MathGenerator) Verify(code, userInput string) bool {
	return g.VerifyWithOptions(code, userInput)
}

// VerifyWithOptions evaluates code and compares it with user input using custom providers.
func (g *MathGenerator) VerifyWithOptions(code, userInput string, opts ...GeneratorOption) bool {
	cfg := applyGeneratorOptions(opts)
	got, err := cfg.parseInt(strings.TrimSpace(userInput))
	if err != nil {
		return false
	}
	v, ok := evalMathExprWithOptions(code, opts...)
	if !ok {
		return false
	}
	return v == got
}

// limit returns the operand upper bound: 1 followed by numberLength zeros.
func (g *MathGenerator) limit() int {
	limit := 1
	for i := 0; i < g.NumberLength; i++ {
		limit *= 10
	}
	return limit
}

// padRight pads s on the right with c until length n.
func padRight(s string, n int, c byte) string {
	if len(s) >= n {
		return s
	}
	pad := make([]byte, n-len(s))
	for i := range pad {
		pad[i] = c
	}
	return s + string(pad)
}

func normalizeRandomIndex(v, max int) int {
	if max <= 0 {
		return 0
	}
	if v < 0 {
		v = -v
	}
	return v % max
}

// evalMathExpr parses a simple padded integer expression in "a op b=" form.
func evalMathExpr(s string) (int, bool) {
	return evalMathExprWithOptions(s)
}

// evalMathExprWithOptions parses a simple padded integer expression using custom providers.
func evalMathExprWithOptions(s string, opts ...GeneratorOption) (int, bool) {
	cfg := applyGeneratorOptions(opts)
	s = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "="))
	for _, op := range []byte{'+', '-', '*'} {
		// Find the first operator that is not the leading character.
		if i := strings.IndexByte(s, op); i > 0 {
			left := strings.TrimSpace(s[:i])
			right := strings.TrimSpace(s[i+1:])
			a, errA := cfg.parseInt(left)
			b, errB := cfg.parseInt(right)
			if errA != nil || errB != nil {
				return 0, false
			}
			switch op {
			case '+':
				return a + b, true
			case '-':
				return a - b, true
			case '*':
				return a * b, true
			}
		}
	}
	return 0, false
}
