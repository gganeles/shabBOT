# Test Suite

This directory contains integration tests for rustWABot.

## Running Tests

### Prerequisites

Before running tests, ensure:

1. **Build the deno-runner binary:**
   ```bash
   cd deno-runner
   cargo build
   cd ..
   ```

2. **Copy required files to main directory:**
   ```bash
   cp deno-runner/target/debug/deno-runner target/debug/
   cp deno-runner/chrono.ts .
   ```

3. **Cache Deno dependencies:**
   ```bash
   deno cache chrono.ts
   ```

### Running All Tests

```bash
cargo test
```

### Running Specific Tests

```bash
# Run datetime extraction tests
cargo test --test test_datetime_extract

# Run with output visible
cargo test --test test_datetime_extract -- --nocapture
```

## Test Files

- `test_datetime_extract.rs` - Tests for the `extract_datetime_from_text()` utility function
  - Tests natural language date parsing
  - Tests text extraction (removing datetime from input)
  - Tests error handling for invalid inputs

## Adding New Tests

1. Create a new test file in this directory: `test_<feature>.rs`
2. Import needed modules from the library:
   ```rust
   use rustWABot::utils::*;
   use rustWABot::commands::*;
   ```
3. Write test functions with the `#[test]` attribute
4. Run `cargo test --test test_<feature>` to verify

## Notes

- Tests use the library interface (`rustWABot::*`) defined in `lib.rs`
- Integration tests run as separate binaries and have access to public modules
- The deno-runner subprocess is started automatically when needed
- Each test file runs in isolation with its own process
