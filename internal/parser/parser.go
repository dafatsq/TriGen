package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"triton-config-studio/internal/model"
)

type tokenType int

const (
	tokenEOF tokenType = iota
	tokenIdent
	tokenString
	tokenInt
	tokenFloat
	tokenColon
	tokenOpenBrace
	tokenCloseBrace
	tokenOpenBracket
	tokenCloseBracket
	tokenComma
)

func (t tokenType) String() string {
	switch t {
	case tokenEOF:
		return "EOF"
	case tokenIdent:
		return "IDENT"
	case tokenString:
		return "STRING"
	case tokenInt:
		return "INT"
	case tokenFloat:
		return "FLOAT"
	case tokenColon:
		return ":"
	case tokenOpenBrace:
		return "{"
	case tokenCloseBrace:
		return "}"
	case tokenOpenBracket:
		return "["
	case tokenCloseBracket:
		return "]"
	case tokenComma:
		return ","
	}
	return "UNKNOWN"
}

type token struct {
	Type  tokenType
	Value string
	Line  int
}

type lexer struct {
	input []rune
	pos   int
	line  int
}

func newLexer(input string) *lexer {
	return &lexer{
		input: []rune(input),
		pos:   0,
		line:  1,
	}
}

func (l *lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	r := l.input[l.pos]
	l.pos++
	if r == '\n' {
		l.line++
	}
	return r
}

func (l *lexer) tokenize() ([]token, error) {
	var tokens []token
	for {
		l.skipWhitespaceAndComments()
		r := l.peek()
		if r == 0 {
			tokens = append(tokens, token{Type: tokenEOF, Line: l.line})
			break
		}

		line := l.line
		switch r {
		case '{':
			l.next()
			tokens = append(tokens, token{Type: tokenOpenBrace, Value: "{", Line: line})
		case '}':
			l.next()
			tokens = append(tokens, token{Type: tokenCloseBrace, Value: "}", Line: line})
		case '[':
			l.next()
			tokens = append(tokens, token{Type: tokenOpenBracket, Value: "[", Line: line})
		case ']':
			l.next()
			tokens = append(tokens, token{Type: tokenCloseBracket, Value: "]", Line: line})
		case ':':
			l.next()
			tokens = append(tokens, token{Type: tokenColon, Value: ":", Line: line})
		case ',':
			l.next()
			tokens = append(tokens, token{Type: tokenComma, Value: ",", Line: line})
		case '"', '\'':
			str, err := l.readString(r)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token{Type: tokenString, Value: str, Line: line})
		default:
			if unicode.IsDigit(r) || r == '-' || r == '+' || r == '.' {
				tok, err := l.readNumber()
				if err != nil {
					return nil, err
				}
				tokens = append(tokens, tok)
			} else if isIdentStart(r) {
				ident := l.readIdent()
				tokens = append(tokens, token{Type: tokenIdent, Value: ident, Line: line})
			} else {
				return nil, fmt.Errorf("line %d: unexpected character %q", l.line, r)
			}
		}
	}
	return tokens, nil
}

func (l *lexer) skipWhitespaceAndComments() {
	for {
		r := l.peek()
		if unicode.IsSpace(r) {
			l.next()
		} else if r == '#' {
			// Skip comment to end of line
			for {
				nr := l.peek()
				if nr == 0 || nr == '\n' {
					break
				}
				l.next()
			}
		} else {
			break
		}
	}
}

func (l *lexer) readString(quote rune) (string, error) {
	l.next() // consume opening quote
	var sb strings.Builder
	for {
		r := l.next()
		if r == 0 {
			return "", fmt.Errorf("line %d: unterminated string literal", l.line)
		}
		if r == quote {
			break
		}
		if r == '\\' {
			nextR := l.next()
			switch nextR {
			case 'n':
				sb.WriteRune('\n')
			case 'r':
				sb.WriteRune('\r')
			case 't':
				sb.WriteRune('\t')
			case '\\':
				sb.WriteRune('\\')
			case '"':
				sb.WriteRune('"')
			case '\'':
				sb.WriteRune('\'')
			default:
				sb.WriteRune('\\')
				sb.WriteRune(nextR)
			}
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String(), nil
}

func (l *lexer) readNumber() (token, error) {
	line := l.line
	var sb strings.Builder
	isFloat := false

	// Handle initial sign
	r := l.peek()
	if r == '-' || r == '+' {
		sb.WriteRune(l.next())
	}

	for {
		r = l.peek()
		if unicode.IsDigit(r) {
			sb.WriteRune(l.next())
		} else if r == '.' {
			isFloat = true
			sb.WriteRune(l.next())
		} else if r == 'e' || r == 'E' {
			isFloat = true
			sb.WriteRune(l.next())
			nextR := l.peek()
			if nextR == '-' || nextR == '+' {
				sb.WriteRune(l.next())
			}
		} else {
			break
		}
	}

	val := sb.String()
	if val == "-" || val == "+" || val == "." {
		return token{}, fmt.Errorf("line %d: invalid number literal %q", line, val)
	}

	if isFloat {
		return token{Type: tokenFloat, Value: val, Line: line}, nil
	}
	return token{Type: tokenInt, Value: val, Line: line}, nil
}

func (l *lexer) readIdent() string {
	var sb strings.Builder
	for {
		r := l.peek()
		if isIdentPart(r) {
			sb.WriteRune(l.next())
		} else {
			break
		}
	}
	return sb.String()
}

func isIdentStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_' || r == '/' || r == '.'
}

func isIdentPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '/' || r == '.' || r == '$' || r == '{' || r == '}'
}

