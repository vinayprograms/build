package lexer

import (
	"testing"
)

func TestIsValidIdentifierStart(t *testing.T) {
	tests := []struct {
		char byte
		want bool
	}{
		{'a', true},
		{'z', true},
		{'A', true},
		{'Z', true},
		{'_', true},
		{'0', false},
		{'9', false},
		{'-', false},
		{'.', false},
		{' ', false},
		{'{', false},
		{'}', false},
		{':', false},
		{'"', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			if got := IsValidIdentifierStart(tt.char); got != tt.want {
				t.Errorf("IsValidIdentifierStart(%q) = %v, want %v", tt.char, got, tt.want)
			}
		})
	}
}

func TestIsValidIdentifierChar(t *testing.T) {
	tests := []struct {
		char byte
		want bool
	}{
		{'a', true},
		{'z', true},
		{'A', true},
		{'Z', true},
		{'_', true},
		{'0', true},
		{'9', true},
		{'.', true}, // for target.dir, target.file
		{'-', false},
		{' ', false},
		{'{', false},
		{'}', false},
		{':', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			if got := IsValidIdentifierChar(tt.char); got != tt.want {
				t.Errorf("IsValidIdentifierChar(%q) = %v, want %v", tt.char, got, tt.want)
			}
		})
	}
}

func TestIsInterpBoundary(t *testing.T) {
	tests := []struct {
		name     string
		prev     byte
		atSOL    bool
		expected bool
	}{
		// Valid boundaries
		{"space", ' ', false, true},
		{"tab", '\t', false, true},
		{"colon", ':', false, true},
		{"equals", '=', false, true},
		{"start of line", 0, true, true},
		{"slash", '/', false, true},
		{"double quote", '"', false, true},
		{"single quote", '\'', false, true},
		{"open paren", '(', false, true},
		{"close paren", ')', false, true},
		{"comma", ',', false, true},
		{"greater than", '>', false, true},
		{"less than", '<', false, true},
		{"hyphen", '-', false, true},
		{"close brace", '}', false, true},

		// Invalid boundaries
		{"letter", 'a', false, false},
		{"digit", '0', false, false},
		{"dollar", '$', false, false},
		{"underscore", '_', false, false},
		{"dot", '.', false, true},
		{"open bracket", '[', false, true},
		{"pipe", '|', false, true},
		{"underscore", '_', false, false},
		{"dollar", '$', false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInterpBoundary(tt.prev, tt.atSOL); got != tt.expected {
				t.Errorf("IsInterpBoundary(%q, %v) = %v, want %v", tt.prev, tt.atSOL, got, tt.expected)
			}
		})
	}
}

