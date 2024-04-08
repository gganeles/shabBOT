const {isComing, findIndicies, countAst, newestIndex, removeCmdandIndicies, itemTyper} = require('./EventUtilities.js')

function updateAttendee(attendee, astNum, event) {
    if (!isComing(attendee.number, event)) {
        event.attendance.push(attendee)
        return `${attendee.id} will be attending ${event.eventName}\n`
    }
    return updateGuests(attendee, event, astNum)
}

function coming(prompt, eventList, attendee) {
    let indexList = findIndicies(prompt, eventList)
    let messageContent = []
    const astNum = countAst(prompt)
    if (prompt.split(/\s*,\s*|\s\s*/).length===1  & eventList.length>0) {
        messageContent.push(updateAttendee(attendee, astNum, eventList.at(newestIndex(eventList))))
    } else if (indexList.length===0) {
        indexList.push(newestIndex(eventList))
    }
    for (let i = 0; i < indexList.length;i++) {
        if (eventList.length>indexList.at(i)) {
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
        indexList.push(newestIndex(eventList))
    }
    for (let i = 0; i < indexList.length;i++) {
        if (eventList.length>indexList.at(i)) {
            messageContent.push(removeAttendee(attendee,eventList.at(indexList.at(i))))
        } else {
            messageContent.push('that event does not exist')
        }
    }
    return messageContent
}

function updateGuests(attendee, event, astNum) {
    for (let p = 0; p < event.attendance.length; p++) {       
        if (event.attendance.at(p).number === attendee.number & event.attendance.at(p).guests != astNum) {
            event.attendance.at(p).guests = astNum;
            return `number of guests updated\n`
        }
    }
    return "You're already signed up silly\n"
}

async function updateAttendeeFood(attendee,event) {
    for (let i = 0;i<event.attendance.length;i++){
        if (attendee.number===event.attendance.at(i).number) {
            event.attendance.at(i).food=attendee.food
            event.attendance.at(i).foodtype = await itemTyper(attendee.food)
            return `${attendee.id} is bringing ${attendee.food} to ${event.eventName}\nthe foodtype is ${event.attendance.at(i).foodtype}\n`
        }
    }
    attendee.foodtype = await itemTyper(attendee.food)
    event.attendance.push(attendee)
    return `${attendee.id} will be attending ${event.eventName} and bringing ${attendee.food}\nthe foodtype is ${attendee.foodtype}\n`
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

module.exports = {coming,leave,bringing, updateAttendee,removeAttendee}