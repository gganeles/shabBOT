const dayjs = require('dayjs');
const chrono = require('chrono-node');

const dateFormat = 'ddd DD.MM.YYYY @ h:mm a'
const {HebrewCalendar, HDate, Location, Event} = require('@hebcal/core')

var events = [{"eventName":"Yom hashoa learning","date":"2023-04-17T23:50:00.000Z","attendance":[{"id":"Gabriel Ganeles","number":"972587120601","guests":0,"food":"nothing"},{"id":"Hodi Sackstein","number":"972539240310","guests":0,"food":"nothing"}]}]


function needs(event) {
    const preset = {
        'main_dish': 2,
        'side_dish': 2,
        'plastics': 1,
        'drinks': 1,
        'wine': 1,
        'challah': 1,
        'dips': 1
    }
    event.attendance.forEach((person) => {
        if (person.foodtype != undefined) {
            person.foodtype.forEach((type) => {
                Object.entries(preset).forEach(([key, value]) => {
                    if (value > 0) {
                        if (type == key) {
                            preset[key] -= 1;
                        }
                    }
                });
            });
        }
    });
    return Object.entries(preset).filter(([key, value]) => value > 0).map(([key, value]) => `${key}: ${value}`).join('\n');
}
async function GPTparse(item) {
    const apiUrl = 'https://api.openai.com/v1/chat/completions'
    //const dishList = 'main_dish, side_dish, disposables, drinks, wine, challah, dips'
    const gptPrompt = `I input item name you output type: main_dish or side_dish. ONLY respond with type. The first prompt is ${item}`
    const key = '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'
    const request = {
        model:'gpt-3.5-turbo',
        messages:[
            {role:'user',content: gptPrompt}
        ],
        temperature: 0,
        max_tokens: 5,
    }
    const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${key}`,
    }

    const response = await fetch(apiUrl, {
        method: 'POST',
        headers: headers,
        body: JSON.stringify(request)
    });
    //.then(response => response.json()).then(item=> console.log(item.choices.at(0).message.content))

    const result = await response.json();
    return await result.choices.at(0).message.content
}

function shabTimes(date) {
    const now = date
    const options = {
      start: now,
      end: now,
      candlelighting: true,
      location: Location.lookup('Haifa'),
    };

    const events = HebrewCalendar.calendar(options);

    const candleLighting = events.filter(item => item.getDesc() == 'Candle lighting').at(0).eventTimeStr
    return {location: 'Haifa', time: candleLighting}
}


function countAst(str) {
    let count = 0
    for (let i = 0; i < str.length; i++) {
        if (str.at(i) === '*') {
            count += 1;
        }
    }
    return count;
}

function find_event_index_by_name(string, eventList) {
    for (let i = 0; i < eventList.length; i++) {
        if (string.match(new RegExp('\\b' + eventList.at(i).eventName + '\\b', 'i'))) {
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
        let indicies = [];
        string.split(/\s*,\s*|\s/).filter(item => !isNaN(parseInt(item))).forEach(item => indicies.push(parseInt(item)-1))
        return indicies
    }
}

function removeCmdandIndicies(string, eventList) {
    for (let i = 0; i < eventList.length; i++) {
        string = string.replace(new RegExp('\\s' + eventList.at(i).eventName + '\\b', 'i'),'')
    }
    return string.split(/\s*,\s*|\s/).slice(1,).filter(item => isNaN(parseInt(item))).join(' ').trim()

}

function sumGuests(event) {
    let summ = 0;
    for (let i=0; i<event.attendance.length; i++) {
        summ += event.attendance.at(i).guests;
    }
    return summ;
}

function upcomingEvents(eventList) {
    const now = new Date()
    return eventList.filter((event) => event.date > now)
}

function changeTime(event, date) {
    event.date = date
    return `${event.eventName} will now be on ${dayjs(date).format(dateFormat)}`
}

function listEvents(events,requestedindex) {
    var messageContent = ``;
    if (requestedindex===-1) {
        messageContent+=`There are ${events.length} events planned.\n`
        events.forEach((event, index) => messageContent += `${index + 1}) ${event.eventName} on ${dayjs(event.date).format(dateFormat)}, ${event.attendance.length+sumGuests(event)} coming\n`);
        messageContent+=`Use !list <event name or indicies> for details`
    } else {
        let location = ''
        const event = events.at(requestedindex)
        if (event.location != undefined) {
            location = ` at ${event.location}`
        }
        messageContent += `${event.eventName} on ${dayjs(event.date).format(dateFormat)}${location}\n`
        for (let j = 0; j < event.attendance.length; j++) {
            let num_of_guests = ''
            let food = ''
            if (event.attendance.at(j).guests > 0) {
                num_of_guests = ` +${event.attendance.at(j).guests.toString()}`;
            }
            if (event.attendance.at(j).food != 'nothing') {
                food = ` is bringing ${event.attendance.at(j).food}`
            }
            messageContent += `  ${event.attendance.at(j).id}${num_of_guests}${food}\n`;
        }
    }
    return messageContent
}

function newestIndex(eventList) {
    for (let i = 0; i<eventList.length; i++) {
        if (eventList.at(i).isNewest) {
            return i
        }
    }
}

function newEvent(name, date, creator, eventList) {
    eventList.forEach(event => event.isNewest = false)
    eventList.push({ eventName: name, date: date.date(), attendance: [creator], isNewest: true});
    eventList.sort(function(a, b) {
    return a.date - b.date;
    });
    return `There's a new "${name}" on ${dayjs(date.date()).format(dateFormat)}.\nTo rename use "!rename <old name> <new name>"\nTo set location use "!location <event name> <where>"\n`
}


