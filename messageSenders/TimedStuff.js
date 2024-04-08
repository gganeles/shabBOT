const {newestIndex} = require('../eventstuff/EventUtilities.js')
const {fastLister} = require('../jewstuff/Fast_checker.js')
const attendees = require('../eventstuff/updateAttendeeCommands.js')
const chrono= require('chrono-node')

const jsonPath = '/home/pi/shabbot/saved-events.json';

const eighteen20 = '120363154021870149@g.us'
const shabbatChat = '120363042348096510@g.us'
const intervalSize = 5;
const DidYouHearThat = '12169789434-1427562427@g.us'

//const soccerChat = 'soccerChat'
function timeTick(client, allEvents, soccerDebug) {
    const date = new Date()
    let soccerChat = '120363029029121540@g.us'
    if (soccerDebug) {
        soccerChat = 'soccerChatTest'
    }
    if (allEvents[soccerChat] !== undefined||soccerDebug) {
        soccerKicker(date,soccerChat, client, allEvents[soccerChat].events.at(newestIndex(allEvents[soccerChat].events)))   
        soccerListTimer(date,soccerChat, client, allEvents, soccerDebug)
    }
    wordOfTheDay(date, "120363208220182956@g.us", client)
    eighteenNineteen(date, eighteen20, client)
    takeLabPictures(date,client)
    fastReminderTrigger(date, shabbatChat, client)
    sendEventTrigger(date,client,allEvents)
    return
}

function eighteenNineteen(date, chat, client) {
    if (date.getHours() == 18 && date.getMinutes() == 19 && date.getDay() < 6 && date.getSeconds() < intervalSize) {
        client.sendMessage(chat, "It's that time of day again!")
    }
}

function wordOfTheDay(date, chat, client) {
    if (date.getHours() == 18 && date.getMinutes() == 30 && date.getDay() < 6 && date.getSeconds() < intervalSize) {
        client.sendMessage(chat, "Studious people my ass")
        client.sendMessage(chat, "More like word of the year amirite")
        client.sendMessage(chat, "Its 6:30 in Haifa - let's go back to work")
    }
}

function takeLabPictures(date,client) {
    if (date.getHours() == 10 && date.getMinutes() == 0 && date.getDay() == 3 && date.getSeconds() < intervalSize) {
        client.sendMessage("120363154424500569@g.us","Hello everyone! This is your friend the formula bot coming to remind you to video your work in the formula workshop using Timelapse🙂")
    }
}

function soccerListTimer(date, chat, client, allEvents, soccerDebug) {
    if ((date.getDay() == 0 && date.getHours() == 20 && date.getMinutes() == 58 && date.getSeconds() < intervalSize)||soccerDebug) {
        client.sendMessage(chat, 'footie list time')
        newSoccerEvent(date, allEvents[chat].events,'upcoming tuesday at 9pm')
    } 
}

function newSoccerEvent(date0, eventList, when) {
    const date = chrono.parse(when, date0, {forwardDate: true}).at(0).date()
    eventList.forEach(event => event.isNewest = false)
    eventList.push({ eventName: 'footie game', date: date, attendance: [], isNewest: true, isSoccer: true})
}

function soccerKicker(date, chat,client,event){
    if (date.getDay() == 2 && date.getHours() == 12 && date.getMinutes() == 0 && date.getSeconds() < intervalSize) {
        for (let i=0; i<event.attendance.length; i++) {
            if (event.attendance.at(i).isConfirmed) {
                attendees.removeAttendee(event.attendance.at(i))
            }
        }
        for (let i=0; i<event.attendance.length; i++) {
            event.attendance.at(i).isConfirmed = false
        }
        client.sendMessage(chat,"People have been kicked, here's the new list:\n"+soccerList(event))
    }
}

function fastReminderTrigger(date, chat, client, location = 'Haifa') {
    if (date.getHours() == 17 && date.getMinutes() == 0 && date.getSeconds() < intervalSize && fastLister(date)) {
        client.sendMessage(chat,fastLister(date,location))
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

//tests

//let allEvents = {soccerChat: {events:[],teams:[],location:'Haifa',timedList:{}}}


module.exports = { timeTick, newSoccerEvent }

