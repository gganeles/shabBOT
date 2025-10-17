package chrononode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// Parser provides natural language parsing capabilities for time expressions.
// It uses a JavaScript implementation via Bun runtime.
type Parser struct {
	scriptPath string
}

// New creates a new natural time expression parser.
// It determines the path to the naturalTime.js script.
//
// Returns:
//   - An initialized Parser
//   - An error if initialization fails
func New() (*Parser, error) {
	// Get the directory of this Go file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}

	// Get the directory containing this file
	dir := filepath.Dir(filename)
	// Use compiled binary instead of .js file for better performance
	scriptPath := filepath.Join(dir, "naturalTime")

	return &Parser{
		scriptPath: scriptPath,
	}, nil
}

type timeResult struct {
	Text             string
	Index            int
	Time             time.Time
	MicrosoftResults string
}

// execBun executes the compiled naturalTime binary
func (p *Parser) execBun(expr string, base time.Time) ([]byte, error) {
	baseStr := base.Format(time.DateTime)

	// Call the compiled binary directly (no bun runtime needed)
	cmd := exec.Command(p.scriptPath, expr, baseStr)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

func (p *Parser) Parse(expr string, base time.Time) ([]timeResult, error) {
	output, err := p.execBun(expr, base)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression %q: %w", expr, err)
	}

	var timeResults []timeResult
	err = json.Unmarshal(output, &timeResults)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal results for expression %q: %w (output: %s)", expr, err, string(output))
	}
	return timeResults, nil
}

// ParseDate parses a natural language date expression and returns the corresponding time.
// It returns the first parsed time from the Parse function.
//
// Parameters:
//   - expr: The natural language expression to parse
//   - base: The reference time for relative expressions
//
// Returns:
//   - A pointer to the parsed time.Time, or nil if the expression could not be parsed
//   - An error if parsing fails
func (p *Parser) ParseDate(expr string, base time.Time) (*time.Time, error) {
	results, err := p.Parse(expr, base)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0].Time, nil
}
