const { createEvent } = require('../eventstuff/EventCreateCmds.js')
const { shabTimes, shabLocation } = require('../jewstuff/ShabbatTimes.js')
const { list, needsCMD } = require('../eventstuff/ListAndNeedsCommands.js')
const { location, rename, remove, settime } = require('../eventstuff/EventEditCmds.js')
const { coming, leave, bringing } = require('../eventstuff/updateAttendeeCommands.js')
const { reminderCMD, doneCMD, snoozeCMD, reminders, schedule, unscheduleMostRecent, scheduled } = require('../utilitystuff/ScheduledSendandReminders.js')
const chrono = require('chrono-node')
const { rent } = require("../utilitystuff/rent.js")
const { testClient, testMsg } = require('./SoccerStuff.js')
const { dontBeAPussy } = require('../utilitystuff/Nico.js')
const { fastLister } = require('../jewstuff/Fast_checker.js')
const { show, quickshab, updateNumber, bringCmd, assignCmd, unbringCmd, unassignCmd } = require('../jewstuff/quickShab.js')
const { tovaTriggerStart, tovaTriggerEnd } = require("./TimedStuff.js")


async function response(client, msg, events, attendee, unfilteredEvents, allEvents) {
    const chat = msg.from;
    const chat0 = allEvents[chat]
    const now = new Date()
    const is_global = true
    const promptList = msg.body.split('\n')
    try {
        for (let q = 0; q < promptList.length; q++) {
            if (/*msg.mentionedIds.includes(client.info.wid._serialized) || this is for checking when tagged */ msg.body.split('\n').at(q).startsWith("!")) {
                //client.sendSeen(msg.getChat()); // Only appear to check messages when tagged or prompted
                prompt = promptList.at(q).replace(new RegExp(`\\s*(^!)\\s*`, "m"), ''); // doesnt remove @${client.info.wid.user}
                console.log('\n', prompt)
                prompt = prompt.replace(new RegExp('\\b(?:my)\\b\\s', 'gi'), `${attendee.id}'s `).trim() // Remove '!' or @Shabbot from message (and other preceding spaces)

                if (prompt.startsWith('list') || prompt.startsWith('ls')) { // list events|people|dishes
                    let messageContent = ''
                    list(prompt, events).forEach(item => messageContent += item);
                    msg.reply(
                        messageContent);
                } else if (prompt.match(/^rent/gi)) {
                    msg.reply(rent(prompt))
                } else if (prompt.startsWith('new') || prompt.startsWith('n ')) { // new event name on date
                    msg.reply(createEvent(prompt, unfilteredEvents, attendee));

                } else if (prompt.startsWith('rename') || prompt.startsWith('rn')) { // rename bad name|index "Good name"
                    msg.reply(rename(prompt, events));

                } else if (prompt.startsWith('coming') || prompt.startsWith('cm') || prompt.startsWith('join') || prompt.startsWith('jn')) { // coming event name
                    let messageContent = ''
                    coming(prompt, events, attendee).forEach(item => messageContent += item);
                    msg.reply(messageContent);

                } else if (prompt.match(/^docs/gi)) {
                    msg.reply("github.com/gganeles/shabBOT")
                    // } else if (prompt.startsWith('bringing')||prompt.startsWith('br')||prompt.startsWith('bring')) { // the food you are bringing
                    //     bringing(prompt,events,attendee).then(thing => {
                    //         let messageContent = ''
                    //         thing.forEach(item => messageContent += item);
                    //         msg.reply(messageContent);
                    //     });
                } else if (prompt.startsWith('daniel')) {
                    const randInt = Math.floor(Math.random() * 5);
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
                } else if (prompt.startsWith('leave') || prompt.startsWith('lv') || prompt.startsWith('uncoming') || prompt.startsWith('uncm')) {
                    let messageContent = ''
                    leave(prompt, events, attendee).forEach(item => messageContent += item);
                    msg.reply(messageContent);

                } else if (prompt.startsWith('help')) {
                    if (prompt.split(' ').length < 2) {
                        const response = "Allow me to introduce myself!\n\ni am the SHABbot!!\nI can help you with you shabbat meals, as well as other things\nThe way it works\n  Use !new, !join, !bring, !leave, !needs, and !list to make events and keeping track of what everyone's bringing\n  Use !shabtimes to find out when shabbat starts\n  Use !remind to set reminders for yourself";
                        msg.reply(response)
                    }

                    // } else if (prompt.startsWith('location')||prompt.startsWith('loc')) {
                    //     msg.reply(location(prompt,events))


                } else if (prompt.startsWith('kira')) {
                    msg.reply(`Wait wait Im not ready yet`);
                    setTimeout(() => { client.sendMessage(chat, "gimme one more second"); }, 3000);
                    setTimeout(() => { client.sendMessage(chat, "kk now im ready"); }, 5000);
                    setTimeout(() => { client.sendMessage(chat, "actually im not sure"); }, 6000);
                    setTimeout(() => { client.sendMessage(chat, "nvm its fine"); }, 9000);
                    setTimeout(() => { msg.reply("here's the kira command:") }, 10000);
                    setTimeout(() => { client.sendMessage(chat, "https://flashing-colors.com/") }, 11000);

                } else if (prompt.startsWith('remove') || prompt.startsWith('rm')) {
                    let messageContent = ''
                    remove(prompt, events, attendee, unfilteredEvents).forEach(item => messageContent += item)
                    msg.reply(messageContent);

                } else if (prompt.startsWith('fraydy')) {
                    const randInt = Math.floor(Math.random() * 2);
                    let link = 'https://www.youtube.com/watch?v=6Z8jBJ6w_7Q'
                    if (randInt) {
                        link = 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'
                    }
                    msg.reply(`I didn't know what to put for this one so here's a video of an elephant doing an handstand\n${link}`)
                } else if (prompt.startsWith('settime')) {
                    msg.reply(settime(prompt, events))
                } else if (prompt.startsWith('gil')) {
                    let link = ['https://www.youtube.com/watch?v=StTqXEQ2l-Y', 'https://www.youtube.com/watch?v=dQw4w9WgXcQ']
                    msg.reply(`everything is awesome!\n${link[Math.floor(Math.random() * 2)]}`)
                }
                else if (prompt.startsWith('surprise')) {
                    let link = ['https://www.youtube.com/watch?v=Bw3JHVExpbk&list=PLEYslXW1oN0e3QEWbKDou4Mto-y-4-STg', 'https://www.youtube.com/watch?v=gcm-QBDgnWM&list=PLEYslXW1oN0e3QEWbKDou4Mto-y-4-STg&index=3']
                    msg.reply(`everything is awesome!\n${link[Math.floor(Math.random() * 2)]}`)
                }
                else if (prompt.startsWith('count')) {
                    let link = 'https://www.chabad.org/holidays/sefirah/omer-count_cdo/jewish/Count-the-Omer.htm'
                    msg.reply(`you forgot that your suppose to count the omer!? ha! no, not tonight\n${link}`)
                }
                else if (prompt.startsWith('Amnon')) {
                    const randInt = Math.floor(Math.random() * 2);
                    let link = 'https://www.youtube.com/watch?v=D_IFNaTEXBA'
                    msg.reply(`Shalom Chabibi!\n${link}`)
                }
                else if (prompt.startsWith('gabe')) {
                    const randInt = Math.floor(Math.random() * 2);
                    let link = 'https://www.youtube.com/watch?v=oWgTqLCLE8k'
                    msg.reply(`time to party!\n${link}`)
                } else if (prompt.match(/\boren\b/gi)) {
                    msg.reply(`you found it ${attendee.id}! Make sure your parents aren't around ;)\nhttps://www.pornhub.com/view_video.php?viewkey=ph5f3b21528d0d5`)
                } else if (prompt.match(/^penis/gi)) {
                    msg.reply(`hahahha nice job heres a gift for your hard work: https://youtu.be/KhqjlVn9q4Y?t=117`)
                } else if (prompt.match(/\bbessie\b/gi)) {
                    msg.reply('bessie you need this tinyurl.com/tatter-totters')
                } else if (prompt.match(/^(shabbattimes|shabtimes)\b/gi)) {
                    const now = new Date()
                    var date = chrono.parse(prompt, now, { forwardDate: true }).at(0)
                    if (date == undefined) {
                        date = new Date()
                    } else {
                        date = date.date()
                    }
                    msg.reply(shabTimes(date, chat0.location, prompt, attendee))
                } else if (prompt.match(/^(shabbatLocation|shabloc|shablocation|chatlocation|location|setloc)\b/gi)) {
                    msg.reply(shabLocation(prompt, chat0))
                } else if (prompt.match(/^needs\b/gi)) {
                    msg.reply(needsCMD(prompt, events));
                } else if (prompt.match(/^remind\b/gi)) {
                    /*console.log(reminderCMD("remind me grapes", chat, testTimedList, "haifa",testUntimedList={}))                   console.log(reminders(testTimedList,testUntimedList))                                                           allEvents[chat]={timedList:testTimedList,untimedList:testUntimedList}                                                   console.log(doneCMD(chat,allEvents))
                   */
                    if (!chat0['untimedList']) {
                        chat0['untimedList'] = {}
                    }

                    if (!chat0['timedList']) {
                        chat0['timedList'] = {}
                    }
                    msg.reply(reminderCMD(prompt, chat, chat0['timedList'], chat0["location"], chat0['untimedList']))
                } else if (prompt.match(/^reminders\b/gi)) {
                    msg.reply(reminders(chat0.timedList, chat0.untimedList))
                } else if (prompt.match(/^send\b/gi)) {
                    if (!chat0['timedList']) {
                        chat0['timedList'] = {}
                    }
                    msg.reply(schedule(prompt, chat, allEvents))
                } else if (prompt.match(/^unsend\b/gi)) {
                    msg.reply(unscheduleMostRecent(chat, allEvents))
                } else if (prompt.match(/^scheduled\b/gi)) {
                    msg.reply(scheduled(chat, allEvents))
                } else if (prompt.match(/^nico\b/gi)) {
                    dontBeAPussy(client, chat)
                } else if (prompt.match(/^(fast|fasting)\b/gi)) {
                    const response = fastLister(new Date(), chat0.location)
                    if (response) {
                        msg.reply(response)
                    } else {
                        msg.reply("there's no fast day")
                    }
                } else if (prompt.match(/^snooze\b/gi)) {
                    msg.reply(snoozeCMD(prompt, chat, allEvents,))
                } else if (prompt.match(/^done\b/gi)) {
                    msg.reply(doneCMD(prompt, chat, allEvents, chat0["location"]))
                } else if (prompt.match(/^quickshab\b/gi)) {
                    const numText = parseInt(prompt.split(/\s/)[1])
                    console.log(numText)
                    if (isNaN(numText)) {
                        msg.reply("Command syntax: !quickshab [number of people]")
                        continue
                    }
                    allEvents[chat]["quickShab"] = []
                    msg.reply(quickshab(allEvents[chat]["quickShab"], numText) + "\n\nYou can sign yourself up by writing !bring [category]")
                } else if (prompt.match(/^(update|up|num|ppl)\b/gi)) {
                    const numText = parseInt(prompt.split(/\s/)[1])
                    if (isNaN(numText)) {
                        msg.reply("Command syntax: !update [number of people]")
                        continue
                    }
                    if (!allEvents[chat]["quickShab"] || allEvents[chat]["quickShab"].length == 0) {
                        msg.reply("You need to create a quickshab first by typing !quickshab [number of people]")
                        continue
                    }
                    msg.reply(updateNumber(allEvents[chat]["quickShab"][0], numText))
                } else if (prompt.match(/^(bring|br|bringing)\b/gi)) {
                    msg.reply(bringCmd(allEvents[chat]["quickShab"][0], prompt, attendee.id))
                } else if (prompt.match(/^assign\b/gi)) {
                    msg.reply(assignCmd(allEvents[chat]["quickShab"][0], prompt))
                } else if (prompt.match(/^(unbr|unbring)\b/gi)) {
                    msg.reply(unbringCmd(allEvents[chat]["quickShab"][0], attendee.id))
                } else if (prompt.match(/^unassign\b/gi)) {
                    const name = prompt.split(/\s+/).slice(1).join(" ");
                    if (!name) {
                        msg.reply("Command syntax: !unassign [name]")
                        continue
                    }
                    msg.reply(unassignCmd(allEvents[chat]["quickShab"][0], name))
                } else if (prompt.match(/^show\b/gi)) {
                    msg.reply(show(allEvents[chat]["quickShab"][0]))
                } else if (prompt.match(/^start\b/gi)) {
                    msg.reply(tovaTriggerStart(prompt))
                } else if (prompt.match(/^end\b/gi)) {
                    msg.reply(tovaTriggerEnd())
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
    } catch (error) {
        console.log(error)
        msg.reply("An error occurred, please try again")
    }
}

module.exports = { response }

//tests

let allEventsTest = { testChat: { events: [], teams: [], location: 'Haifa', timedList: {} } }
const testAttendee = { id: 'Gabe', number: '972587120601', guests: 0, food: 'nothing' }

function responseTest(text) {
    console.log(allEventsTest);
    const now = new Date()
    const msg = new testMsg('!' + text)
    msg.from = "testChat"
    response(new testClient(), msg, allEventsTest.testChat.events.filter(item => item.date > now), testAttendee, allEventsTest.testChat.events, allEventsTest)
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
    responseTest("remind me to pee in 3 minutes")
    responseTest('reminders')
    responseTest('remind uncle')
    responseTest('reminders')
}


//autoResponseTest()
