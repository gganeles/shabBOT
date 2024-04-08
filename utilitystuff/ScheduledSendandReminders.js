const dayjs = require('dayjs')
const chrono = require('chrono-node')


class timedMsg {
    constructor(message, time, chat, id, type) {
        this.message = message
        this.time = time
        this.chat = chat
	    this.id = id
        this.type = type
    }

    send(client,allEvents) {
        if (this.type == 'remind') {
            client.sendMessage(this.chat,this.toReminderString())
        } else {
            client.sendMessage(this.chat,this.message)
        }
        console.log('sending message and then deleting '+this.id.toString())
	    delete allEvents[this.chat].timedList[this.id.toString()]
    }

    tick(date,client,allEvents) {
        if (date >= this.time) {
            this.send(client,allEvents)
            return true
        } else {
            return false
        }
    }

    toReminderString() {
        return `you wanted me to remind you:\n "${this.message}"`
    }

}

let timedId = 0
function reminderCMD(prompt,chat,timedList) {
    if (prompt.split(' ').length==1) {
        return 'command usage: !remind   thing to remind   when'
    } else {
        prompt = prompt.split(' ').slice(1,).join(' ')
        const now = new Date()
        const dateObj = chrono.parse(prompt, now, { forwardDate: true }).at(0);
        
        if (dateObj == undefined || dateObj.date() < now) {
            return 'Try rephrasing your date, like "Saturday at 14" or "Tomorrow at 6pm"'
        } else {
            const message = prompt.slice(0,dateObj.index).trim().replace(/(^(me\sto\s|me\s))|\sin$/gi,'')
            timedList[timedId.toString()] = new timedMsg(message,dateObj.date(),chat,timedId,'remind')
	        timedId ++
            return 'Ok, I will remind you "'+message+'"'+whichWeek(dateObj.date())+dayjs(dateObj.date()).format(coolDateFormat)
        }
    }
}

const coolDateFormat = 'ddd @ h:mm a'

function whichWeek(date) {
    const now = new Date()
    const weekDiff = Math.floor((date-now)/(1000*60*60*24*7))
    if (weekDiff < 0) {return ' already happened '}
    if (weekDiff == 0) {return ' '}
    if (weekDiff == 1) {return ' next '}
    if (weekDiff > 1) {return ' in '+weekDiff.toString()+' weeks '}
}

function reminders(timedList) {
    let messageContent = 'Current reminders:'
    for (let key in timedList) {
        if (timedList[key].type == 'remind')
            messageContent+="\n  "+timedList[key].message+whichWeek(timedList[key].time)+dayjs(timedList[key].time).format(coolDateFormat)
    }
    return messageContent
}

function scheduled(chat,allEvents) {
    let messageContent = 'Scheduled Messages:'
    for (let idNum in allEvents[chat].timedList) {
        for (let key in allEvents) {
            if (key !== chat && allEvents[key].timedList[idNum]) {
                messageContent+='\n  "'+allEvents[key].timedList[idNum].message+'" to '+key.replace('@c.us','')
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
        console.log(prompt.substring(contact.index,))
        const now = new Date()
        const dateObj = chrono.parse(prompt, now, { forwardDate: true }).at(0);
        if (dateObj == undefined || dateObj.date() < now) {
            return 'Try rephrasing your date, like "Saturday at 14" or "Tomorrow at 6pm"'
        } else {
            const message = prompt.slice(0,contact.index).trim().replace(new RegExp('(\\sto$)|\\s+'+contact+'|\\sin$','gi'),'')
            if (!contact) {
                return 'put in a phone number'
            }
            let contactChat = contact+'@c.us' 
            if (!allEvents[contactChat]) {
                allEvents[contactChat] = {timedList:{}}
            }
            if (!allEvents[chat]) {
                allEvents[chat] = {timedList:{}}
            }
            allEvents[contactChat]['timedList'][timedId.toString()] = new timedMsg(message,dateObj.date(),contactChat,timedId,'schedule',chat.replace('@c.us',''))
            allEvents[chat]['timedList'][timedId.toString()] = new timedMsg('Sent scheduled message to '+contact,dateObj.date(),chat,timedId,'schedule')
	        timedId ++
            return 'Ok, I will send "'+message+'" to '+contact+" on"+whichWeek(dateObj.date())+dayjs(dateObj.date()).format(coolDateFormat)
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


module.exports = {reminderCMD,reminders,timedMsg, schedule, unscheduleMostRecent, scheduled}
