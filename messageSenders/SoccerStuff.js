
const dayjs = require('dayjs')
const {newestIndex} = require('../eventstuff/EventUtilities.js')
const {updateAttendee,removeAttendee} = require('../eventstuff/updateAttendeeCommands.js') 
const dateFormat = 'ddd DD.MM.YYYY @ h:mm a'
const {timeTick, newSoccerEvent} = require('./TimedStuff.js')

function soccerList(event) {
    var messageContent = `Footie Game on ${dayjs(event.date).format(dateFormat)}:\n`

    for (let p = 0; p < 15 && p < event.attendance.length; p++) {
        var confirmStatus = ''
        if (event.attendance.at(p).isConfirmed) {
            confirmStatus = ' '+event.attendance.at(p).isConfirmed
        }
        if (event.attendance.at(p).isConfirmed && event.attendance.at(p).number == '972584938089') {
            confirmStatus = ' 🍪'
        }
        messageContent+=` ${p+1}) ${event.attendance.at(p).id}${confirmStatus}\n`
    }
    messageContent += 'People on the waitlist:\n'
    for (let p = 15; p < event.attendance.length; p++) {
        var confirmStatus = ''
        if (event.attendance.at(p).isConfirmed) {
            confirmStatus = ' '+event.attendance.at(p).isConfirmed
        }
        if (event.attendance.at(p).isConfirmed && event.attendance.at(p).number == '972584938089') {
            confirmStatus = ' 🍪'
        }
        messageContent+= ` ${p-14}) ${event.attendance.at(p).id}${confirmStatus}\n`
    }
    return messageContent
}




function soccerResponse(prompt,msg,attendee,soccerChat) {
    const now = new Date()
    let eventList = soccerChat.events.filter(item=> item.date>now)
    prompt = prompt.replace('!','')
    if (newestIndex(eventList) == -1 && prompt.match(/^(in|coming|list|confirm|alegay)/gi)) {
	    msg.reply('there is no upcoming game')
    } else if (prompt.match(/\s*^(in|coming)\s*/gi)) {
        updateAttendee(attendee,0,eventList.at(newestIndex(eventList)))
        msg.reply(soccerList(eventList.at(newestIndex(eventList))))
    } else if (prompt.match(/\s*^(list|ls)\s*/gi)) {
        msg.reply(soccerList(eventList.at(newestIndex(eventList))))
    } else if (prompt.match(/\s*^(leave|lv)\s*/gi)) {
        removeAttendee(attendee,eventList.at(newestIndex(eventList)))
    } else if (prompt.match(/\s*^(confirm|alegay)\s*/gi)) {
        confirm(attendee,prompt,eventList.at(newestIndex(eventList)))
    } else if (prompt.match(/\s*^add\s+/gi)) {
        addFriend(attendee,prompt,eventList.at(newestIndex(eventList)))
    } else if (prompt.match(/\s*^remove\s+/gi)) {
        removeFriend(attendee,prompt,eventList.at(newestIndex(eventList)))
    } else if (prompt.match(/\bhummus\b/gi) && attendee.number == "972586190901") {
        msg.reply('shut up ben')
    } /*else if (attendee.number == '972546390886' && prompt.match(/e/i)) {
	    msg.reply('danneth youve angered the gods with your very words')
    }*/ else if (prompt.match(/\blegwork\b/g) && attendee.number == '972587120601') {
        msg.reply('new event created');
        newSoccerEvent(now,soccerChat.events,prompt);
    } else if (prompt.match(/\b^maketeams\b/gi)) {
        allEvents[soccerChat].teams = createTeams(eventList.at(newestIndex(eventList)).attendance)
        const soccerteams = soccerChat.teams
        let messageContent = ''
        for (i in soccerteams) {
            messageContent += 'Team ' + (parseInt(i)+1) + ':'
            for (j in soccerteams.at(i)) {
                messageContent += '\n    '+soccerteams.at(i).at(j).id
            }
            messageContent += '\n\n'
        }
        msg.reply(messageContent)
    } else if (prompt.match(/\b^teams\b/gi)) {
        const soccerteams = soccerChat.teams
        let messageContent = ''
        for (i in soccerteams) {
            messageContent += 'Team ' + (parseInt(i)+1) + ':'
            for (j in soccerteams.at(i)) {
                messageContent += '\n    '+soccerteams.at(i).at(j).id
            }
            messageContent += '\n\n'
        }
        msg.reply(messageContent)
    }
}