type parser struct {
	tokens []token
	pos    int
}

func newParser(tokens []token) *parser {
	return &parser{
		tokens: tokens,
		pos:    0,
	}
}

func (p *parser) peek() token {
	if p.pos >= len(p.tokens) {
		return token{Type: tokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *parser) next() token {
	t := p.peek()
	if t.Type != tokenEOF {
		p.pos++
	}
	return t
}

func (p *parser) consume(expected tokenType) (token, error) {
	t := p.next()
	if t.Type != expected {
		return token{}, fmt.Errorf("line %d: expected token %s, got %s (%q)", t.Line, expected, t.Type, t.Value)
	}
	return t, nil
}

func (p *parser) isEOF() bool {
	return p.peek().Type == tokenEOF
}

func parseMessage(p *parser) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	for !p.isEOF() {
		if p.peek().Type == tokenCloseBrace {
			break
		}
		// Expect key (identifier)
		keyTok, err := p.consume(tokenIdent)
		if err != nil {
			return nil, err
		}
		key := keyTok.Value

		// Consume optional colon
		if p.peek().Type == tokenColon {
			p.next()
		}

		var val interface{}
		peekTok := p.peek()
		if peekTok.Type == tokenOpenBrace {
			p.next() // consume {
			sub, err := parseMessage(p)
			if err != nil {
				return nil, err
			}
			_, err = p.consume(tokenCloseBrace)
			if err != nil {
				return nil, err
			}
			val = sub
		} else if peekTok.Type == tokenOpenBracket {
			p.next() // consume [
			list, err := parseList(p)
			if err != nil {
				return nil, err
			}
			_, err = p.consume(tokenCloseBracket)
			if err != nil {
				return nil, err
			}
			val = list
		} else {
			// Scalar value
			valTok := p.next()
			switch valTok.Type {
			case tokenString:
				val = valTok.Value
			case tokenInt:
				i, err := strconv.ParseInt(valTok.Value, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("line %d: invalid integer %s", valTok.Line, valTok.Value)
				}
				val = i
			case tokenFloat:
				f, err := strconv.ParseFloat(valTok.Value, 64)
				if err != nil {
					return nil, fmt.Errorf("line %d: invalid float %s", valTok.Line, valTok.Value)
				}
				val = f
			case tokenIdent:
				// Bool or Enum
				if valTok.Value == "true" {
					val = true
				} else if valTok.Value == "false" {
					val = false
				} else {
					val = valTok.Value // Enum string
				}
			default:
				return nil, fmt.Errorf("line %d: unexpected token %s for key %s", valTok.Line, valTok.Value, key)
			}
		}

		// Merge value
		if existing, ok := m[key]; ok {
			if slice, ok := existing.([]interface{}); ok {
				m[key] = append(slice, val)
			} else {
				m[key] = []interface{}{existing, val}
			}
		} else {
			m[key] = val
		}

		// Optional comma between fields
		if p.peek().Type == tokenComma {
			p.next()
		}
	}
	return m, nil
}

func parseList(p *parser) ([]interface{}, error) {
	var list []interface{}
	for {
		if p.peek().Type == tokenCloseBracket {
			break
		}
		var val interface{}
		peekTok := p.peek()
		if peekTok.Type == tokenOpenBrace {
			p.next() // consume {
			sub, err := parseMessage(p)
			if err != nil {
				return nil, err
			}
			_, err = p.consume(tokenCloseBrace)
			if err != nil {
				return nil, err
			}
			val = sub
		} else if peekTok.Type == tokenOpenBracket {
			p.next() // consume [
			subList, err := parseList(p)
			if err != nil {
				return nil, err
			}
			_, err = p.consume(tokenCloseBracket)
			if err != nil {
				return nil, err
			}
			val = subList
		} else {
			valTok := p.next()
			switch valTok.Type {
			case tokenString:
				val = valTok.Value
			case tokenInt:
				i, err := strconv.ParseInt(valTok.Value, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("line %d: invalid integer %s", valTok.Line, valTok.Value)
				}
				val = i
			case tokenFloat:
				f, err := strconv.ParseFloat(valTok.Value, 64)
				if err != nil {
					return nil, fmt.Errorf("line %d: invalid float %s", valTok.Line, valTok.Value)
				}
				val = f
			case tokenIdent:
				if valTok.Value == "true" {
					val = true
				} else if valTok.Value == "false" {
					val = false
				} else {
					val = valTok.Value
				}
			default:
				return nil, fmt.Errorf("line %d: unexpected token %s in list", valTok.Line, valTok.Value)
			}
		}

		list = append(list, val)

		// Optional comma between list items
		if p.peek().Type == tokenComma {
			p.next()
		}
	}
	return list, nil
}

// Parse parses a Triton config.pbtxt content into a ModelConfig struct
func Parse(content string) (*model.ModelConfig, error) {
	lex := newLexer(content)
	tokens, err := lex.tokenize()
	if err != nil {
		return nil, fmt.Errorf("lexical error: %w", err)
	}

	p := newParser(tokens)
	m, err := parseMessage(p)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	if !p.isEOF() {
		return nil, errors.New("parse error: extra tokens at end of file")
	}

	// Normalize single values that should be slices if they weren't parsed as slices.
	// Since protobuf allows writing a repeated field as a single block:
	//   input { name: "..." }
	// instead of:
	//   input [ { name: "..." } ]
	// Our parseMessage merges multiple duplicate keys into slice: []interface{}.
	// If there's only one key, it parses it as a map. Go's json.Unmarshal will FAIL
	// if it expects a slice but receives a map/object.
	// To make this fully robust, we normalize the map keys that map to slices in ModelConfig.
	normalizeSlices(m)

	// Serialize back to JSON and Unmarshal to struct
	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("json serialization error: %w", err)
	}

	var cfg model.ModelConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("struct mapping error: %w", err)
	}

	return &cfg, nil
}

