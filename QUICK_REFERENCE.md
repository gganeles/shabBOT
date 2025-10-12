# Quick Reference Card - ShabBOT Commands

## Command Categories

### 📚 Information Commands
| Command | Description | Example |
|---------|-------------|---------|
| `!help` | Show bot introduction | `!help` |
| `!docs` | GitHub repository link | `!docs` |
| `!count` | Omer counting link | `!count` |

### 🕯️ Jewish Calendar Commands
| Command | Description | Example |
|---------|-------------|---------|
| `!shabtimes [date]` | Shabbat candle lighting times | `!shabtimes`, `!shabtimes next friday` |
| `!shablocation [city]` | Set default location | `!shablocation Jerusalem` |
| `!fast` | Check upcoming fast days | `!fast` |

### 🍽️ Meal Planning (QuickShab)
| Command | Description | Example |
|---------|-------------|---------|
| `!quickshab [num]` | Start meal planning | `!quickshab 15` |
| `!update [num]` | Update attendee count | `!update 20` |
| `!bring [category]` | Sign up to bring item | `!bring main` |
| `!assign [name] [cat]` | Assign to someone | `!assign Sarah dessert` |
| `!unbring` | Remove your assignment | `!unbring` |
| `!unassign [name]` | Remove someone's assignment | `!unassign Sarah` |
| `!show` | Display assignments | `!show` |

### 🛒 Shopping List
| Command | Description | Example |
|---------|-------------|---------|
| `!shop [item]` | Add to shopping list | `!shop wine`, `!shop 3 challah` |
| `!shoplist` | View shopping list | `!shoplist` |
| `!unshop [item]` | Remove from list | `!unshop wine`, `!unshop 2` |
| `!shop clear` | Clear entire list | `!shop clear` |

### ⏰ Reminders
| Command | Description | Example |
|---------|-------------|---------|
| `!remind [msg] [time]` | Create reminder | `!remind buy milk tomorrow 3pm` |
| `!reminders` | List all reminders | `!reminders` |
| `!snooze [duration]` | Snooze last reminder | `!snooze 10 minutes` |
| `!done` | Mark reminder complete | `!done` |

### 📨 Scheduled Messages
| Command | Description | Example |
|---------|-------------|---------|
| `!send [msg] [time]` | Schedule a message | `!send Meeting starts at 3pm today` |
| `!scheduled` | List scheduled messages | `!scheduled` |
| `!unsend` | Cancel last scheduled msg | `!unsend` |

## QuickShab Categories

Standard categories (calculated automatically):
- **main** - Main dishes (1 per 4 people)
- **side** - Side dishes (1 per 6 people)
- **plastics** - Plates/utensils (if >5 people)
- **drinks** - Beverages (if >5 people)
- **wine** - Wine/grape juice (1 per 8 people)
- **challah** - Challah bread (1 per 10 people)
- **dips** - Dips/appetizers (1 per 12 people)
- **dessert** - Desserts (1 per 8 people, if >5)

You can also use custom categories!

## Database Tables

### Quick Reference
```
chats                    - Chat settings (location)
quickshab               - Meal planning header
quickshab_assignments   - Who's bringing what
shopping_list           - Shopping items
reminders              - Reminders & scheduled messages
```

## Common Workflows

### Setting Up a Shabbat Meal
```
!shablocation Jerusalem
!shabtimes
!quickshab 20
!bring main
!assign David wine
!show
```

### Managing Shopping
```
!shop 2 wine bottles
!shop challah
!shop hummus
!shoplist
!unshop 1
```

### Creating Reminders
```
!remind buy challah tomorrow 10am
!remind call mom
!reminders
```

## Command Aliases

Many commands have shorter versions:

| Full Command | Aliases |
|--------------|---------|
| `!shablocation` | `!shabloc`, `!location`, `!setloc` |
| `!shabtimes` | `!shabbattimes` |
| `!bring` | `!br`, `!bringing` |
| `!unbring` | `!unbr` |
| `!update` | `!up`, `!num`, `!ppl` |
| `!reminders` | `!rems`, `!todo` |
| `!shop` | `!shp`, `!sh` |
| `!unshop` | `!unshp`, `!unsh` |
| `!shoplist` | `!shplist`, `!shlist`, `!shls` |
| `!fast` | `!fasting` |

## Multiple Commands

You can send multiple commands at once:

```
!help
!quickshab 10
!bring main
```

Or on one line:
```
!help !docs
```

## Tips & Tricks

1. **Check location first**: Set your location before using `!shabtimes`
2. **Quantity in shopping**: Use numbers: `!shop 3 wine bottles`
3. **Custom categories**: QuickShab accepts any category name
4. **Remove by index**: `!unshop 2` removes second item
5. **Untimed reminders**: Just `!remind call mom` (no time)

## Error Messages

Common issues and solutions:

| Error | Solution |
|-------|----------|
| "You need to create a quickshab first" | Run `!quickshab [number]` first |
| "Command syntax: ..." | Check command format |
| "No reminders found" | You have no active reminders |
| "Your shopping list is empty" | Add items with `!shop` |

## Development Quick Start

### Initialize Database
```go
import "yourmodule/golangShabBOT/commands"

db, _ := sql.Open("sqlite3", "shabbot.db?_foreign_keys=on")
commands.InitDB(db)
```

### Route Commands
```go
router := commands.NewCommandRouter(db)
response := router.Route(message, chatID, userID)
```

### Direct Function Calls
```go
commands.QuickShabCmd(db, "!quickshab 10", chatID)
commands.BringCmd(db, "!bring main", chatID, userID)
commands.ShowCmd(db, chatID)
```

## Status Indicators

- ✅ **Fully Implemented** - Ready to use
- ⚠️ **Simplified** - Works but limited features
- ⚠️ **Placeholder** - Needs additional work
- ⚠️ **Deprecated** - Being phased out

## Support

For issues, see:
- `README_AI_notes.md` - Full documentation
- `COMMAND_MAPPING.md` - Implementation details
- `GO_IMPLEMENTATION_SUMMARY.md` - Architecture overview
- `INTEGRATION_EXAMPLE.go` - Code examples

Repository: github.com/gganeles/shabBOT
