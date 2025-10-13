package chrononode

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dop251/goja"
)

//go:embed dist/naturalTime.bundle.js
var naturaltimeJavaScript string

// Parser provides natural language parsing capabilities for time expressions.
// It uses a JavaScript implementation embedded in the Go binary.
type Parser struct {
	runtime        *goja.Runtime
	parseRangeFunc goja.Callable
	parseDateFunc  goja.Callable
	thisContext    goja.Value
	parseFunc      goja.Callable
}

// New creates a new natural time expression parser.
// It initializes the JavaScript runtime and prepares the parsing functions.
//
// Returns:
//   - An initialized Parser
//   - An error if initialization fails
func New() (*Parser, error) {
	runtime := goja.New()

	// Compile and run the embedded JavaScript
	program, err := goja.Compile("naturaltime.js", naturaltimeJavaScript, false)
	if err != nil {
		return nil, fmt.Errorf("failed to compile naturaltime JavaScript: %w", err)
	}

	_, err = runtime.RunProgram(program)
	if err != nil {
		return nil, fmt.Errorf("failed to run naturaltime JavaScript: %w", err)
	}

	// Extract the JavaScript object and its methods
	jsObject := runtime.Get("naturaltime").ToObject(runtime)

	parseRangeFunc, ok := goja.AssertFunction(jsObject.Get("parseRange"))
	if !ok {
		return nil, fmt.Errorf("failed to get 'parseRange' function from JavaScript")
	}

	parseDateFunc, ok := goja.AssertFunction(jsObject.Get("parseDate"))
	if !ok {
		return nil, fmt.Errorf("failed to get 'parseDate' function from JavaScript")
	}

	parseFunc, ok := goja.AssertFunction(jsObject.Get("parse"))
	if !ok {
		return nil, fmt.Errorf("failed to get 'parse' function from JavaScript")
	}

	return &Parser{
		thisContext:    runtime.ToValue(map[string]interface{}{}),
		runtime:        runtime,
		parseRangeFunc: parseRangeFunc,
		parseDateFunc:  parseDateFunc,
		parseFunc:      parseFunc,
	}, nil
}

type timeResult struct {
	Text             string
	Index            int
	Time             time.Time
	MicrosoftResults string
}

func (p *Parser) Parse(expr string, base time.Time) ([]timeResult, error) {
	result, err := p.parseFunc(p.thisContext, p.runtime.ToValue(expr), p.runtime.ToValue(base.Format(time.DateTime)))

	if err != nil {
		return nil, fmt.Errorf("failed to parse expression %q: %w", expr, err)
	}

	// Convert JavaScript result to Go
	jsonBytes, err := json.Marshal(result.Export())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal results for expression %q: %w", expr, err)
	}

	var timeResults []timeResult
	err = json.Unmarshal(jsonBytes, &timeResults)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal results for expression %q: %w", expr, err)
	}
	return timeResults, nil
}

// ParseDate parses a natural language date expression and returns the corresponding time.
//
// Parameters:
//   - expr: The natural language expression to parse
//   - base: The reference time for relative expressions
//
// Returns:
//   - A pointer to the parsed time.Time, or nil if the expression could not be parsed
//   - An error if parsing fails
func (p *Parser) ParseDate(expr string, base time.Time) (*time.Time, error) {
	result, err := p.parseDateFunc(p.thisContext, p.runtime.ToValue(expr), p.runtime.ToValue(base.Format(time.RFC3339)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse date expression %q: %w", expr, err)
	}

	switch parsedValue := result.Export().(type) {
	case time.Time:
		return &parsedValue, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("unexpected result type when parsing date expression %q", expr)
	}
}
