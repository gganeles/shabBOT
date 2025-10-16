#!/usr/bin/env bash

# Test script for the deno-runner

cd "$(dirname "$0")"

echo "Building deno-runner..."
cargo build --quiet

echo "Starting deno-runner and testing..."

# Start the deno-runner in the background
./target/debug/deno-runner &
RUNNER_PID=$!

# Give it a moment to start
sleep 1

# Test ping
echo '{"type":"ping"}' | ./target/debug/deno-runner &
RUNNER_PID2=$!
sleep 1

# Test date parsing
echo '{"type":"parse_date","expression":"tomorrow at 3pm","reference_date":null}' | ./target/debug/deno-runner &
RUNNER_PID3=$!
sleep 2

# Cleanup
kill $RUNNER_PID $RUNNER_PID2 $RUNNER_PID3 2>/dev/null

echo "Test complete!"