func TestScanInterpolation(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		pos        int // position of '{'
		prev       byte
		atSOL      bool
		wantResult InterpResult
		wantEnd    int // position after the interpolation
	}{
		// Valid interpolations
		{
			name:       "simple var at SOL",
			input:      "{var}",
			pos:        0,
			prev:       0,
			atSOL:      true,
			wantResult: InterpResult{Kind: InterpValid, Name: "var", Raw: false},
			wantEnd:    5,
		},
		{
			name:       "var after space",
			input:      "x {var}",
			pos:        2,
			prev:       ' ',
			atSOL:      false,
			wantResult: InterpResult{Kind: InterpValid, Name: "var", Raw: false},
			wantEnd:    7,
		},
		{
			name:       "var after equals",
			input:      "x={var}",
			pos:        2,
			prev:       '=',
			atSOL:      false,
			wantResult: InterpResult{Kind: InterpValid, Name: "var", Raw: false},
			wantEnd:    7,
		},
		{
			name:       "var after colon",
			input:      "x:{var}",
			pos:        2,
			prev:       ':',
			atSOL:      false,
			wantResult: InterpResult{Kind: InterpValid, Name: "var", Raw: false},
			wantEnd:    7,
		},
		{
			name:       "var with raw modifier",
			input:      "{var:raw}",
			pos:        0,
			prev:       0,
			atSOL:      true,
			wantResult: InterpResult{Kind: InterpValid, Name: "var", Raw: true},
			wantEnd:    9,
		},
		{
			name:       "var with underscore",
			input:      "{my_var}",
			pos:        0,
			prev:       0,
			atSOL:      true,
			wantResult: InterpResult{Kind: InterpValid, Name: "my_var", Raw: false},
			wantEnd:    8,
		},
		{
			name:       "var with numbers",
			input:      "{var123}",
			pos:        0,
			prev:       0,
			atSOL:      true,
			wantResult: InterpResult{Kind: InterpValid, Name: "var123", Raw: false},
			wantEnd:    8,
		},
		{
			name:       "var with dot (target.dir)",
			input:      "{target.dir}",
			pos:        0,
			prev:       0,
			atSOL:      true,
			wantResult: InterpResult{Kind: InterpValid, Name: "target.dir", Raw: false},
			wantEnd:    12,
		},
		{
			name:       "var with dot and raw",
			input:      "{target.file:raw}",
			pos:        0,
			prev:       0,
			atSOL:      true,
			wantResult: InterpResult{Kind: InterpValid, Name: "target.file", Raw: true},
			wantEnd:    17,
		},
		{
			name:       "var after slash (path pattern)",
			input:      "build/{name}.o",
			pos:        6,
			prev:       '/',
			atSOL:      false,
			wantResult: InterpResult{Kind: InterpValid, Name: "name", Raw: false},
			wantEnd:    12,
		},
		{
			name:       "var after quote (quoted string)",
			input:      `"{var}"`,
			pos:        1,
			prev:       '"',
			atSOL:      false,
			wantResult: InterpResult{Kind: InterpValid, Name: "var", Raw: false},
			wantEnd:    6,
		},
		{
			name:       "var after single quote",
			input:      `'{var}'`,
			pos:        1,
			prev:       '\'',
			atSOL:      false,
			wantResult: InterpResult{Kind: InterpValid, Name: "var", Raw: false},
			wantEnd:    6,
		},
		{
			name:       "var after open paren (function arg)",
			input:      "shell({cmd})",
			pos:        6,
			prev:       '(',
			atSOL:      false,
			wantResult: InterpResult{Kind: InterpValid, Name: "cmd", Raw: false},
			wantEnd:    11,
		},
		{
			name:       "var after comma (function args)",
			input:      "replace({src},{from},{to})",
			pos:        14,
			prev:       ',',
			atSOL:      false,
			wantResult: InterpResult{Kind: InterpValid, Name: "from", Raw: false},
			wantEnd:    20,
		},
		{
			name:       "var after close paren",
			input:      "func(){var}",
			pos:        6,
			prev:       ')',
			atSOL:      false,
			wantResult: InterpResult{Kind: InterpValid, Name: "var", Raw: false},
			wantEnd:    11,
		},
		{
			name:       "var after greater than (redirection)",
			input:      "echo >{file}",
			pos:        6,
			prev:       '>',
			atSOL:      false,
			wantResult: InterpResult{Kind: InterpValid, Name: "file", Raw: false},
			wantEnd:    12,
		},
		{
			name:       "var after less than (redirection)",
			input:      "cat <{input}",
			pos:        5,
			prev:       '<',
			atSOL:      false,
			wantResult: InterpResult{Kind: InterpValid, Name: "input", Raw: false},
			wantEnd:    12,
		},

		// Escape sequences
		{
			name:       "double open brace",
			input:      "{{var}}",
			pos:        0,
			prev:       0,
			atSOL:      true,
			wantResult: InterpResult{Kind: InterpEscapedOpen},
			wantEnd:    2,
		},
		{
			name:       "double open brace mid-string",
			input:      "x {{y",
			pos:        2,
			prev:       ' ',
			atSOL:      false,
			wantResult: InterpResult{Kind: InterpEscapedOpen},
			wantEnd:    4,
		},

		// Not an interpolation (no boundary)
		{
			name:       "no boundary - letter before",
			input:      "x{var}",
			pos:        1,
			prev:       'x',
			atSOL:      false,
			wantResult: InterpResult{Kind: InterpNotInterp},
			wantEnd:    1,
		},
		{
			name:       "no boundary - dollar before",
			input:      "${var}",
			pos:        1,
			prev:       '$',
			atSOL:      false,
			wantResult: InterpResult{Kind: InterpNotInterp},
			wantEnd:    1,
		},
		{
			name:       "no boundary - digit before",
			input:      "0{var}",
			pos:        1,
			prev:       '0',
			atSOL:      false,
			wantResult: InterpResult{Kind: InterpNotInterp},
			wantEnd:    1,
		},

		// Not an interpolation (invalid identifier start)
		{
			name:       "digit after brace",
			input:      "{0var}",
			pos:        0,
			prev:       0,
			atSOL:      true,
			wantResult: InterpResult{Kind: InterpNotInterp},
			wantEnd:    0,
		},
		{
			name:       "quote after brace",
			input:      `{"key"}`,
			pos:        0,
			prev:       0,
			atSOL:      true,
			wantResult: InterpResult{Kind: InterpNotInterp},
			wantEnd:    0,
		},
		{
			name:       "space after brace",
			input:      "{ var}",
			pos:        0,
			prev:       0,
			atSOL:      true,
			wantResult: InterpResult{Kind: InterpNotInterp},
			wantEnd:    0,
		},

		// Error cases
		{
			name:       "unclosed interpolation",
			input:      "{var",
			pos:        0,
			prev:       0,
			atSOL:      true,
			wantResult: InterpResult{Kind: InterpError, Name: "var", Error: "unclosed interpolation: {var"},
			wantEnd:    4,
		},
		{
			name:       "unclosed with raw",
			input:      "{var:raw",
			pos:        0,
			prev:       0,
			atSOL:      true,
			wantResult: InterpResult{Kind: InterpError, Name: "var", Error: "unclosed interpolation: {var:"},
			wantEnd:    8,
		},
		{
			name:       "invalid modifier",
			input:      "{var:foo}",
			pos:        0,
			prev:       0,
			atSOL:      true,
			wantResult: InterpResult{Kind: InterpError, Name: "var", Error: "invalid modifier ':foo', expected ':raw'"},
			wantEnd:    9,
		},
		{
			name:       "empty identifier",
			input:      "{}",
			pos:        0,
			prev:       0,
			atSOL:      true,
			wantResult: InterpResult{Kind: InterpNotInterp},
			wantEnd:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, end := ScanInterpolation(tt.input, tt.pos, tt.prev, tt.atSOL)

			if result.Kind != tt.wantResult.Kind {
				t.Errorf("ScanInterpolation() Kind = %v, want %v", result.Kind, tt.wantResult.Kind)
			}
			if result.Name != tt.wantResult.Name {
				t.Errorf("ScanInterpolation() Name = %q, want %q", result.Name, tt.wantResult.Name)
			}
			if result.Raw != tt.wantResult.Raw {
				t.Errorf("ScanInterpolation() Raw = %v, want %v", result.Raw, tt.wantResult.Raw)
			}
			if tt.wantResult.Error != "" && result.Error != tt.wantResult.Error {
				t.Errorf("ScanInterpolation() Error = %q, want %q", result.Error, tt.wantResult.Error)
			}
			if end != tt.wantEnd {
				t.Errorf("ScanInterpolation() end = %d, want %d", end, tt.wantEnd)
			}
		})
	}
}

