# Command Implementation Mapping

This document maps each JavaScript command from ResponseCmd.js to its Go implementation.

## Command Reference Table

| Command | Aliases | JS File | Go File | Status | Notes |
|---------|---------|---------|---------|--------|-------|
| help | - | ResponseCmd.js | help.go | ✅ Complete | Static text response |
| docs | - | ResponseCmd.js | docs.go | ✅ Complete | Returns GitHub link |
| gil | - | ResponseCmd.js | gil.go | ✅ Complete | Easter egg |
| count | - | ResponseCmd.js | count.go | ✅ Complete | Omer counter link |
| shabtimes | shabbattimes | ShabbatTimes.js | shabtimes.go | ⚠️ Placeholder | Needs hebcal integration |
| shablocation | shabloc, setloc, location | ShabbatTimes.js | shablocation.go | ✅ Complete | Uses database |
| needs | - | ListAndNeedsCommands.js | needs.go | ⚠️ Deprecated | Events system deprecated |
| remind | - | ScheduledSendandReminders.js | remind.go | ⚠️ Simplified | Needs better date parsing |
| reminders | rems, todo | ScheduledSendandReminders.js | remind.go | ✅ Complete | Lists from database |
| send | - | ScheduledSendandReminders.js | send.go | ⚠️ Simplified | Needs better date parsing |
| unsend | - | ScheduledSendandReminders.js | send.go | ✅ Complete | Removes last scheduled |
| scheduled | - | ScheduledSendandReminders.js | send.go | ✅ Complete | Lists scheduled messages |
| fast | fasting | Fast_checker.js | fast.go | ⚠️ Placeholder | Needs hebcal integration |
| snooze | - | ScheduledSendandReminders.js | remind.go | ⚠️ Simplified | Needs better duration parsing |
| done | - | ScheduledSendandReminders.js | remind.go | ✅ Complete | Marks reminder done |
| quickshab | - | quickShab.js | quickshab.go | ✅ Complete | Full implementation |
| update | up, num, ppl | quickShab.js | quickshab.go | ✅ Complete | Updates attendee count |
| bring | br, bringing | quickShab.js | bring.go | ✅ Complete | Sign up for category |
| assign | - | quickShab.js | bring.go | ✅ Complete | Assign to person |
| unbring | unbr | quickShab.js | bring.go | ✅ Complete | Remove assignment |
| unassign | - | quickShab.js | bring.go | ✅ Complete | Remove specific assignment |
| show | - | quickShab.js | quickshab.go | ✅ Complete | Display assignments |
| start | - | TimedStuff.js | tova.go | ⚠️ Placeholder | Needs TimedStuff review |
| end | - | TimedStuff.js | tova.go | ⚠️ Placeholder | Needs TimedStuff review |
| shop | shp, sh | shopping.js | shop.go | ✅ Complete | Add to shopping list |
| unshop | unshp, unsh | shopping.js | shop.go | ✅ Complete | Remove from list |
| shoplist | shplist, shoppinglist, etc. | shopping.js | shop.go | ✅ Complete | Display list |

## Status Legend

- ✅ **Complete**: Fully implemented with equivalent functionality
- ⚠️ **Simplified**: Core functionality implemented but missing advanced features
- ⚠️ **Placeholder**: Structure in place but needs external dependencies or additional work
- ⚠️ **Deprecated**: JavaScript version marked as deprecated

## Implementation Details by Category

### 1. Simple Static Responses (✅ Complete)
These commands return static text or simple logic without external dependencies:
- `help.go`: Static help text
- `docs.go`: GitHub link
- `gil.go`: Random link selection
- `count.go`: Omer counter link

**No additional work needed.**

### 2. Database-Backed Commands (✅ Complete)
These commands use SQLite3 for persistence:

#### Location Management
- `shablocation.go`: Store/retrieve chat location

#### QuickShab System (Meal Planning)
- `quickshab.go`: Create/update/show meal planner
- `bring.go`: Manage assignments (bring, assign, unbring, unassign)

#### Shopping Lists
- `shop.go`: Add/remove/list shopping items

#### Reminders & Scheduled Messages
- `remind.go`: Create/list/snooze/complete reminders
- `send.go`: Schedule/cancel/list messages

**All core functionality implemented. Works with SQLite3.**

### 3. Commands Needing Date/Time Parsing (⚠️ Simplified)
These commands work but use simplified date parsing:

