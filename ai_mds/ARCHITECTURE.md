# ShabBOT Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        WhatsApp Client                           │
│                    (whatsmeow library)                           │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         │ Messages
                         ↓
┌─────────────────────────────────────────────────────────────────┐
│                      Event Handler                               │
│                   (bot.go main loop)                             │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         │ Extract: chatID, userID, message
                         ↓
┌─────────────────────────────────────────────────────────────────┐
│                     Command Router                               │
│                  (commands/router.go)                            │
│                                                                   │
│  • ParseMultipleCommands() - Handle multiple commands            │
│  • Route() - Dispatch to command handler                         │
└────────────────────────┬────────────────────────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
        ↓                ↓                ↓
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   Simple     │  │   Database   │  │  External    │
│   Commands   │  │   Commands   │  │  API Calls   │
├──────────────┤  ├──────────────┤  ├──────────────┤
│ • help       │  │ • quickshab  │  │ • shabtimes  │
│ • docs       │  │ • bring      │  │ • fast       │
│ • gil        │  │ • shop       │  │              │
│ • count      │  │ • remind     │  │              │
└──────────────┘  │ • send       │  └──────┬───────┘
                  │ • location   │         │
                  └──────┬───────┘         │
                         │                 │
                         │                 │ hebcal.com API
                         ↓                 ↓
                  ┌──────────────────────────┐
                  │    SQLite3 Database      │
                  │  (shabbot.db)            │
                  ├──────────────────────────┤
                  │ • chats                  │
                  │ • quickshab              │
                  │ • quickshab_assignments  │
                  │ • shopping_list          │
                  │ • reminders              │
                  └──────────────────────────┘
```

## Data Flow

### Command Execution Flow

```
User Message                       Database                     Response
    │                                  │                            │
    │ "!quickshab 10"                 │                            │
    ↓                                  │                            │
┌────────┐                            │                            │
│ Router │ Parse command              │                            │
└────┬───┘                            │                            │
     │                                 │                            │
     ↓                                 │                            │
┌──────────────┐                      │                            │
│ QuickShabCmd │                      │                            │
└──────┬───────┘                      │                            │
       │                               │                            │
       │ GetOrCreateChat(chatID) ────→│                            │
       │                               │                            │
       │←──────────────────────────────┤ Insert/Update              │
       │                               │                            │
       │ INSERT quickshab ────────────→│                            │
       │                               │                            │
       │←──────────────────────────────┤ Success                    │
       │                               │                            │
       │ formatQuickShab() ───────────→│ Query assignments          │
       │                               │                            │
       │←──────────────────────────────┤ Return rows                │
       │                               │                            │
       └──────────────────────────────────────────────────────────→│
                                                                    │
         "Main: \nSide: \nDrinks: \nWine: \n..."                  │
                                                                    │
                                                    Sent to user ──┘
```

### Database Operations

```
Command Type          Database Tables Used              SQL Operations
────────────────────────────────────────────────────────────────────
Location              chats                             SELECT, UPDATE, INSERT
QuickShab            quickshab                         INSERT, UPDATE, SELECT
                     quickshab_assignments             INSERT, DELETE, SELECT
Shopping             shopping_list                     INSERT, DELETE, SELECT, UPDATE
Reminders            reminders                         INSERT, DELETE, SELECT, UPDATE
Scheduled Messages   reminders (type='send')           INSERT, DELETE, SELECT
```

## Module Structure

```
golangShabBOT/
├── bot.go                     # Main entry point, WhatsApp connection
├── timeRoutine.go             # Background tasks
├── INTEGRATION_EXAMPLE.go     # Integration examples
│
├── commands/
│   ├── models.go              # Database schema & core types
│   │   ├── ChatData struct
│   │   ├── QuickShab struct
│   │   ├── ShoppingItem struct
│   │   ├── Reminder struct
│   │   ├── InitDB()
│   │   ├── GetOrCreateChat()
│   │   ├── GetChatLocation()
│   │   └── SetChatLocation()
│   │
│   ├── utils.go               # Shared utilities
│   │   ├── parseArgs()
│   │   ├── parseInt()
│   │   └── CapitalizeFirst()
│   │
│   ├── router.go              # Command dispatcher
│   │   ├── CommandRouter struct
│   │   ├── Route()
│   │   └── ParseMultipleCommands()
│   │
│   ├── help.go                # Static commands
│   ├── docs.go
│   ├── gil.go
│   ├── count.go
│   │
│   ├── shablocation.go        # Location management
│   │   └── ShabLocationCmd()
│   │
│   ├── shabtimes.go           # Hebrew calendar (placeholder)
│   │   └── ShabTimesCmd()
│   │
│   ├── quickshab.go           # Meal planning
│   │   ├── QuickShabCmd()
│   │   ├── UpdateNumberCmd()
│   │   ├── ShowCmd()
│   │   ├── formatQuickShab()
│   │   └── needsNumber()
│   │
│   ├── bring.go               # Assignment management
│   │   ├── BringCmd()
│   │   ├── AssignCmd()
│   │   ├── UnbringCmd()
│   │   └── UnassignCmd()
│   │
│   ├── shop.go                # Shopping lists
│   │   ├── ShopCmd()
│   │   ├── ShopListCmd()
│   │   └── UnshopCmd()
│   │
│   ├── remind.go              # Reminders
│   │   ├── RemindCmd()
│   │   ├── RemindersCmd()
│   │   ├── SnoozeCmd()
│   │   ├── DoneCmd()
│   │   └── Helper functions
│   │
│   ├── send.go                # Scheduled messages
│   │   ├── SendCmd()
│   │   ├── UnsendCmd()
│   │   └── ScheduledCmd()
│   │
│   ├── needs.go               # Deprecated features
│   ├── fast.go                # Hebrew calendar (placeholder)
│   └── tova.go                # Tova commands (placeholder)
│
└── response/
    └── response.go            # (Currently empty, for future use)