func TestScanEscapedCloseBrace(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		pos     int
		isEsc   bool
		wantEnd int
	}{
		{
			name:    "double close brace",
			input:   "}}",
			pos:     0,
			isEsc:   true,
			wantEnd: 2,
		},
		{
			name:    "double close brace mid-string",
			input:   "x}}y",
			pos:     1,
			isEsc:   true,
			wantEnd: 3,
		},
		{
			name:    "single close brace",
			input:   "}x",
			pos:     0,
			isEsc:   false,
			wantEnd: 0,
		},
		{
			name:    "close brace at end",
			input:   "}",
			pos:     0,
			isEsc:   false,
			wantEnd: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEsc, end := ScanEscapedCloseBrace(tt.input, tt.pos)
			if isEsc != tt.isEsc {
				t.Errorf("ScanEscapedCloseBrace() isEsc = %v, want %v", isEsc, tt.isEsc)
			}
			if end != tt.wantEnd {
				t.Errorf("ScanEscapedCloseBrace() end = %d, want %d", end, tt.wantEnd)
			}
		})
	}
}

func TestInterpResultKindString(t *testing.T) {
	tests := []struct {
		kind InterpResultKind
		want string
	}{
		{InterpValid, "Valid"},
		{InterpEscapedOpen, "EscapedOpen"},
		{InterpNotInterp, "NotInterp"},
		{InterpError, "Error"},
		{InterpResultKind(99), "InterpResultKind(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.want {
				t.Errorf("InterpResultKind.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
