# Timezone-Aware Date/Time Parsing Implementation

## Overview

This document explains the timezone-aware datetime parsing implementation in the Rust WhatsApp bot, which matches the Go implementation's behavior.

## Architecture

### Key Components

1. **Database Schema** (`schema.rs`)
   - `chats` table stores timezone per chat (IANA timezone names)
   - Default timezone: `Asia/Jerusalem`

2. **Timezone Retrieval** (`database.rs`)
   - `get_chat_timezone()` - Gets timezone from database with fallback

3. **Date Parsing** (`utils.rs`)
   - `extract_datetime_from_text()` - Timezone-aware datetime extraction
   - Uses chrono-node (JavaScript) via Deno for natural language parsing

4. **Command Integration** (`commands/`)
   - All time-based commands now accept timezone parameter
   - Times displayed in user's local timezone

## How It Works

### 1. Timezone Context Flow

```
User sends message
    ↓
Router gets chat timezone from DB
    ↓
Command receives timezone parameter
    ↓
extract_datetime_from_text() uses timezone as reference
    ↓
Parsed time stored as UTC timestamp
    ↓
Display converts UTC back to user's timezone
```

### 2. Date Parsing Process

```rust
pub fn extract_datetime_from_text(
    text: &str,
    timezone_name: &str,
) -> Result<(String, DateTime<Utc>), String>
```

**Steps:**
1. Parse timezone string into `chrono_tz::Tz` object
2. Get current time in user's timezone as reference
3. Pass reference date to chrono-node for natural language parsing
4. Parse result is in user's local context (e.g., "tomorrow" means tomorrow in Tel Aviv, not UTC)
5. Convert parsed time to UTC for storage
6. Remove datetime text from message and return cleaned message

### 3. Example Flow

**User in Jerusalem (Asia/Jerusalem, UTC+2):**
```
Input: "!remind buy milk tomorrow at 3pm"
Current time: 2025-01-15 10:00:00 Jerusalem time

Process:
1. Get timezone: "Asia/Jerusalem"
2. Reference date: 2025-01-15T08:00:00Z (10:00 Jerusalem = 08:00 UTC)
3. Parse "tomorrow at 3pm" with Jerusalem context
4. Result: 2025-01-16 15:00:00 Jerusalem time
5. Convert to UTC: 2025-01-16T13:00:00Z
6. Store timestamp: 1737033600
7. Display: "Ok, "buy milk" scheduled for Thu Jan 16 @ 03:00 PM"
```

## Go Implementation Comparison

### Go Version (`commands/utils.go`)

```go
func parseTime(s string, timezone *time.Location) (int64, string, error) {
    loc := time.UTC
    if timezone != nil {
        loc = timezone
    }
    
    now := time.Now().In(loc)  // Get current time in user's timezone
    
    parsed, err := dateParse.New()
    parsedDate, err := parsed.Parse(s, now)  // Parse with timezone context
    
    // ... extract message ...
    
    cleanedTime, err := time.ParseInLocation(time.DateTime, timeStr, loc)
    return cleanedTime.Unix(), strings.TrimSpace(message), nil
}
```

### Rust Version (`utils.rs`)

```rust
pub fn extract_datetime_from_text(
    text: &str,
    timezone_name: &str,
) -> Result<(String, DateTime<Utc>), String> {
    let tz: Tz = timezone_name.parse()?;
    
    // Get current time in the target timezone as reference
    let now_in_tz = chrono::Utc::now().with_timezone(&tz);
    let reference_date = Some(now_in_tz.to_rfc3339());
    
    // Parse using chrono-node with timezone-aware reference date
    let parsed_dates = parse_natural_time(text, reference_date.as_deref())?;
    
    // Parse ISO 8601 datetime string and convert to UTC
    let datetime = DateTime::parse_from_rfc3339(time_str)?
        .with_timezone(&Utc);
    
    // ... extract message ...
    
    Ok((cleaned_text, datetime))
}
```

**Key Similarities:**
- Both get current time in user's timezone
- Both pass timezone-aware reference to date parser
- Both store as Unix timestamp (seconds since epoch)
- Both clean extracted text the same way

## Updated Commands

### Commands with Timezone Support

1. **`!remind`** (`commands/reminders.rs`)
   - Parses datetime with user's timezone
   - Displays confirmation in user's local time
   