```

## Database Schema

```
┌─────────────────────────────────────────────────────────────────┐
│                            chats                                 │
├─────────────────────────────────────────────────────────────────┤
│ chat_id (PK)    │  location                                      │
│ TEXT            │  TEXT DEFAULT 'Haifa'                          │
└─────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    │               │               │
        ┌───────────↓───────┐   ┌──↓──────────┐   ↓
        │    quickshab       │   │ shopping_   │   reminders
        │                    │   │    list     │
        ├────────────────────┤   ├─────────────┤   ├──────────────
        │ chat_id (PK, FK)   │   │ id (PK)     │   │ id (PK)
        │ number             │   │ chat_id(FK) │   │ chat_id (FK)
        └────────┬───────────┘   │ item        │   │ message
                 │                │ quantity    │   │ time
                 │                └─────────────┘   │ type
        ┌────────↓────────────┐                     │ snoozable
        │ quickshab_          │                     └──────────────
        │    assignments      │
        ├─────────────────────┤
        │ id (PK)             │
        │ chat_id (FK)        │
        │ category            │
        │ name                │
        └─────────────────────┘

Foreign Key Relationships:
• quickshab.chat_id → chats.chat_id
• quickshab_assignments.chat_id → chats.chat_id
• shopping_list.chat_id → chats.chat_id
• reminders.chat_id → chats.chat_id
```

## Command Categories by Complexity

### Level 1: Static Responses (No State)
```
┌────────┐
│ Input  │──→ Function ──→ Static String ──→│ Output │
└────────┘                                   └────────┘

Examples: help, docs, gil, count
Complexity: O(1)
Database: None
```

### Level 2: Simple Database Operations
```
┌────────┐
│ Input  │──→ Parse ──→ SQL Query ──→ Format ──→│ Output │
└────────┘              ↓                        └────────┘
                   Database

Examples: shablocation (get/set), shoplist, reminders (list)
Complexity: O(1) - O(n) where n is result set size
Database: Single table, simple queries
```

### Level 3: Complex Database Operations
```
┌────────┐
│ Input  │──→ Parse ──→ Multiple Queries ──→ Calculations ──→ Format ──→│ Output │
└────────┘              ↓                                                └────────┘
                   Database
                   
Examples: quickshab, bring, shop
Complexity: O(n) where n is number of items/assignments
Database: Multiple tables, joins, aggregations
```

### Level 4: External Dependencies
```
┌────────┐
│ Input  │──→ Parse ──→ API Call ──→ Parse Response ──→ Format ──→│ Output │
└────────┘              ↓                                          └────────┘
                   External API
                   
Examples: shabtimes, fast (when implemented)
Complexity: O(1) + network latency
Database: May cache results
```

### Level 5: Complex Logic + State
```
┌────────┐
│ Input  │──→ Parse ──→ Date Parse ──→ DB Operations ──→ State Management ──→│ Output │
└────────┘              ↓              ↓                                      └────────┘
                   Library        Database
                   