function updateGuests(attendee, event, astNum) {
    for (let p = 0; p < event.attendance.length; p++) {       
        if (event.attendance.at(p).number === attendee.number & event.attendance.at(p).guests != astNum) {
            event.attendance.at(p).guests = astNum;
            return `number of guests updated`
        }
    }
    return "You're already signed up silly\n"
}

function isComing(phoneNumber, event) {
    for (let p = 0; p < event.attendance.length; p++) {
        if (event.attendance.at(p).number === phoneNumber) {
            return true;
        }
    }
    return false
}

function updateAttendee(attendee, astNum, event) {
    if (!isComing(attendee.number, event)) {
        event.attendance.push(attendee)
        return `${attendee.id} will be attending ${event.eventName}`
    }
    return updateGuests(attendee, event, astNum)
}

function coming(prompt, eventList, attendee) {
    const indexList = findIndicies(prompt, eventList)
    let messageContent = []
    const astNum = countAst(prompt)
//    if (prompt.split(' ').length===1) {
//        messageContent.push(updateAttendee(attendee, astNum, eventList.at(newestIndex(eventList))))
//    } else 
    if (indexList.length===0) {
        return ["Usage: coming <event name or indicies> <asterisk per extra guest>"]
    }
    for (let i = 0; i < indexList.length;i++) {
        if (events.length>indexList.at(i)) {
            messageContent.push(updateAttendee(attendee, astNum, eventList.at(indexList.at(i))))
        } else {
            messageContent.push('that event does not exist')
        }
    }
    return messageContent
}

function removeAttendee(attendee,event) {
    for (let p = 0; p < event.attendance.length; p++) {
        if (event.attendance.at(p).number === attendee.number) {
            event.attendance.splice(p,1);
            return `${attendee.id} will no longer be attending ${event.eventName}`
        }
    }
    return "You're not even signed up for that one you cuck"
}

function leave(prompt,eventList,attendee) {
    const indexList = findIndicies(prompt, eventList)
    let messageContent = []
    if (indexList.length===0) {
        return ["Usage: leave <event name or indicies>"]
    } else {
        for (let i = 0; i < indexList.length;i++) {
            if (eventList.length>indexList.at(i)) {
                messageContent.push(removeAttendee(attendee,eventList.at(indexList.at(i))))
            } else {
                messageContent.push('that event does not exist')
            }
        }
    }
    return messageContent
}

