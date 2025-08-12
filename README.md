# The Shab-BOT

A Whatsapp bot for planning shabbat/potluck meals with your friends, with a few extra utilities.



## Features

- Group chat based Shabbat Meal Planning
- Weekly shabbat times from @hebcal
- Scheduled messages and reminders

## Usage

Add this Whatsapp number to your group chat to get started [+1 (203) 802-5238](https://wa.me/message/RCFUO6SIZTRDO1).

In order to use the bot, you must write messages to a Whatsapp chat that the Shab-BOT is in. You can also text the Shab-BOT directly.

Any message that begins with `!` will invoke the Shab-BOT.


## Getting started with Quick Shabbat Meals

Text `!quickshab <Number of guests>` to generate a list of the needs for a meal of that size.

People can sign up to bring things by writing `!bring <category>`. If the category doesn't exist, it will be created.

If you want to remove yourself, you can write `!unbring`. 

You can also assign other people to bring things by writing `!assign <person> <category>`. This can also be undone with `!unassign`

Example Usage:
```
User:
    !quickshab 4

Shab-BOT:
    Main: 
    Main: 
    Side: 
    Wine: 
    Challah: 
    Dips: 

    You can sign yourself up by writing !bring [category]

User:
    !bring main

Shab-BOT:
    Main: User
    Main: 
    Side: 
    Wine: 
    Challah: 
    Dips:

User:
    !unbring

Shab-BOT:
    Main: 
    Main: 
    Side: 
    Wine: 
    Challah: 
    Dips:

User:
    !assign friend pasta

Shab-BOT:
    Main: 
    Main: 
    Side: 
    Wine: 
    Challah: 
    Dips: 
    Pasta: friend

User:
    !unassign friend

Shab-BOT:
    Main: 
    Main: 
    Side: 
    Wine: 
    Challah: 
    Dips: 
```


## Getting started with Shabbat Times

Text `!shabtimes <optional city name>` to find out when shabbat starts this week. If you do not specify a city, it uses the shabbat location associated with that chat. 

Text `!shablocation <city name>` to change the location of a chat. The default location is Haifa.



## Getting started with Events

Use `!new <event name> <time of event>` to make your first event.

The Shab-BOT uses chronos to parse natural language dates, so feel free to write the date however you'd like.

You can list all existing events by writing `!list`.

Once your event has been created, you are automatically added to it as an attendee.
To join someone else's event, use `!join <event name OR index>`.

If you do not specify which event is bring joined, it defaults to the most recently created event. This is true for all commands where an event must be referenced.

If you are no longer coming to an event, use `!leave <event name OR index>`

This is the basic usage of the events of the bot. For more detailed information about commands, read the docs below.


## Docs

Aliases of the command will be written in parenthases after the command, seperated by commas.

Multiple commands may be included in the same message, each seperated by a new line 


### Event Creating

`!new (n) <event name> <time of event>` 
- "Event name" can be anything
- Time of event must include a date, and may optionally include a time (parsed by [chronos](https://github.com/wanasit/chrono))

`!remove (rm) <event name OR index>`
- Removes event
- You must be the first person in the attendee list to remove (check this using `!list <event name OR index>`)
- DOES NOT DEFAULT TO MOST RECENTLY CREATED EVENT

`!rename (rn) <optional event name OR index> <new name>`
- Fairly self explanatory
- If no event is referenced, the most recently created event is renamed

`!settime <optional event name OR index> <new time>`
- Uses [chronos](https://github.com/wanasit/chrono)
- If no event is referenced, defaults to most recently created
- New time can only change the time, not the date.

### Attendee Commands

`!list (ls) <optional event name OR index>`
- If an event is referenced (with name or index), it lists the participants of that events, as well as what they are bringing
- Otherwise, it lists all events and how many people are coming to each.

`!join (jn,coming,cm) <optional event name OR index> <optional *>`
- Adds yourself to the list of attendees of an event.
- If no event is reference, you are added to the most recently created event.
- You can optionally include a number of asterisks *, one for every guest you plan on bringing. This will reflect in the number of people coming when you invoke `!list`

`!leave (lv, uncoming, uncm) <optional event name OR index>`
- Remove yourself from the list of attendees of an event.
- If no event is referenced, the most recently created event is referenced.

`!needs <optional event name OR index> <optional *{number}>`
- Determines what type of dishes still need to be brought.
- If no event is referenced, defaults to the most recently created.
- You can use this command to determine the dishes needed to be brought for a hypothetical amount of people by adding `*<number of people>`. For example, to determine the dishes that need to be brought for a meal of 5 people, use `!needs *5`


### Shabbat Time Commands

`!shabtimes (shabbattimes) <optional city name>`
- If no city is referenced, defaults to the city associated with that chat.

`!shablocation (shabloc, shabbatlocation) <city name>`
- Sets the city associated with your chat.
- Uses the database from [@hebcal](https://github.com/hebcal/hebcal-es6)


### Other Misc Commands

`!remind <what to remind> <when to remind>`
- Sends you message with the reminder at the time requested.
- Date parsed by chrono

`!reminders`
- Lists all future reminders

`!send <message to send> <recipient's phone number> <when to send>`
- Sends a message with the content written at the time requested

`!unsend`
- Will no longer send the most recently queued message

`!scheduled`
- Lists all messages that will be sent from you

`!help`
- Responds with a short version of the docs



