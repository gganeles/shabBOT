const chrono = require('chrono-node')
const dayjs = require('dayjs')
const { findIndicies, removeCmdandIndicies, newestIndex } = require('./EventUtilities.js');
const dateFormat = 'ddd DD.MM.YYYY @ h:mm a';


function updateLocation(eventList, index, place) {
    eventList.at(index).location = place;
    return `the location of ${eventList.at(index).eventName} was set to ${place}`
}

function location(prompt,eventList) {
    const indexList = findIndicies(prompt,eventList)
    const location = removeCmdandIndicies(prompt,eventList)
    if (eventList.length===0) {
        return 'there are no events what are you doing'
    } else if (eventList.length>0 & indexList.length===0) {
        return updateLocation(eventList, newestIndex(eventList), location)
    } else if (indexList.length===0) {
        return "Usage: !location <event name or index> <where>\n"
    }
    for (let i = 0; i < indexList.length;i++) {
        if (eventList.length>indexList.at(i)) {
            return updateLocation(eventList, indexList.at(i), location)
        } else {
            return 'that event does not exist'
        }
    }
}
function newName(prompt, event) {
    return prompt.replace(new RegExp(`\\b${event.eventName}(\\s|)|rename\\s|rn\\s|\\b\\d\\b`,'gi'),'').trim()
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

function remove(prompt,eventList,attendee,EVENTS_FOR_REMOVAL) {
    const indexList = findIndicies(prompt,eventList)
    let messageContent = []
    if (eventList.length===0) {
        return ['there are no events what are you doing']
    }
    if (indexList.length===0) {
        return ["Usage: !remove <event name or index>"]
    }
    for (let i = 0; i < indexList.length;i++) {
        if (indexList.at(i) >= eventList.length) {
            messageContent.push('that event already does not exist you insolant monkey\n')
        } else {
            const makerId = eventList.at(indexList.at(i)).attendance.at(0).number
            if (!(makerId === attendee.number || attendee.number === '972587120601')) {
                messageContent.push('you have no permission you billiards baby\n')
            } else if (eventList.length>indexList.at(i)) {
                messageContent.push(`Removing event ${eventList.at(indexList.at(i)).eventName}\n`)
                const eventToRemove = EVENTS_FOR_REMOVAL.findIndex((event) => event.eventName === eventList.at(indexList.at(i)).eventName)
                EVENTS_FOR_REMOVAL.splice(eventToRemove,1)
            } else {
                messageContent.push('that event already does not exist you insolant monkey\n')
            }
        }
    }
    return messageContent
}

function settime(prompt,eventList) {
    const indexList = findIndicies(prompt,eventList)
    const now = new Date()
    const date = chrono.parse(prompt, now, { forwardDate: true }).at(0)
    if (date === undefined) {
        return "Usage: !Try rephrasing your time, for example 17:00 or 5pm\n"
    } else if (indexList.length===0) {
        indexList.push(newestIndex(eventList))
    }
    for (let i = 0; i < indexList.length;i++) {
        if (eventList.length>indexList.at(i)) {
            eventList.at(indexList.at(i)).date.setHours(date.date().getHours(),date.date().getMinutes())
            return `${eventList.at(indexList.at(i)).eventName} will now occur on ${dayjs(eventList.at(indexList.at(i)).date).format(dateFormat)}\n`
        } else {
            return 'that event does not exist\n'
        }
    }
}

module.exports = { location, rename, remove, settime }