function list(prompt, eventList) {
    const indexList = findIndicies(prompt, eventList)
    let messageContent = []
    if (indexList.length===0) {
        messageContent.push(listEvents(eventList,-1));
    }
    for (let i = 0; i < indexList.length;i++) {
        if (eventList.length>indexList.at(i)) {
            messageContent.push(listEvents(eventList,indexList.at(i)));
        } else {
            messageContent.push('that event does not exist')
        }
    }
    return messageContent;
}

function createEvent(prompt, eventList, creator) {
    const now = new Date();
    prompt = prompt.replace(new RegExp('\\s(on|at|in)','gi'),'')
    const date = chrono.parse(prompt, now, { forwardDate: true }).at(0);
    if (date === undefined|| ''===prompt.substr(prompt.split(/\s*,\s*|\s/).at(0).length+1, date.index - (prompt.split(/\s*,\s*|\s/).at(0).length+2))) {
        return `Usage: "!new <event name> <when>"`
    } else {
        const name = prompt.substr(prompt.split(/\s*,\s*|\s/).at(0).length+1, date.index - (prompt.split(/\s*,\s*|\s/).at(0).length+2));
        return newEvent(name,date,creator,events)   
    }
}

function newName(prompt, event) {
    return prompt.replace(new RegExp(`\\b${event.eventName}(\\s|)|rename\\s|rn\\s|\\s\\d|\\d`,'gi'),'').trim()
}

function rename(prompt, eventList) {
    const indexList = findIndicies(prompt, eventList)
    if (indexList.length===0) {
        return "Usage: !rename <event name or index> <new-name>"
    }
    for (let i = 0; i < indexList.length;i++) {
        if (eventList.length>indexList.at(i)) {
            eventList.at(indexList.at(i)).eventName = newName(prompt,eventList.at(indexList.at(i)))
            return `It was renamed to ${eventList.at(indexList.at(i)).eventName}`
        } else {
            return 'that event does not exist'
        }
    }
}

async function itemTyper(item) {
    const foodtypes = item.match(/(drinks|wine|challah|salatim|plastics)/gi)
    console.log(foodtypes)
    if (foodtypes) {
        return foodtypes;
    } else {
        const type = await GPTparse(item)
        if (type.match(/(main_dish|side_dish)/gi)) {
            return [type]
        } else {
            return ["other"]
        }
    }
}
async function updateAttendeeFood(attendee,event) {
    for (let i = 0;i<event.attendance.length;i++){
        if (attendee.number===event.attendance.at(i).number) {
            event.attendance.at(i).food=attendee.food
            event.attendance.at(i).foodtype = await itemTyper(attendee.food)
            console.log(event.attendance.at(i).foodtype)
            return `${attendee.id} is bringing ${attendee.food} to ${event.eventName}\nthe foodtype is ${event.attendance.at(i).foodtype}\n`
        }
    }
    attendee.foodtype = await itemTyper(attendee.food)
    event.attendance.push(attendee)
    return `${attendee.id} will be attending ${event.eventName} and bringing ${attendee.food} ${attendee.foodtype}\n`
}

async function bringing(prompt,eventList,attendee) {
    const indexList = findIndicies(prompt, eventList)
    attendee.food = removeCmdandIndicies(prompt, eventList).split(' ').filter(item => (item != 'to')).join(' ')
    let messageContent = []
    if (eventList.length>0 & indexList.length===0) {
        messageContent.push(await updateAttendeeFood(attendee,eventList.at(newestIndex(eventList))))
    } else if (indexList.length===0||attendee.food==='') {
        return ["Usage: !bringing <food you're bringing> <event name or indicies>"]
    }
    for (let i = 0; i < indexList.length;i++) {
        if (eventList.length>indexList.at(i)) {
            messageContent.push(await updateAttendeeFood(attendee,eventList.at(indexList.at(i))))
        } else {
            messageContent.push('that event does not exist')
        }
    }
    return messageContent
}