#### Current Implementation
- `remind.go`: Basic parsing for "in X minutes/hours"
- `send.go`: Basic time parsing
- `snooze.go`: Basic duration parsing

#### What's Missing
- Natural language date parsing (like chrono-node)
- Timezone-aware date handling
- Relative dates ("tomorrow", "next Friday", etc.)
- Complex time specifications ("3pm tomorrow", "in 2 days at noon")

#### Recommended Solution
Use a Go date parsing library like:
- `github.com/olebedev/when` (Go port of chrono)
- `github.com/araddon/dateparse`
- Custom parsing with `time.Parse()` for specific formats

### 4. Commands Needing Hebrew Calendar Integration (⚠️ Placeholder)

#### `shabtimes.go`
**Current State**: Placeholder that explains what's needed

**Requirements**:
1. Hebrew calendar calculations
2. Sunset/sunrise times by location
3. Candle lighting time (18-40 min before sunset, varies by location)
4. Havdalah time (42-72 min after sunset, varies by custom)

**Recommended Solutions**:
- Call hebcal.com REST API: `https://www.hebcal.com/shabbat?cfg=json&geonameid=XXX`
- Use a Go Hebrew calendar library (if one exists)
- Port relevant parts of @hebcal/core to Go

**API Example**:
```go
import "net/http"
import "encoding/json"

type HebcalResponse struct {
    Items []struct {
        Category string `json:"category"`
        Title    string `json:"title"`
        Date     string `json:"date"`
    } `json:"items"`
}

func getShabbatTimes(location string) (*HebcalResponse, error) {
    resp, err := http.Get(fmt.Sprintf(
        "https://www.hebcal.com/shabbat?cfg=json&geo=city&city=%s", 
        location))
    // ... parse JSON response
}
```

#### `fast.go`
**Current State**: Placeholder explaining Jewish fast days

