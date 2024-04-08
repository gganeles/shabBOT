const chrono = require('chrono-node');
const dayjs = require('dayjs');
const { findIndicies } = require('./EventUtilities.js');
const { record } = require('./EventUtilities.js');
const dateFormat = 'ddd DD.MM.YYYY @ h:mm a';



function newEvent(name, date, creator, eventList) {
    eventList.forEach(event => event.isNewest = false)
    eventList.push({ eventName: name, date: date.date(), attendance: [creator], isNewest: true});
    console.log(eventList)
    eventList.sort(function(a, b){
    return a.date - b.date;
    });
    return `There's a new "${name}" on ${dayjs(date.date()).format(dateFormat)}.\n`
}

function createEvent(prompt, EVENTS_FOR_REMOVAL, creator) {
    const now = new Date();
    console.log(dayjs(now).format(dateFormat))
    prompt = prompt.replace(new RegExp('\\s(on)\\b','gi'),'');
    prompt = prompt.replace('.','/');
    let date = chrono.en.GB.parse(prompt, now, { forwardDate: true }).at(0);
    if (date === undefined) {
        date = chrono.parse(prompt, now, { forwardDate: true }).at(0)
    }
    if (prompt.trim().split(' ').length===1) {
        let messageContent = ''
        const dinnertime = chrono.parse('Upcoming friday night at 8pm', now, { forwardDate: true }).at(0)
        const lunchtime = chrono.parse('Upcoming saturday at 1pm', now, { forwardDate: true }).at(0)
        messageContent+=newEvent('Shabbat Dinner',dinnertime,creator,EVENTS_FOR_REMOVAL)
        messageContent+=newEvent('Shabbat Lunch',lunchtime,creator,EVENTS_FOR_REMOVAL)
        messageContent+=' - To list all current events, type !list\n - You can edit your event with !rename, !settime, or !location'
        return messageContent
    } else if (date === undefined) {
        return `Usage: "!new <event name> <when>"`
    } else if (date.date()<now) {
        return 'It is already past that time'
    } else {
        const name = prompt.substr(prompt.split(/\s*,\s*|\s/).at(0).length+1, date.index - (prompt.split(/\s*,\s*|\s/).at(0).length+2)).replace(/\s+in$/gi,'');
        return newEvent(name, date, creator, EVENTS_FOR_REMOVAL)+' - To list all current events, type !list\n - You can edit your event with !rename, !settime, or !location'

    }
}


module.exports = { createEvent, newEvent }