function location(prompt,eventList) {
    const indexList = findIndicies(prompt,eventList)
    const location = removeCmdandIndicies(prompt,eventList)
    if (indexList.length===0) {
        return "Usage: !location <event name or index> <where>\n"
    }
    for (let i = 0; i < indexList.length;i++) {
        if (eventList.length>indexList.at(i)) {
            eventList.at(indexList.at(i)).location = location
            return `the location of ${eventList.at(indexList.at(i)).eventName} was set to ${location}`
        } else {
            return 'that event does not exist'
        }
    }
}

function remove(prompt,eventList,attendee,EVENTS_FOR_REMOVAL) {
    const indexList = findIndicies(prompt,eventList)
    let messageContent = []
    if (indexList.length===0) {
        return ["Usage: !remove <event name or index>\n"]
    }
    for (let i = 0; i < indexList.length;i++) {
        if (!(attendee.number === '972587120601' || eventList.at(indexList.at(i)).attendance.at(0).number === attendee.number)) {
            messageContent.push('you have no permission you billiards baby\n')
        } else if (eventList.length>indexList.at(i)) {
            messageContent.push(`Removing event ${eventList.at(indexList.at(i)).eventName}\n`)
            const eventToRemove = EVENTS_FOR_REMOVAL.findIndex((event) => event.eventName === eventList.at(indexList.at(i)).eventName)
            EVENTS_FOR_REMOVAL.splice(eventToRemove,1)
            eventList.splice(indexList.at(i),1)
        } else {
            messageContent.push('that event already does not exist you insolant monkey\n')
        }
    }
    return messageContent
}

function settime(prompt,eventList) {
    const indexList = findIndicies(prompt,eventList)
    const now = new Date()
    const date = chrono.parse(prompt, now, { forwardDate: true }).at(0)
    if (indexList.length===0 || date === undefined) {
        return "Usage: !settime <event name or index> <when>\n"
    }
    for (let i = 0; i < indexList.length;i++) {
        if (eventList.length>indexList.at(i)) {
            eventList.at(indexList.at(i)).date = date.date()
            return `${eventList.at(indexList.at(i)).eventName} will now occur on ${dayjs(date.date()).format(dateFormat)}\n`
        } else {
            return 'that event does not exist\n'
        }
    }
}


function soccerListTimer(chat,client,eventsList) {
    //var date = new Date()
   // if (date.getDay() == 0 && date.getHours() == 1 && date.getMinutes() == 37 && date.getSeconds() < 30) {
       // client.sendMessage(chat,'footie list time')
    newSoccerEvent(eventsList)
 //   }
}

function newSoccerEvent(eventList) {
    date = chrono.parse('next tuesday at 9pm').at(0).date()
    eventList.forEach(event => event.isNewest = false)
    eventList.push({ eventName: 'footie game', date: date, attendance: [], isNewest: true, isSoccer: true})
    eventList.push({ eventName: 'waitlist', date: date, attendance: [], isNewest: false, isSoccer: true})
}

