const dayjs = require('dayjs')
const chrono = require('chrono-node')
const fs = require("fs")
const luxon = require('luxon')
const cityTimezones = require('city-timezones')
const { DateTime } = luxon;
const utc = require('dayjs/plugin/utc');
const { randomUUID } = require('crypto');

dayjs.extend(utc);


// ---------- small helpers (no behavior change) ----------

// Write the events object to disk
function record(arr, jsonPath = '/home/pi/shabbot/saved-events.json') {
    fs.writeFile(jsonPath, JSON.stringify(arr), (err) => {
        if (err) {
            console.log(err)
        }
        console.log('the database has been updated')
    });
}

// Normalize/removes date text and common filler like "me to", "in", "at"
function extractUserMessage(source, dateText = '') {
    return source
        .replace(dateText, '')
        .trim()
        .replace(/^(me\s(to\s)?)|\s(in|at)$/gi, '')
        .trim();
}

// Ensure a chat entry exists with timed/untimed lists
function ensureChatLists(allEvents, chat) {
    if (!allEvents[chat]) allEvents[chat] = {};
    if (!allEvents[chat].timedList) allEvents[chat].timedList = {};
    if (!allEvents[chat].untimedList) allEvents[chat].untimedList = {};
    return allEvents[chat];
}

// Filter and sort reminders by time
function getSortedReminders(timedList) {
    return Object.values(timedList ?? {})
        .filter(x => x.type == "remind")
        .sort((a, b) => a.time - b.time);
}

// Format helper used in many user-facing strings
const coolDateFormat = 'ddd @ h:mm a'
function formatWeekAndTime(date) {
    return whichWeek(date) + dayjs(date).format(coolDateFormat)
}

class timedMsg {
    constructor(message, time, chat, id, type, snoozable = false, lastWentOff = null) {
        this.message = message
        this.time = time
        this.chat = chat
	    this.id = id
        this.type = type
        this.snoozable = snoozable
        this.lastWentOff = lastWentOff
    }

    send(client,allEvents) {
        if (this.type == 'remind') {
            this.lastWentOff = new Date()
            client.sendMessage(this.chat,this.toReminderString())
            this.snoozable=true
        } else {
            client.sendMessage(this.chat,this.message)
            delete allEvents[this.chat].timedList[this.id.toString()]
        }
        console.log('sending message '+this.id.toString())
    }

    tick(date,client,allEvents) {
        const timeDiff = date-this.time
        if (!this.snoozable && timeDiff>=0) {
            this.send(client,allEvents)
            record(allEvents)
            return true
        } else if (this.snoozable && timeDiff>=(1000*60*60*24)) {
            delete allEvents[this.chat].timedList[this.id.toString()]
            record(allEvents)
        } else {
            return false
        }
    }

    toReminderString() {
        return `you wanted me to remind you:\n "${this.message}".\n\nTo snooze this reminder, please reply with:\n"!snooze   new time" within 24 hours.`
    }

    snooze(date) {
        this.snoozable = false
        this.time = date
        this.lastWentOff = null;
        return "Alarm snoozed until " + formatWeekAndTime(date)
    }

    done(allEvents) {
        if (this.type=="timeless") {
            delete allEvents[this.chat].untimedList[this.id.toString()]
        } else {
            delete allEvents[this.chat].timedList[this.id.toString()]
        }
    }
}


function timelessReminder(reminderText, chat, untimedList={}){
	 
      const message = reminderText
        .trim()
        .replace(/^(me\s(to\s)?)|\s(in|at)$/gi, '')
        .trim();

      // 9) Schedule it
    //   console.log(parsedDate.toJSDate());
    //   console.log(parsedDate.toJSDate().toString());
      const id = randomUUID();
      untimedList[id] = new timedMsg(
        message, 'untimed', chat, id, 'timeless'
      );
  
      // 10) Confirmation
      return `Ok, "${message}" will be added to your reminders`
  
  	
}