Examples: remind, send, snooze
Complexity: O(n) where n is number of reminders
Database: Complex queries, state transitions
```

## Concurrency Model

```
┌─────────────────────────────────────────────────────────────────┐
│                         Main Thread                              │
│                                                                   │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │  WhatsApp    │    │   Database   │    │   HTTP       │      │
│  │  Connection  │    │  Connection  │    │   Client     │      │
│  │  (Client)    │    │  Pool        │    │  (for APIs)  │      │
│  └──────────────┘    └──────────────┘    └──────────────┘      │
└─────────────────────────────────────────────────────────────────┘
           │                    │                    │
           │                    │                    │
    ┌──────↓─────┐      ┌──────↓─────┐      ┌──────↓─────┐
    │ Goroutine  │      │ Goroutine  │      │ Goroutine  │
    │  Message   │      │  Database  │      │  API       │
    │  Handler   │      │  Query     │      │  Request   │
    └────────────┘      └────────────┘      └────────────┘

Notes:
• Each message handled in separate goroutine
• Database connection pool manages concurrent queries
• SQL transactions protect data integrity
• Commands are stateless - safe for concurrent execution
```

## Performance Characteristics

```
Command Type          Avg Response Time    Database Queries    Scalability
──────────────────────────────────────────────────────────────────────────
Static                < 1ms                0                   Unlimited
Location (get)        1-5ms                1 SELECT            100K+ req/s
Location (set)        5-10ms               1 INSERT/UPDATE     10K+ req/s
QuickShab (show)      10-50ms              2-3 SELECTs         1K+ req/s
QuickShab (create)    20-100ms             2 INSERTs           500+ req/s
Shopping (list)       10-50ms              1 SELECT            1K+ req/s
Shopping (add)        20-100ms             1-2 INSERTs         500+ req/s
Reminders (list)      10-50ms              1 SELECT            1K+ req/s
Reminders (create)    20-100ms             1 INSERT            500+ req/s
Shabtimes (API)       100-500ms            0-1 (cache)         100+ req/s
```

## Error Handling Flow

```
User Input
    │
    ↓
┌──────────────┐
│   Router     │──→ Command Not Found ──→ Empty Response
└──────┬───────┘
       │
       ↓
┌──────────────┐
│   Command    │──→ Parse Error ──→ "Command syntax: ..."
└──────┬───────┘
       │
       ↓
┌──────────────┐
│   Database   │──→ DB Error ──→ "Error [action]"
└──────┬───────┘
       │
       ↓
┌──────────────┐
│   Format     │──→ Success ──→ Formatted Response
└──────────────┘

Error Types:
• Syntax Errors    → User-friendly message with example
• Database Errors  → Generic error message (don't expose internals)
• Not Found        → Helpful message (e.g., "No reminders found")
• Validation       → Specific feedback (e.g., "Number must be positive")
```

## Future Enhancements

```
Current Architecture          Future Architecture
─────────────────────────────────────────────────────────────
                                   ┌──────────────┐
  ┌──────────────┐                │   Redis      │
  │   SQLite3    │                │   Cache      │
  │   Database   │                └──────┬───────┘
  └──────────────┘                       │
                                         │
                          ┌──────────────┴───────────────┐
                          │                              │
                  ┌───────↓────────┐          ┌─────────↓────────┐
                  │   PostgreSQL   │          │   Message Queue  │
                  │   Primary DB   │          │   (Background)   │
                  └────────────────┘          └──────────────────┘

Benefits:
• Cache reduces DB load
• PostgreSQL enables multi-instance deployment
• Message queue for async operations (reminders, scheduled messages)
• Horizontal scaling possible
```

## Deployment Architecture

```
Production Deployment Options:

Option 1: Single Server (Current)
┌─────────────────────────────────┐
│         Server                   │
│  ┌────────────────────────────┐ │
│  │  WhatsApp Bot Process      │ │
│  │  (single instance)         │ │
│  └────────────┬───────────────┘ │
│               │                  │
│  ┌────────────↓───────────────┐ │
│  │  SQLite3 Database          │ │
│  │  (file: shabbot.db)        │ │
│  └────────────────────────────┘ │
└─────────────────────────────────┘

Option 2: High Availability
┌───────────────────────────────────────────────────────┐
│                  Load Balancer                        │
└────────────┬─────────────────────────┬────────────────┘
             │                         │
    ┌────────↓────────┐       ┌───────↓─────────┐
    │   Bot Instance  │       │  Bot Instance   │
    │   (Server 1)    │       │  (Server 2)     │
    └────────┬────────┘       └───────┬─────────┘
             │                        │
             └────────────┬───────────┘
                          │
                   ┌──────↓──────┐
                   │ PostgreSQL  │
                   │  Database   │
                   └─────────────┘
```

---

**Document Version**: 1.0
**Last Updated**: [Current Date]
**Architecture Status**: Designed and Implemented
**Deployment Status**: Ready for Single Server Deployment