function saveToLog(data) {log+=data}
function sampleResponse(message, events, attendee, EVENTS_FOR_REMOVAL) {
    let log = ''
    for (let q=0; q<message.split('\n').length;q++) {
        prompt = message.split('\n').at(q)
        if (prompt.startsWith('list')||prompt.startsWith('ls')) { // list events|people|dishes
            let messageContent = ''
            list(prompt,events).forEach(item => messageContent += item);
            log += messageContent;

        } else if (prompt.startsWith('new')||prompt.startsWith('n')) { // new event name on date
            log += createEvent(prompt, events, attendee);

        } else if (prompt.startsWith('rename')||prompt.startsWith('rn')) { // rename bad name|index "Good name"
            log += rename(prompt, events);

        } else if (prompt.startsWith('coming')||prompt.startsWith('cm')||prompt.startsWith('join')||prompt.startsWith('jn')) { // coming event name
            let messageContent = ''
            coming(prompt,events,attendee).forEach(item => messageContent += item);
            log += messageContent;

        } else if (prompt.startsWith('bringing')||prompt.startsWith('br')||prompt.startsWith('bring')) { // the food you are bringing
            bringing(prompt,events,attendee).then(thing => {
                let messageContent = ''
                thing.forEach(item => messageContent += item);
                log += messageContent;
            });

        } else if (prompt.startsWith('daniel')) {
                const randInt = Math.floor(Math.random()*5);
                var link = '';
                switch (randInt) {
                    case 0:
                        link = 'https://www.youtube.com/watch?v=vXYVfk7agqU';
                        break
                    case 1:
                        link = 'https://www.youtube.com/watch?v=ETfiUYij5UE';
                        break
                    case 2:
                        link = 'https://www.youtube.com/watch?v=uO8SeXh_LaA';
                        break
                    case 3:
                        link = 'https://www.pornhub.com/view_video.php?viewkey=ph5f3b21528d0d5';
                        break
                    case 4:
                        link = 'https://youtu.be/KhqjlVn9q4Y?t=117'
                }
                msg.reply(`YOU HAVE AWAKENED THE WRATH OF SATAN\n${link}`);
        } else if (prompt.startsWith('hodi')) {
                msg.reply('lets sing some shabbat songs\nhttps://www.youtube.com/watch?v=EWMPVn1kgIQ');
        } else if (prompt.startsWith('leave')||prompt.startsWith('lv')||prompt.startsWith('uncoming')||prompt.startsWith('uncm')) {
            let messageContent = ''
            leave(prompt,events,attendee).forEach(item => messageContent += item);
            log += messageContent;

        } else if (prompt.startsWith('help')) {
            msg.reply('Allow me to introduce myself!\nI am the SHABbot!!\nI can help you organize your SHABbot and holiday meals!\nThe way it works:\nUse !new (n) to create a new event\nUse !coming (jn) to add yourself to events\nUse !leave (lv) to remove yourself\nUse !bring (br) to let us know what you are bringing\nUse !list (ls) to list the events and their details\nUse !daniel for a salty suprise!\nUse !help to bring up this list')

        } else if (prompt.startsWith('location')||prompt.startsWith('loc')) {
            log += location(prompt,events)

        } else if (prompt.startsWith('kira')) {
            msg.reply(`Wait wait Im not ready yet`);
            setTimeout(() => {client.sendMessage(chat,"gimme one more second");}, 3000);
            setTimeout(() => {client.sendMessage(chat,"kk now im ready");}, 5000);
            setTimeout(() => {client.sendMessage(chat,"actually im not sure");}, 6000);
            setTimeout(() => {client.sendMessage(chat,"nvm its fine");}, 9000);
            setTimeout(() => {msg.reply("here's the kira command:")},10000);
            setTimeout(() => {client.sendMessage(chat,"https://flashing-colors.com/")},11000);

        } else if (prompt.startsWith('remove')||prompt.startsWith('rm')){
            let messageContent = ''
            remove(prompt,events,attendee,EVENTS_FOR_REMOVAL).forEach(item => messageContent += item)
            log += messageContent;

        } else if (prompt.startsWith('fraydy')) {
            const randInt = Math.floor(Math.random()*2);
            console.log(randInt)
            let link = 'https://www.youtube.com/watch?v=6Z8jBJ6w_7Q'
            if (randInt) {
                link = 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'
            }
            msg.reply(`I didn't know what to put for this one so here's a video of an elephant doing an handstand\n${link}`)
        } else if (prompt.startsWith('settime')) {
            log += settime(prompt,events)
        } else if (prompt.startsWith('remind')) {
            msg.reply(reminderCMD(prompt,chat,allEvents[chat].timedstuff))
        }
    }
    return log
}


