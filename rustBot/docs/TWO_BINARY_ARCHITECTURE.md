# Two-Binary Architecture for WhatsApp Bot + Deno Runtime

## Problem
The `whatsapp-rust` and `deno_runtime` crates have conflicting SQLite dependencies that cannot coexist in the same binary.

## Solution
Split the application into **two separate binaries** that communicate via IPC:

### 1. `rustWABot` (Main WhatsApp Bot)
- **Location**: `/home/gabe/code/shabBOT/rustBot/`
- **Dependencies**: whatsapp-rust, diesel, etc.
- **Role**: Handles WhatsApp connection, message handling, command routing

### 2. `deno-runner` (JavaScript/TypeScript Executor)
- **Location**: `/home/gabe/code/shabBOT/rustBot/deno-runner/`
- **Dependencies**: deno_runtime (with compatible SQLite version)
- **Role**: Executes JavaScript/TypeScript code on demand

## Communication Protocol

The two binaries communicate via **JSON over stdin/stdout**:

### Request Format (rustWABot → deno-runner)
```json
{"type": "eval", "code": "console.log('Hello from Deno!')"}
{"type": "run_file", "path": "/path/to/script.js"}
{"type": "ping"}
```

### Response Format (deno-runner → rustWABot)
```json
{"status": "success", "result": "execution result"}
{"status": "error", "message": "error description"}
```

## Usage from rustWABot

```rust
use crate::deno_client::DenoRunner;

// Start the deno-runner subprocess
let mut deno = DenoRunner::new()?;

// Check if it's running
if deno.ping()? {
    // Evaluate JavaScript code
    let result = deno.eval("2 + 2")?;
    println!("Result: {}", result);
    
    // Run a JavaScript file
    let output = deno.run_file("./scripts/process_data.js")?;
    println!("Output: {}", output);
}
```

## Building

```bash
# Build the main WhatsApp bot
cd /home/gabe/code/shabBOT/rustBot
cargo build --bin rustWABot

# Build the deno-runner (in separate directory)
cd /home/gabe/code/shabBOT/rustBot/deno-runner
cargo build --bin deno-runner
```

## Running

```bash
# The main bot will automatically start the deno-runner when needed
cd /home/gabe/code/shabBOT/rustBot
./target/debug/rustWABot
```

The `DenoRunner` struct will spawn `./target/debug/deno-runner` as a subprocess.

## Files Created

1. `/home/gabe/code/shabBOT/rustBot/deno-runner/` - Standalone deno binary project
   - `Cargo.toml` - Dependencies for deno_runtime
   - `src/main.rs` - JSON IPC server
   
2. `/home/gabe/code/shabBOT/rustBot/deno_client.rs` - IPC client module
   - `DenoRunner` struct for communication
   - Helper methods: `eval()`, `run_file()`, `ping()`

## Benefits

✅ **No dependency conflicts** - Each binary has independent dependencies  
✅ **Process isolation** - Deno crashes won't affect the WhatsApp bot  
✅ **Simple protocol** - JSON over pipes is easy to debug  
✅ **Extensible** - Can add more request/response types as needed  
✅ **Language agnostic** - Could replace deno-runner with Python/Node later
