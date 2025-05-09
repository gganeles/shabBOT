const {createEvent} = require('../eventstuff/EventCreateCmds.js')
const {shabTimes, shabLocation} = require('../jewstuff/ShabbatTimes.js')
const {list, needsCMD} = require('../eventstuff/ListAndNeedsCommands.js')
const {location, rename, remove, settime} = require('../eventstuff/EventEditCmds.js')
const {coming, leave, bringing} = require('../eventstuff/updateAttendeeCommands.js')
const {reminderCMD, reminders, schedule, unscheduleMostRecent, scheduled} = require('../utilitystuff/ScheduledSendandReminders.js')
const chrono = require('chrono-node')
const { rent } = require("../utilitystuff/rent.js")
const { testClient, testMsg } = require('./SoccerStuff.js')


async function response(client,msg,events,attendee,unfilteredEvents,allEvents) {
    const chat = msg.from;
	console.log(chat);
	const chat0 = allEvents[chat]
    const now = new Date()
    const is_global = true
    console.log(chat0);
    const promptList = msg.body.split('\n')
    for (let q = 0; q < promptList.length; q++) {
        if (/*msg.mentionedIds.includes(client.info.wid._serialized) || this is for checking when tagged */ msg.body.split('\n').at(q).startsWith("!")) {
            //client.sendSeen(msg.getChat()); // Only appear to check messages when tagged or prompted
            prompt = promptList.at(q).replace(new RegExp(`\\s*(^!)\\s*`, "m"), ''); // doesnt remove @${client.info.wid.user}
            console.log('\n',prompt)
            prompt = prompt.replace(new RegExp('\\b(?:my)\\b\\s', 'gi'), `${attendee.id}'s `).trim() // Remove '!' or @Shabbot from message (and other preceding spaces)

            if (prompt.startsWith('list')||prompt.startsWith('ls')) { // list events|people|dishes
                let messageContent = ''
                list(prompt,events).forEach(item => messageContent += item);
                msg.reply(
                    messageContent);
            } else if (prompt.match(/^rent/gi)) {
                msg.reply(rent(prompt))
            } else if (prompt.startsWith('new')||prompt.startsWith('n ')) { // new event name on date
                msg.reply(createEvent(prompt, unfilteredEvents, attendee));

            } else if (prompt.startsWith('rename')||prompt.startsWith('rn')) { // rename bad name|index "Good name"
                msg.reply(rename(prompt, events));

            } else if (prompt.startsWith('coming')||prompt.startsWith('cm')||prompt.startsWith('join')||prompt.startsWith('jn')) { // coming event name
                let messageContent = ''
                coming(prompt,events,attendee).forEach(item => messageContent += item);
                msg.reply(messageContent);

            } else if (prompt.startsWith('bringing')||prompt.startsWith('br')||prompt.startsWith('bring')) { // the food you are bringing
                bringing(prompt,events,attendee).then(thing => {
                    let messageContent = ''
                    thing.forEach(item => messageContent += item);
                    msg.reply(messageContent);
                });
            } else if (prompt.startsWith('leave')||prompt.startsWith('lv')||prompt.startsWith('uncoming')||prompt.startsWith('uncm')) {
                let messageContent = ''
                leave(prompt,events,attendee).forEach(item => messageContent += item);
                msg.reply(messageContent);

            } else if (prompt.startsWith('help')) {
                if (prompt.split(' ').length < 2) {
                    msg.reply('Allow me to introduce myself!\nI am the SHABbot!!\nI can help you with you shabbat meals, as well as other things\n\nThe way it works:\n  Use !new, !join, !bring, !leave, !needs, and !list to make events and keeping track of what everyone\'s bringing\n  Use !shabtimes to find out when shabbat starts\n  Use !remind to set reminders for yourself')
                }
                
            } else if (prompt.startsWith('location')||prompt.startsWith('loc')) {
                msg.reply(location(prompt,events))

            } else if (prompt.startsWith('remove')||prompt.startsWith('rm')){
                let messageContent = ''
                remove(prompt,events,attendee,unfilteredEvents).forEach(item => messageContent += item)
                msg.reply(messageContent);

            } else if (prompt.startsWith('settime')) {
                msg.reply(settime(prompt,events))
            } else if (prompt.startsWith('count')) {
                let link = 'https://www.chabad.org/holidays/sefirah/omer-count_cdo/jewish/Count-the-Omer.htm'
                msg.reply(`you forgot that your suppose to count the omer!? ha! no, not tonight\n${link}`)
            } else if (prompt.match(/^(shabbattimes|shabtimes)\b/gi)) {
                const now = new Date()
                var date = chrono.parse(prompt, now, {forwardDate: true}).at(0)
                if (date == undefined) {
                    date = new Date()
                } else {
                    date = date.date()
                }
                msg.reply(shabTimes(date,chat0.location,prompt,attendee))
            } else if (prompt.match(/^(shabbatLocation|shabloc|shablocation)\b/gi)) {
                msg.reply(shabLocation(prompt,chat0))
            } else if (prompt.match(/^needs\b/gi)) {
                msg.reply(needsCMD(prompt,events));
            } else if (prompt.match(/^remind\b/gi)) {
                if (!chat0['timedList']) {
                    chat0['timedList'] = {}
                }
                msg.reply(reminderCMD(prompt,chat,chat0['timedList']))
            } else if (prompt.match(/^reminders\b/gi)) {
                msg.reply(reminders(chat0.timedList))
            } else if (prompt.match(/^send\b/gi)) {
                if (!chat0['timedList']) {
                    chat0['timedList'] = {}
                }
                msg.reply(schedule(prompt,chat,allEvents))
            } else if (prompt.match(/^unsend\b/gi)) {
                msg.reply(unscheduleMostRecent(chat,allEvents))
            } else if (prompt.match(/^scheduled\b/gi)) {
                msg.reply(scheduled(chat,allEvents))
            }
            /* else {
                let name = attendee.id;
                
                var message_list = [', that is an invalid command, try !help to see more options :)', ', nope, try again. Perhaps type !help ;)', ', no cookies for you. Try !help', ', that\'s what she said. Try !help for more options', ', xD. Shabbat Shalom. wrong command, try !help for more options :)', ', Oopises. invalid, perhaps you should try !help  :)', ', Gabe suks, perhaps you should try !help  :)']
                var message = message_list[Math.floor(Math.random() * 7)]
                var greeting_list = ['Hi ', 'Hello ', 'Hey ', 'Whatsup! ', 'Shalom '] 
                var greeting = greeting_list[Math.floor(Math.random() * 5)]
                msg.reply(`${greeting}${name}${message}`)
            } */
        }
    }
}

module.exports = {response}

//tests

let allEventsTest = {testChat: {events:[],teams:[],location:'Haifa',timedList:{}}}
const testAttendee = {id: 'Gabe', number: '972587120601', guests: 0, food: 'nothing'}

function responseTest(text) {
	console.log(allEventsTest);    
    const now = new Date()
    const msg = new testMsg('!'+text)
	msg.from="testChat"
    response(new testClient(),msg,allEventsTest.testChat.events.filter(item=>item.date>now),testAttendee,allEventsTest.testChat.events,allEventsTest)
}

function autoResponseTest() {
    responseTest("new event0 tomorrow")
    responseTest("list")
    responseTest("rename event0 event1")
    responseTest('leave')
    //responseTest("bringing chicken")
    responseTest("needs")
    responseTest("location my house")
    responseTest("new thing in 4 minutes")
    responseTest("coming 1 2 ***")
    responseTest("list 1 2")
    responseTest("list event1")
    responseTest("settime 2 9:30pm")
    responseTest("remove 1")
    responseTest("shabbatlocation New York")
    responseTest("shabbattimes jerusalem")
    responseTest('reminders')
}