**Requirements**:
1. Hebrew calendar date checking
2. Fast day database (Gedaliah, Yom Kippur, Tevet, Esther, Tammuz, Tisha B'Av)
3. Fast start/end times by location

**Recommended Solution**: Same as shabtimes - use hebcal.com API

### 5. Commands Needing Original Implementation Review (⚠️ Placeholder)

#### `tova.go` (start/end commands)
**Issue**: Original implementation in `TimedStuff.js` not reviewed

**Current State**: Placeholder acknowledging the commands exist

**Next Steps**:
1. Read `TimedStuff.js` to understand `tovaTriggerStart()` and `tovaTriggerEnd()`
2. Implement equivalent logic in Go
3. Update `tova.go` with proper implementation

#### `needs.go`
**Issue**: Depends on deprecated events system

**Current State**: Returns message suggesting use of quickshab instead

**Next Steps**:
1. Confirm events system is being phased out
2. If still needed, implement events table in database
3. Otherwise, keep current redirect to quickshab

## Database Schema Details

The Go implementation uses a normalized SQLite3 schema:

```sql
-- Core chat settings
CREATE TABLE chats (
    chat_id TEXT PRIMARY KEY,
    location TEXT DEFAULT 'Haifa'
);

-- QuickShab meal planning
CREATE TABLE quickshab (
    chat_id TEXT PRIMARY KEY,
    number INTEGER DEFAULT 0,
    FOREIGN KEY (chat_id) REFERENCES chats(chat_id)
);

CREATE TABLE quickshab_assignments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id TEXT NOT NULL,
    category TEXT NOT NULL,
    name TEXT NOT NULL,
    FOREIGN KEY (chat_id) REFERENCES chats(chat_id)
);

-- Shopping lists
CREATE TABLE shopping_list (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id TEXT NOT NULL,
    item TEXT NOT NULL,
    quantity INTEGER DEFAULT 0,
    FOREIGN KEY (chat_id) REFERENCES chats(chat_id)
);

-- Reminders and scheduled messages
CREATE TABLE reminders (
    id TEXT PRIMARY KEY,
    chat_id TEXT NOT NULL,
    message TEXT NOT NULL,
    time INTEGER DEFAULT 0,  -- Unix timestamp, 0 = untimed
    type TEXT NOT NULL,       -- 'remind', 'send', 'timeless'
    snoozable INTEGER DEFAULT 0,
    FOREIGN KEY (chat_id) REFERENCES chats(chat_id)
);
```

## Migration from JavaScript to Go

### Data Migration
If you have existing `saved-events.json` data:

1. Parse the JSON file
2. Extract relevant data structures:
   - `allEvents[chat].location` → `chats.location`
   - `allEvents[chat].quickShab` → `quickshab` + `quickshab_assignments`
   - `allEvents[chat].timedList` → `reminders` (with time)
   - `allEvents[chat].untimedList` → `reminders` (time=0)
   - `allEvents[chat].shoppingList` → `shopping_list`

3. Insert into SQLite3 database

### Example Migration Script (Pseudocode)
```go
type SavedEvents struct {
    AllEvents map[string]ChatEvents `json:"allEvents"`
}

type ChatEvents struct {
    Location     string                   `json:"location"`
    QuickShab    []QuickShabData         `json:"quickShab"`
    TimedList    map[string]TimedMessage `json:"timedList"`
    UntimedList  map[string]TimedMessage `json:"untimedList"`
    ShoppingList []ShoppingItem          `json:"shoppingList"`
}

func migrate(jsonPath string, db *sql.DB) error {
    // 1. Read and parse JSON
    data, _ := ioutil.ReadFile(jsonPath)
    var saved SavedEvents
    json.Unmarshal(data, &saved)
    
    // 2. Migrate each chat
    for chatID, events := range saved.AllEvents {
        // Insert chat
        db.Exec("INSERT INTO chats VALUES (?, ?)", chatID, events.Location)
        
        // Migrate quickshab
        if len(events.QuickShab) > 0 {
            // ... insert quickshab data
        }
        
        // Migrate reminders
        // ... insert timedList and untimedList
        
        // Migrate shopping list
        // ... insert shoppingList
    }
    
    return nil
}
```

## Testing Checklist

### Unit Tests Needed
- [ ] `models.go`: Database initialization and helper functions
- [ ] Each command file: Input parsing and response generation
- [ ] Date/time parsing functions in `remind.go`
- [ ] Category calculation in `quickshab.go` (`needsNumber`)

### Integration Tests Needed
- [ ] Database operations: CRUD operations for each table
- [ ] Command sequences: Create quickshab, add assignments, show results
- [ ] Reminder lifecycle: Create, snooze, mark done
- [ ] Shopping list: Add, update quantity, remove by name/index

### End-to-End Tests Needed
- [ ] WhatsApp message → command parsing → database → response
- [ ] Multiple concurrent users
- [ ] Multiple chats with separate data
- [ ] Edge cases: empty lists, invalid inputs, special characters

## Performance Considerations

### Current Implementation
- Single SQLite3 database file
- Synchronous database operations
- No caching layer

### Potential Optimizations
1. **Connection Pooling**: Use `sql.DB` connection pool properly
2. **Prepared Statements**: Cache frequently used queries
3. **Indexes**: Add indexes on frequently queried columns:
   ```sql
   CREATE INDEX idx_quickshab_chat ON quickshab_assignments(chat_id);
   CREATE INDEX idx_reminders_chat ON reminders(chat_id);
   CREATE INDEX idx_reminders_time ON reminders(time);
   ```
4. **Batch Operations**: Group multiple inserts into transactions

### Scalability Notes
- SQLite3 suitable for single-bot deployment
- For multi-bot deployment, consider PostgreSQL or MySQL
- Current schema designed to work with any SQL database with minor modifications

## Summary

### Fully Implemented (✅)
- Help system (help, docs)
- Location management (shablocation)
- QuickShab meal planning (quickshab, update, show, bring, assign, unbring, unassign)
- Shopping lists (shop, unshop, shoplist)
- Reminders (remind, reminders, done)
- Scheduled messages (send, unsend, scheduled)
- Easter eggs (gil, count)

### Needs Enhancement (⚠️)
- Date/time parsing (remind, send, snooze) - works but limited
- Shabbat times (shabtimes) - needs Hebrew calendar
- Fast days (fast) - needs Hebrew calendar
- Tova commands (start, end) - needs original implementation review

### Migration Path
1. ✅ Core functionality: Already complete
2. 🔄 Enhanced date parsing: Add library like `when` or `dateparse`
3. 🔄 Hebrew calendar: Integrate hebcal.com API
4. 🔄 Tova commands: Review and port from TimedStuff.js
5. 🔄 Data migration: Create tool to import saved-events.json
6. 🔄 Testing: Add comprehensive test suite