function parseDate(prompt, location="Haifa") {
      // 1) Resolve user zone
      const lookup = cityTimezones.lookupViaCity(location);
      const timezone = lookup.length > 0 ? lookup[0].timezone : 'Asia/Jerusalem';
      
      const _ = false;
      // 2) “Now” in user zone, for both refDate and comparisons
      const now = DateTime.now().setZone(timezone);
      //console.log('now:', now.toString());
  
      // 3) Fake a JS Date so Chrono’s “today” = user’s today  
      const refDate = new Date(
        now.year, now.month - 1, now.day,
        now.hour, now.minute, now.second, now.millisecond
      );
  
      // 4) Parse
      const results = chrono.parse(prompt, refDate, { forwardDate: true });
      if (!results.length) {
		return [_, _, -1]
      }
      const dateObj = results[0];
  
      // 5) Extract Chrono’s components
      const { knownValues, impliedValues } = dateObj.start;
      const Y = knownValues.year   ?? impliedValues.year;
      const M = knownValues.month  ?? impliedValues.month;
      const D = knownValues.day    ?? impliedValues.day;
      const h = knownValues.hour   ?? impliedValues.hour  ?? 0;
      const m = knownValues.minute ?? impliedValues.minute ?? 0;
      const s = knownValues.second ?? impliedValues.second ?? 0;
  
      // 6) Build a Luxon DateTime in the user zone
      let parsedDate = DateTime.fromObject(
        { year: Y, month: M, day: D, hour: h, minute: m, second: s },
        { zone: timezone }
      );
  
      // 7) If they only gave a clock time and it’s already passed, roll into tomorrow
      if (
        parsedDate < now &&
        dateObj.start.isCertain('hour') &&
        !dateObj.start.isCertain('day')
      ) {
        parsedDate = parsedDate.plus({ days: 1 });
      }
  
      if (parsedDate < now) {
        return [_, _ , 2]
      }

      return [parsedDate, dateObj, 1];
}

function reminderCMD(prompt, chat, timedList, location,untimedList={}) {

    try {
      if (prompt.split(' ').length === 1) {
        return 'command usage: !remind   thing to remind   when';
      }
      prompt = prompt.split(' ').slice(1).join(' ');

      const [parsedDate, dateObj, parseStatus] = parseDate(prompt, location);
      
      if (parseStatus === -1) {
        return timelessReminder(prompt,chat,untimedList)
      } else if (parseStatus === 2) {
        return "The date you chose has already passed. Please choose a future date.";
      }

    const message = extractUserMessage(prompt, dateObj.text)
  
      // 9) Schedule it
    //   console.log(parsedDate.toJSDate());
    //   console.log(parsedDate.toJSDate().toString());
      const id = randomUUID();
      timedList[id] = new timedMsg(
        message, parsedDate.toJSDate(), chat, id, 'remind'
      );
  
      // 10) Confirmation
      return `Ok, I will remind you "${message}"`
            + whichWeek(parsedDate.toJSDate())
           + `${parsedDate.toFormat("ccc 'at' HH:mm")}`;
  
    } catch (e) {
      console.error(e);
      return 'Something went wrong. Please try again.';
    }
  }


function find_event_index_by_name(string, eventList) {
    for (let i = 0; i < eventList.length; i++) {
        const grp = new RegExp('\\b' + string, 'i')
        if (eventList[i].match(grp)) {
            return i
        }
    }
    return -1
}

function findIndicies(string, eventList) {
    const eventByName = find_event_index_by_name(string, eventList)
    if (eventByName != -1) {
        return [eventByName]
    } else {
        let indicies = []
	    const matchResult = string.match(/\b\d+\b/g) 
        if (!matchResult) return []
        matchResult.forEach(item => {
            const num = parseInt(item)-1
            if (num >= 0 && num < eventList.length) {
                indicies.push(num)
            }
        })
        return indicies
    }
}

function getMostRecentWentOffEvent(chat, allEvents) {
    const eventsByWentoff = Object.values(allEvents[chat]?.timedList ?? []).filter(x=>x.type=="remind" && x.lastWentOff!==null).sort((a,b)=>b.lastWentOff-a.lastWentOff);
    return eventsByWentoff.length > 0 ? eventsByWentoff[0] : null;
}

function doneCMD(prompt,chat,allEvents) { //add untimed reminders from the untimed list
	const untimedReminders = Object.values(allEvents[chat]?.untimedList ?? [])
    const sorted_reminders = getSortedReminders(allEvents[chat]?.timedList)
    if (sorted_reminders.length===0 && untimedReminders.length===0) return 'No reminders nice job'
    
    const reminderList = ([...untimedReminders,...sorted_reminders]).map(x=>x.message)
    const query = prompt.trim().split(/\s+/).slice(1).join(" ").trim()
    let indicies = []
    if (query) {
        indicies = findIndicies(query, reminderList)
        indicies.sort((a,b)=>a-b)
    }
    if (indicies.length!==0) {
        let returnStr = ''
        indicies.forEach(index => {
            if (index < untimedReminders.length) {
                returnStr += 'Done with "'+untimedReminders[index].message+'".\n'
                untimedReminders[index].done(allEvents);
            } else {
                returnStr += 'Done with "'+sorted_reminders[index - untimedReminders.length].message+'".\n'
                sorted_reminders[index - untimedReminders.length].done(allEvents);
            }
        });
        return returnStr;
    } else {
        const mostRecentEvent = getMostRecentWentOffEvent(chat, allEvents);
        if (mostRecentEvent) {
            mostRecentEvent.done(allEvents)
            return 'Done with "'+mostRecentEvent.message+'".'
        } else if (sorted_reminders.length!==0) {
            sorted_reminders[0].done(allEvents)
            return 'Done with "'+sorted_reminders[0].message+'".'
        } else {
            untimedReminders[0].done(allEvents)
            return 'Done with "'+untimedReminders[0].message+'".'
        }
    }
}

