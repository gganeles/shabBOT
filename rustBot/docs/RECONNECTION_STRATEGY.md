# WhatsApp WebSocket Reconnection Strategy

## Overview

This document explains the auto-reconnection mechanism implemented in the Rust WhatsApp bot to handle connection failures and network interruptions.

## Problem

WhatsApp WebSocket connections can fail due to:
- Network interruptions
- Server-side disconnections
- Connection timeouts
- "Connection reset by peer" errors (OS error 104)

## Solution: Exponential Backoff with Auto-Reconnect

### Implementation Details

The bot implements an **infinite reconnection loop** with **exponential backoff** inspired by production WhatsApp libraries like:
- **Baileys** (TypeScript): Uses keep-alive mechanism and connection state tracking
- **whatsmeow** (Go): Implements `AutoReconnectErrors` counter and exponential delays

### Key Features

1. **Exponential Backoff**
   - Delay formula: `min(2^attempt, 60)` seconds
   - Example sequence: 2s → 4s → 8s → 16s → 32s → 60s (max)
   - Prevents overwhelming the server with rapid reconnection attempts

2. **Maximum Retry Limit**
   - Default: 10 attempts
   - Configurable via `MAX_RECONNECT_ATTEMPTS`
   - Prevents infinite failures from draining resources

3. **Automatic Recovery**
   - Bot automatically reconnects on any disconnect
   - No manual intervention required
   - Maintains service availability

4. **Connection State Logging**
   - Logs each connection attempt with attempt number
   - Logs delay before next retry
   - Helps with debugging and monitoring

## Code Structure

```rust
loop {
    // 1. Attempt to connect and run bot
    let bot_handle = match bot.run().await {
        Ok(handle) => handle,
        Err(e) => {
            // Handle startup failure with backoff
            reconnect_attempts += 1;
            let delay = std::cmp::min(2u64.pow(reconnect_attempts), 60);
            tokio::time::sleep(Duration::from_secs(delay)).await;
            continue;
        }
    };

    // 2. Wait for disconnection
    bot_handle.await;
    
    // 3. Reconnect with exponential backoff
    reconnect_attempts += 1;
    let delay = std::cmp::min(2u64.pow(reconnect_attempts), 60);
    tokio::time::sleep(Duration::from_secs(delay)).await;
}
```

## Error Scenarios

### Scenario 1: Connection Reset by Peer (Error 104)
**Cause**: Server closes connection unexpectedly
**Response**: 
- Log error
- Wait with exponential backoff
- Attempt reconnection
- Reset attempt counter after successful reconnection

### Scenario 2: Network Timeout
**Cause**: Poor network conditions
**Response**:
- Detect timeout via bot.run() failure
- Apply exponential backoff
- Retry connection

### Scenario 3: Authentication Failure
**Cause**: Invalid credentials or session expired
**Response**:
- Log authentication error
- Stop after MAX_RECONNECT_ATTEMPTS
- User must re-authenticate (scan QR code)

## Configuration

### Tunable Parameters

```rust
const MAX_RECONNECT_ATTEMPTS: u32 = 10;  // Max retries before giving up
```

To adjust:
1. **Increase** for flaky networks (e.g., 20)
2. **Decrease** for faster failure detection (e.g., 5)

### Delay Formula

Current: `min(2^attempt, 60)` seconds

Alternative strategies:
- Linear: `attempt * 5` seconds (5s, 10s, 15s...)
- Custom cap: `min(2^attempt, 30)` for faster retries

## Best Practices

1. **Monitor Logs**: Check for frequent reconnections indicating network issues
2. **Alert on Max Attempts**: Set up monitoring for "Max reconnection attempts reached"
3. **QR Code Ready**: Keep device ready to re-scan QR if session expires
4. **Stable Network**: Use stable server hosting with good uptime

## Related Features

### Reminder Routine
- Runs independently in background task
- Continues checking reminders even during reconnection
- Uses separate database connection

### Keep-Alive Mechanism
The underlying `whatsapp_rust` library handles:
- Periodic ping/pong messages
- Connection health checks
- Dead connection detection

## Future Improvements

Potential enhancements:
1. **Smart backoff reset**: Reset counter after X minutes of stable connection
2. **Connection health metrics**: Track uptime, failures, recovery time
3. **Notification system**: Alert admin on repeated failures
4. **Graceful degradation**: Continue basic operations during reconnection

## Troubleshooting

### Bot keeps reconnecting every few seconds
- Check network stability
- Verify WhatsApp account is not banned
- Ensure QR code session is valid

### Max attempts reached too quickly
- Increase `MAX_RECONNECT_ATTEMPTS`
- Check for authentication issues
- Verify server has internet access

### Long delays between reconnections
- Normal behavior with exponential backoff
- Consider reducing max delay cap (currently 60s)

## References

- [Baileys Socket Implementation](https://github.com/WhiskeySockets/Baileys/blob/master/src/Socket/socket.ts)
- [whatsmeow Client Auto-Reconnect](https://github.com/tulir/whatsmeow/blob/main/client.go)
- [WebSocket Best Practices](https://datatracker.ietf.org/doc/html/rfc6455)
