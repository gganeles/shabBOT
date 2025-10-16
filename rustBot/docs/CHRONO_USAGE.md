# Using Chrono Date Parser in Router

## Overview

The router now has access to natural language date/time parsing via the `parse_natural_time()` function, which uses the `chrono-node` library through the deno-runner subprocess.

## Usage in router.rs

```rust
use crate::router::parse_natural_time;

// Parse a natural language date expression
match parse_natural_time("tomorrow at 3pm", None) {
    Ok(results) => {
        for parsed in results {
            println!("Found: {} at {}", parsed.text, parsed.time.unwrap_or_default());
            // parsed.text contains the matched text like "tomorrow at 3pm"
            // parsed.time contains the ISO timestamp like "2025-10-16T15:00:00.000Z"
            // parsed.index is the position in the string
        }
    }
    Err(e) => eprintln!("Parse error: {}", e),
}

// With a custom reference date
match parse_natural_time("next Friday", Some("2025-10-15T12:00:00Z")) {
    Ok(results) => { /* ... */ }
    Err(e) => { /* ... */ }
}
```

## ParsedDate Structure

```rust
pub struct ParsedDate {
    pub text: String,           // The matched text from the input
    pub index: i32,             // Position in the input string
    pub time: Option<String>,   // ISO 8601 timestamp (e.g., "2025-10-16T15:00:00.000Z")
}
```

## Examples of Supported Expressions

- "tomorrow at 3pm"
- "next Friday"
- "in 2 hours"
- "5pm"
- "October 20th"
- "next week Tuesday"
- "in 30 minutes"
- "tomorrow morning"

## Integration with Reminder Commands

Example of using this in a reminder command:

```rust
fn remind_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str, timezone: &str) -> String {
    // Extract the time expression from the prompt
    // e.g., "!remind me tomorrow at 3pm to buy milk"
    
    match parse_natural_time(prompt, None) {
        Ok(results) if !results.is_empty() => {
            let parsed = &results[0];
            let reminder_time = parsed.time.as_ref().unwrap();
            
            // Convert to Unix timestamp
            // let timestamp = chrono::DateTime::parse_from_rfc3339(reminder_time)
            //     .unwrap()
            //     .timestamp();
            
            // Save reminder to database
            // ...
            
            format!("✅ Reminder set for {} ({})", parsed.text, reminder_time)
        }
        Ok(_) => "❌ Could not parse a date/time from your message".to_string(),
        Err(e) => format!("❌ Error parsing date: {}", e),
    }
}
```

## Requirements

- The deno-runner binary must be built: `cd deno-runner && cargo build`
- The deno-runner binary must be accessible at `./target/debug/deno-runner` relative to the main bot's working directory
- Deno must be installed on the system for the date parsing to work (the deno-runner calls `deno eval` internally)

## Installing Deno

```bash
# On Linux/macOS
curl -fsSL https://deno.land/install.sh | sh

# On Arch Linux
sudo pacman -S deno

# On macOS with Homebrew
brew install deno
```

## Architecture

```
router.rs
   ↓ calls parse_natural_time()
   ↓
deno_client.rs (DenoRunner)
   ↓ sends JSON request via stdin
   ↓
deno-runner binary
   ↓ calls `deno eval` with chrono.ts
   ↓
chrono.ts (imports npm:chrono-node)
   ↓ returns JSON results
   ↓
back to router.rs as Vec<ParsedDate>
```

## Troubleshooting

If date parsing fails:
1. Check that deno-runner is built: `ls -la target/debug/deno-runner`
2. Check that Deno is installed: `deno --version`
3. Check the stderr output of the main bot for Deno-related errors
4. Test chrono.ts directly: `cd deno-runner && deno run --allow-net chrono.ts "tomorrow"`
