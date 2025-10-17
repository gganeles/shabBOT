package chrononode

import (
    "fmt"
    "os/exec"
    "testing"
    "time"
)

func TestCommandExecution(t *testing.T) {
    // Test the exact command
    cmd := exec.Command("bun", "/home/opc/shabBOT/chrono-node/naturalTime.js", "me in 5 minutes", "2025-10-17T16:44:50+13:00")
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("Command failed: %v\nOutput: %s", err, string(output))
    }
    
    fmt.Printf("Output: %s\n", string(output))
}

func TestParser(t *testing.T) {
    parser, err := New()
    if err != nil {
        t.Fatalf("Failed to create parser: %v", err)
    }
    
    baseTime, _ := time.Parse(time.RFC3339, "2025-10-17T16:44:50+13:00")
    
    result, err := parser.ParseDate("me in 5 minutes", baseTime)
    if err != nil {
        t.Fatalf("Parse failed: %v", err)
    }
    
    if result == nil {
        t.Fatal("Expected result, got nil")
    }
    
    fmt.Printf("Parsed time: %v\n", result)
}