function snoozeCMD(prompt,chat,allEvents,location) {

    if (prompt.split(' ').length==1) {
        return 'command usage: !snooze (optional name or #) when to remind'
    } else {
        const now = new Date()
    prompt = prompt.split(' ').slice(1).join(' ');

    // Parse the date (Luxon DateTime, chrono result, status)
    const [dateObj, parsedDate, parseStatus] = parseDate(prompt, location);

    if (parseStatus === -1) {
        return "Try rephrasing your date, like 'Saturday at 14' or 'Tomorrow at 6pm'";
    } else if (parseStatus === 2) {
        return "The date you chose has already passed. Please choose a future date.";
    } else {
            const untimedReminders = Object.values(allEvents[chat]?.untimedList ?? [])
            const sorted_reminders = getSortedReminders(allEvents[chat]?.timedList)
            if (sorted_reminders.length==0 && untimedReminders.length==0) return 'No reminders nice job'

            const reminderList = ([...untimedReminders,...sorted_reminders]).map(x=>x.message)
            const query = prompt.replace(parsedDate.text,'').trim()
            let indicies = []
            if (query) {
                indicies = findIndicies(query, reminderList)
                indicies.sort((a,b)=>a-b)
            }
            

            try {
                ensureChatLists(allEvents, chat)
                let chatObj = null;
                if (indicies.length!==0) {
                    indicies.forEach(index => {
                        if (index < untimedReminders.length) {
                            chatObj = allEvents[chat]["untimedList"][untimedReminders[index].id]
                            const id = randomUUID();
                            allEvents[chat]["timedList"][id] = new timedMsg(chatObj.message, dateObj.toJSDate(), chat, id, 'remind', true)
                            chatObj = allEvents[chat]["timedList"][id]
                            delete allEvents[chat]["untimedList"][untimedReminders[index].id]
                            chatObj.snooze(dateObj.toJSDate())
                        } else {
                            chatObj = allEvents[chat]["timedList"][sorted_reminders[index - untimedReminders.length].id]
                            chatObj.snooze(dateObj.toJSDate())
                        }
                    });
                } else {
                    const mostRecentEvent = getMostRecentWentOffEvent(chat, allEvents);
                    if (mostRecentEvent) {
                        chatObj = allEvents[chat]["timedList"][mostRecentEvent.id]
                        chatObj.snooze(dateObj.toJSDate())
                    } else if (sorted_reminders.length!==0) {
                        chatObj = allEvents[chat]["timedList"][sorted_reminders[0].id]
                        chatObj.snooze(dateObj.toJSDate())
                    } else if (untimedReminders.length!==0) {
                        chatObj = allEvents[chat]["untimedList"][untimedReminders[0].id]
                        const id = randomUUID();
                        allEvents[chat]["timedList"][id] = new timedMsg(chatObj.message, dateObj.toJSDate(), chat, id, 'remind', true)
                        chatObj = allEvents[chat]["timedList"][id]
                        delete allEvents[chat]["untimedList"][untimedReminders[0].id]
                        chatObj.snooze(dateObj.toJSDate())
                    } else {
                            return 'No reminders nice job'
                    }
                }
            return 'Ok. I will now remind you "'+chatObj.message+'"'+whichWeek(dateObj.toJSDate())+dayjs(dateObj.toJSDate()).format(coolDateFormat)
            }
            catch (err) {
                console.log(err)
            }

        }
    }
}

function whichWeek(date) {
    const now = new Date()
    const weekDiff = Math.floor((date-now)/(1000*60*60*24*7))
    if (weekDiff < 0) {return ' already happened '}
    if (weekDiff == 0) {return ' '}
    if (weekDiff == 1) {return ' next '}
    if (weekDiff > 1) {return ' in '+weekDiff.toString()+' weeks '}
}