func normalizeSlices(m map[string]interface{}) {
	normalizeFieldAliases(m)

	sliceKeys := map[string]bool{
		"input":                     true,
		"output":                    true,
		"instance_group":            true,
		"model_warmup":              true,
		"parameters":                true,
		"preferred_batch_size":      true,
		"priority_queue_policy":     true,
		"control_input":             true,
		"control":                   true,
		"int32_value":               true,
		"fp32_value":                true,
		"state":                     true,
		"initial_state":             true,
		"gpu_execution_accelerator": true,
		"cpu_execution_accelerator": true,
		"gpus":                      true,
		"versions":                  true,
		"inputs":                    true,
		"step":                      true,
		"input_map":                 true,
		"output_map":                true,
	}

	for k, v := range m {
		if sliceKeys[k] {
			if v != nil {
				if _, ok := v.([]interface{}); !ok {
					// Convert single item to list
					m[k] = []interface{}{v}
					v = m[k]
				}
			}
		}

		if k == "priority_queue_policy" {
			normalizePriorityQueuePolicy(v)
		}

		// Recursively normalize maps and slices of maps
		if subMap, ok := m[k].(map[string]interface{}); ok {
			normalizeSlices(subMap)
		} else if slice, ok := m[k].([]interface{}); ok {
			for _, item := range slice {
				if itemMap, ok := item.(map[string]interface{}); ok {
					normalizeSlices(itemMap)
				}
			}
		}
	}
}

func normalizeFieldAliases(m map[string]interface{}) {
	if legacyDims, ok := m["dims"]; ok {
		if _, hasShape := m["shape"]; !hasShape {
			m["shape"] = legacyDims
		}
	}
	if legacyTimeout, ok := m["timeout_microseconds"]; ok {
		if _, hasDefaultTimeout := m["default_timeout_microseconds"]; !hasDefaultTimeout {
			m["default_timeout_microseconds"] = legacyTimeout
		}
	}
	if legacyAction, ok := m["action"]; ok {
		if _, hasTimeoutAction := m["timeout_action"]; !hasTimeoutAction {
			m["timeout_action"] = legacyAction
		}
	}

	for _, key := range []string{"input_pinned_memory", "output_pinned_memory"} {
		if nested, ok := m[key].(map[string]interface{}); ok {
			if enabled, ok := nested["enable"]; ok {
				m[key] = enabled
			}
		}
	}
}

func normalizePriorityQueuePolicy(v interface{}) {
	entries, ok := v.([]interface{})
	if !ok {
		return
	}
	for _, entry := range entries {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		if key, ok := entryMap["key"]; ok {
			entryMap["priority"] = key
		}
		if value, ok := entryMap["value"]; ok {
			entryMap["queue_policy"] = value
		}
	}
}