class timedMsg {
    constructor(message, time, chat) {
        this.message = message
        this.time = time
        this.chat = chat
    }

    send() {
        client.sendMessage(this.chat,this.toString())
        //destroy this object
    }

    tick(date) {
        if (date >= this.time) {
            this.send()
            return true
        } else {
            return false
        }
    }

    toString() {
        return `you wanted me to remind you:\n "${this.message}" at ${dayjs(this.date).format(dateFormat)}`
    }

    static create(message,date,chat) {
        return new timedMsg(message, date, chat)
    }
}

chat = 'aa'

let allEvents = {}
allEvents[chat] = {events:[],timedList:{}}

let timedId = 0

function reminderCMD(prompt,chat,timedList) {
    if (prompt.split(' ').length==1) {
        return 'command usage: !remind   thing to remind   when'
    } else {
        prompt = prompt.split(' ').slice(1,).join(' ')
        const now = new Date()
        const dateObj = chrono.parse(prompt, now, { forwardDate: true }).at(0);
        if (dateObj == undefined) {
            return 'Try rephrasing your date, like "Saturday at 14" or "Tomorrow at 6pm"'
        } else {
            const message = prompt.slice(0,dateObj.index).trim().replace(/(^(me\sto\s|me\s))|\sin$/gi,'')
            timedList[chat+timedId.toString()] = new timedMsg(message,dateObj.date(),chat,timedId,'remind')
	        timedId ++
            return 'Ok, I will remind you "'+message+'" on '+dayjs(dateObj.date()).format(dateFormat)
        }
    }
}


function sendEventTrigger(now,client,allEvents) {
    for (var key in allEvents) {
        if (!allEvents[key]['timedList']) {
            continue
        }
        for (var i in allEvents[key].timedList) {
            allEvents[key].timedList[i].tick(now,client,allEvents)
        }
    }
}

const coolDateFormat = 'ddd @ h:mm a'

function whichWeek(date) {
    const now = new Date()
    const weekDiff = Math.floor((date-now)/(1000*60*60*24*7))
    if (weekDiff == 0) {return ' '}
    if (weekDiff == 1) {return ' next '}
    if (weekDiff > 1) {return ' in '+weekDiff.toString()+' weeks '}
}

function reminders(timedList) {
    let messageContent = 'Current reminders:'
    for (let key in timedList) {
        messageContent+="\n  "+timedList[key].message+whichWeek(timedList[key].time)+dayjs(timedList[key].time).format(coolDateFormat)
    }
    return messageContent
}


console.log(reminderCMD('remind me to poop me in 20 days',chat,allEvents[chat].timedList))
console.log(reminderCMD('remind i have a dentist appointment in 29 days',chat,allEvents[chat].timedList))
console.log(reminders(allEvents[chat].timedList))


//{'reminders','timedMsgs',''}

//const attendancelist = events.at(0).attendance
//console.log(findIndicies(string4, events))

const stringComing0 = 'yom hashoa learning **' // maybe only allow endswith a digit
const stringComing1 = 'yom hashoa learning *'
const stringNums = '1 2 3 yams and sweet potatoes'
const stringList = 'list'
const garb = 'jfasdkl oaslkjdf'
const stringNew1 = 'new party friday'
const stringNew2 = 'new party'
const stringNew3 = 'new friday'
const rename0 = 'rename ballsslapp yom hashoa learning'
const bringing0 = 'bringing chiken to 1'
const bringing1 = 'bringing chicken and nuts ballsslapp'

const attendee0 = {"id":"Gabriel Ganeles","number":"972587120601","guests":0,"food":"nothing"}
const attendee1 = {"id":"Gabriel Ganeles","number":"22","guests":0,"food":"nothing"}


module.exports = { timedMsg, reminderCMD}