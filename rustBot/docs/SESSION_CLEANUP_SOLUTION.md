# Automatic Session Cleanup for UntrustedIdentity Errors

## Problem
The bot was experiencing persistent "UntrustedIdentity" decryption errors:
```
Message from 972587120601.0 failed to decrypt
Candidate session 0 failed with 'untrusted identity for address 972587120601.0'
Batch session decrypt failed (type: pkmsg): UntrustedIdentity(ProtocolAddress { name: "972587120601", device_id: DeviceId(0) })
```

This happens when a contact reinstalls WhatsApp, changes devices, or clears their app data. The Signal protocol detects that the stored identity key doesn't match the new one being sent.

## Solution
Implemented an automatic session cleanup mechanism that:

1. **Detects UntrustedIdentity errors** in the `Event::UndecryptableMessage` handler
2. **Automatically deletes corrupted sessions and identities** from the database
3. **Allows the library to re-establish** a fresh session with the new identity

## Implementation

### Files Modified/Created:

1. **`session_cleanup.rs`** (new)
   - `cleanup_user_sessions()` - Deletes sessions and identities for a user from the SQLite database using Diesel
   - `extract_phone_number()` - Extracts phone number from JID format

2. **`main.rs`** (modified)
   - Added `mod session_cleanup`
   - Cloned backend Arc for use in event handler
   - Updated `Event::UndecryptableMessage` handler to automatically clean up sessions

3. **`Cargo.toml`** (no changes needed)
   - Already has `diesel` with SQLite support

## How It Works

1. When an `UndecryptableMessage` event occurs:
   - Extract the sender's phone number from the JID
   - Log the decryption failure for debugging
   
2. Automatic cleanup:
   - Delete all sessions for that user: `DELETE FROM sessions WHERE address LIKE '972587120601%'`
   - Delete all identity keys for that user: `DELETE FROM identities WHERE address LIKE '972587120601%'`
   
3. Re-establishment:
   - The whatsapp-rust library automatically creates a new session with the contact's current identity
   - Future messages from that contact will decrypt successfully

## Testing

To test the fix:
1. Run `cargo run` to start the bot
2. When the contact sends a message, the error will be logged
3. The cleanup will execute automatically
4. The next message from that contact should decrypt successfully

## Database Tables Affected

- **sessions**: Stores Signal protocol session state
- **identities**: Stores contact identity keys for trust verification

## Benefits

- ✅ Automatic recovery from identity changes
- ✅ No manual database intervention required
- ✅ Graceful handling of device changes
- ✅ Detailed logging for troubleshooting
- ✅ Uses existing backend Arc (clean architecture)