function reminders(timedList,untimedList={}) {

    if ((!timedList||Object.keys(timedList).length === 0)&&(!untimedList||Object.keys(untimedList).length === 0)) {
        return 'No reminders nice job'
    }
    if (!timedList) {
        timedList = {}
    }
    if (!untimedList) {
        untimedList = {}
    }
    
    const sorted_reminders = getSortedReminders(timedList)

    let messageContent = 'Todo:\n'
    let i = 1
	for (let key in untimedList){
        messageContent+="\n"+i+". "+untimedList[key].message
        i++
	}
    
	messageContent+="\n\nReminders:\n"
    for (let reminder of sorted_reminders) {
        messageContent+="\n"+i+". "+reminder.message+formatWeekAndTime(reminder.time)
        i++
    }
    return messageContent
}

function scheduled(chat,allEvents) {
    let messageContent = 'Scheduled Messages:'
    for (let idNum in allEvents[chat].timedList) {
        for (let key in allEvents) {
            if (key !== chat && allEvents[key].timedList[idNum]) {
                messageContent+='\n  "'+allEvents[key].timedList[idNum].message+'" to '+key.replace('@s.whatsapp.us','')
            }
        }
    }
    return messageContent
}

function schedule(prompt,chat,allEvents) {
    if (prompt.split(' ').length==1) {
        return 'command usage: !send   thing to send   who to send it to   when'
    } else {
        prompt = prompt.split(' ').slice(1,).join(' ')
        let contact = prompt.match(/\d{11}\d*/)
	if (!contact) {
             return 'put in a valid phone number (just digits no spaces or anything)'
        }
        
        console.log(prompt.substring(contact.index,))
        const now = new Date()
        const dateObj = chrono.parse(prompt, now, { forwardDate: true }).at(0);
        if (dateObj == undefined || dateObj.date() < now) {
            return 'Try rephrasing your date, like "Saturday at 14" or "Tomorrow at 6pm"'
        } else {
            const message = prompt.slice(0,contact.index).trim().replace(new RegExp('(\\sto$)|\\s+'+contact+'|\\sin$','gi'),'')
           let contactChat = contact+'@s.whatsapp.us' 
            ensureChatLists(allEvents, contactChat)
            ensureChatLists(allEvents, chat)
            const id = randomUUID();

            allEvents[contactChat]['timedList'][id] = new timedMsg(message,dateObj.date(),contactChat, id,'schedule',chat.replace('@s.whatsapp.us',''))
            allEvents[chat]['timedList'][id] = new timedMsg('Sent scheduled message to '+contact,dateObj.date(),chat,id,'schedule')
            return 'Ok, I will send "'+message+'" to '+contact+" on"+whichWeek(dateObj.date())+dayjs(dateObj.date()).utc().format(coolDateFormat)
        }
    }
}

function unscheduleMostRecent(chat,allEvents) {
    const msgId = Object.values(allEvents[chat]['timedList']).filter(item=>item.type='schedule').at(-1).id
    for (let key in allEvents) {
        for (let key2 in allEvents[key]['timedList']) {
            if (key2 == msgId) {
                const messageContent = '"'+allEvents[chat]['timedList'][msgId].message+'" will no longer be sent'
                delete allEvents[chat]['timedList'][msgId]
                delete allEvents[key]['timedList'][key2]
                return messageContent
            }
        }
    }
    return 'no scheduled messages'
}


module.exports = {doneCMD, snoozeCMD, reminderCMD, reminders, timedMsg, schedule, unscheduleMostRecent, scheduled}


function testRemind(){
    testUntimedList = {}
    testTimedList = {}
    chat="1"
    allEvents = {}

    console.log(reminderCMD("remind me grapes", chat, testTimedList, "haifa",testUntimedList))
    console.log(reminders(testTimedList,testUntimedList))

    console.log(reminderCMD("remind me lettuce jokers in 10 minutes", chat, testTimedList, "haifa",testUntimedList))
    console.log(reminders(testTimedList,testUntimedList))

    allEvents[chat]={timedList:testTimedList,untimedList:testUntimedList}
    //console.log(doneCMD("!done lett",chat,allEvents))
    console.log(snoozeCMD("!snooze 1 at 11 tomorrow",chat,allEvents,"haifa"))
    console.log(reminders(allEvents[chat]["timedList"],allEvents[chat]["untimedList"]))
    //console.log(reminders(allEvents[chat]["timedList"],allEvents[chat]["untimedList"]))

    console.log(snoozeCMD("!snooze tomorrow",chat,allEvents,"haifa"))
    console.log(reminders(allEvents[chat]["timedList"],allEvents[chat]["untimedList"]))

}

if (require.main === module) testRemind()