function confirm(attendee,prompt,event) {
    for (let i = 0; i < event.attendance.length; i++) {
        if (event.attendance.at(i).friendOf===attendee.number && prompt.match(new RegExp("\\b"+event.attendance.at(i).id+"\\b","gi"))) {
            event.attendance.at(i).isConfirmed = '☑️'
        } else if (event.attendance.at(i).number==attendee.number) {
            event.attendance.at(i).isConfirmed = '☑️'
            if (prompt.split(/\s+/).length>1) {event.attendance.at(i).isConfirmed = prompt.split(/\s+/).at(1)}
        }
    }
}

function addFriend(attendee,prompt,event) {
    let friend = prompt.match(/(?<=add).+/i).at(0).trim()
    if (friend) {
        event.attendance.push({id: friend, friendOf: attendee.number, confirmStatus: ''})
        return `Added ${friend.at(0)} to ${event.eventName}`
    }
    return 'who do you want to add'
}

function removeFriend(attendee,prompt,event) {
    const friend = prompt.match(/(?<=remove).+/i).at(0).trim()
    if (friend) {
        for (i in event.attendance) {
            if (event.attendance.at(i).friendOf == attendee.number) {
                event.attendance.splice(i,1)
                return 'removed that other guy'
            }
        }
    }
    return 'you dont have any friends playing'
}

  
function createTeams(players) {
    if (players.length > 15) {
        players = players.slice(0,15);
    }
  
    const shuffledPlayers = players.sort(() => Math.random() - 0.5);
    const teamSize = Math.floor(players.length / 3);

    const teams = Array.from({ length: 3 }, () => []);

    let playerIndex = 0;
    for (let teamIndex = 0; teamIndex < 3; teamIndex++) {;
        if (playerIndex > players.length) {
            break;
        }
        for (let i = 0; i < teamSize; i++) {
            teams[teamIndex].push(shuffledPlayers[playerIndex]);
            playerIndex++;
        }
    }
    return teams;
}


//tests 

//attendee,eventList,allEvents
//


let allEventsTest = {soccerChatTest: {events:[],teams:[],location:'Haifa',timedList:{}}}

function soccerTest(prompt) {
    const msg = new testMsg()
    const attendee = {id: 'Gabe', number: '972587120601', guests: 0, food: 'nothing'}
    soccerResponse(prompt,msg,attendee,allEventsTest['soccerChatTest'])
}

function autoSoccerTest() {
    soccerTest('legwork next tuesday at 5')
    soccerTest('in')
    soccerTest('confirm')
    soccerTest('list')
    

    console.log("\nNow with timetick:\n\n")
    timeTick(new testClient(), allEventsTest, true)
    soccerTest('in')
    soccerTest('confirm')
    soccerTest('list')
    soccerTest('remove 0')
    soccerTest('list')

}

class testMsg {
    constructor(body='') {
        this.from = '972587120601@c.us'
        this.body = body
    }

    reply(text) {
        console.log("\nReply:\n", text)
    }

}

class testClient {
    constructor() {
    }

    sendMessage(chat,message) {
    console.log('\nSending the message:', message, '\nto chat:', chat)
    }
}
/*


setInterval(()=>console.log(new Date()), 5000)

*/
module.exports = {soccerResponse, soccerList, createTeams, testClient, testMsg, autoSoccerTest}