2. **`!reminders`** (`commands/reminders.rs`)
   - Lists reminders with times in user's timezone
   - Formats: "Mon @ 03:00 PM"

3. **`!snooze`** (`commands/reminders.rs`)
   - Parses snooze time with timezone context

4. **`!send`** (`commands/scheduled.rs`)
   - Schedules messages with timezone-aware parsing

### Display Format

All times shown to users are converted to their local timezone:

```rust
let tz: Tz = timezone.parse().unwrap_or(chrono_tz::Asia::Jerusalem);
let dt = DateTime::<Utc>::from_timestamp(timestamp, 0)
    .unwrap()
    .with_timezone(&tz);  // Convert to user's timezone
let time_str = dt.format("%a @ %I:%M %p").to_string();
```

## Timezone Storage

### Default Timezone

If no timezone is set: `Asia/Jerusalem`

### Setting Timezone

Timezone is automatically set when user sets location:

```rust
pub fn set_chat_location(
    conn: &mut SqliteConnection,
    chat_id: &str,
    location: &str,
    timezone: &str,
) -> Result<(), Box<dyn std::error::Error>>
```

Uses `GeoName` database to map locations to IANA timezones:
- Haifa → Asia/Jerusalem
- New York → America/New_York
- Tel Aviv → Asia/Jerusalem

## Benefits

### 1. Accurate Time Interpretation
- "tomorrow" means tomorrow in user's local time
- "3pm" means 3pm in user's timezone
- No confusion with UTC offsets

### 2. Daylight Saving Time Support
- IANA timezones handle DST automatically
- "tomorrow at 3pm" during DST transition works correctly

### 3. International Support
- Users in different timezones see their local times
- Same bot serves users globally

### 4. Matches Go Behavior
- Drop-in replacement for Go implementation
- Identical parsing logic and results

## Testing

### Test Scenarios

1. **Basic Time Parsing**
```rust
// User in Jerusalem (UTC+2)
extract_datetime_from_text("tomorrow at 3pm", "Asia/Jerusalem")
// Should parse as tomorrow 15:00 Jerusalem time
```

2. **Cross-Timezone**
```rust
// User in New York (UTC-5)
extract_datetime_from_text("in 5 hours", "America/New_York")
// Should add 5 hours to current New York time
```

3. **Relative Times**
```rust
extract_datetime_from_text("next Monday at noon", "Asia/Jerusalem")
// Should find next Monday in Jerusalem timezone
```

4. **Display Conversion**
```rust
// Stored UTC: 2025-01-16T13:00:00Z
// Display in Jerusalem (UTC+2): Thu Jan 16 @ 03:00 PM
// Display in New York (UTC-5): Thu Jan 16 @ 08:00 AM
```

## Dependencies

- **chrono** - DateTime manipulation and UTC timestamps
- **chrono-tz** - IANA timezone database and conversion
- **Deno + chrono-node** - Natural language date parsing

## Edge Cases Handled

1. **Invalid Timezone**
   - Falls back to `Asia/Jerusalem`
   - Returns error message to user

2. **No Timezone Stored**
   - Uses default `Asia/Jerusalem`
   - Should be set via `!shablocation` command

3. **Ambiguous Times (DST)**
   - chrono-tz handles ambiguous times during DST transitions
   - Prefers standard time interpretation

4. **Past Dates**
   - chrono-node interprets relative to reference date
   - "tomorrow" always means next day, never past

## Migration from Old Code

### Before (No Timezone Awareness)
```rust
extract_datetime_from_text(&text, None)  // No timezone context
```

### After (Timezone Aware)
```rust
let timezone = get_chat_timezone(db, chat_id);
extract_datetime_from_text(&text, &timezone)  // Timezone context
```

## Future Enhancements

1. **Timezone Command**
   - Add `!timezone` to manually set timezone
   - Show current timezone setting

2. **Multi-Timezone Display**
   - Show time in multiple timezones for international groups

3. **Smart Detection**
   - Detect timezone from phone number country code

4. **Performance Optimization**
   - Cache parsed timezones
   - Reuse Deno runner instances

## References

- [IANA Time Zone Database](https://www.iana.org/time-zones)
- [chrono-tz Documentation](https://docs.rs/chrono-tz/)
- [chrono-node NPM Package](https://www.npmjs.com/package/chrono-node)
- Go Implementation: `/commands/utils.go:parseTime()`
