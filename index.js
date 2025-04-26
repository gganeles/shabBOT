const fs = require('fs');
const dayjs = require('dayjs');

const { Client, LocalAuth } = require('whatsapp-web.js');

const qrcode = require('qrcode-terminal');
const responder = require('./messageSenders/ResponseCmd.js')
const { timeTick } = require('./messageSenders/TimedStuff.js')
const { timedMsg } = require('./utilitystuff/ScheduledSendandReminders.js')

const { soccerResponse, autoSoccerTest } = require('./messageSenders/SoccerStuff')
const piChromiumPath = '/usr/bin/chromium'
const piAuthPath = '/home/pi/shabbot/.wwebjs_auth'
const jsonPath = '/home/pi/shabbot/saved-events.json';
//const jsonPath = 'saved-events.json';


const client = new Client({
    puppeteer: { 
        args: ['--no-sandbox'], 
        headless: true, 
        executablePath: piChromiumPath }, 
    webVersionCache: {
        type: 'remote',
        remotePath: 'https://raw.githubusercontent.com/wppconnect-team/wa-version/main/html/2.2410.1.html',
    },
    authStrategy: new LocalAuth({
        dataPath: piAuthPath
    }),
    clientId: '12038025238@c.us'
});

//const client = new Client({ puppeteer: {args: ['--no-sandbox'], headless: true , executablePath: piChromiumPath }, authStrategy: new LocalAuth({dataPath: piAuthPath}),webVersionCache: new LocalWebCache(), clientId: '12038025238@c.us' });
//const client = new Client({puppeteer: {args: [], headless: false }, authStrategy: new LocalAuth(), clientId: '12038025238@c.us' });

const dateFormat = 'ddd DD.MM.YYYY @ h:mm a';

var events = []; /*= [{ // Solely for easier auto-completion
    eventName: String.prototype, date: Date.prototype, attendance: [{ id: String.prototype, name: String.prototype, guests: string.prototype, food: string prototype }]
}];*/

let allEvents = {};

const isoDateRegex = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2}(?:\.\d*))(?:Z|(\+|-)([\d|:]*))?$/;

const now_test = new Date()
console.log(dayjs(now_test).format(dateFormat))

function record(arr, jsonPath) {
    fs.writeFile(jsonPath, JSON.stringify(arr), (err) => {
        if (err) {
            console.log(err);
        }
        console.log('the database has been updated')
    });
}


if (fs.existsSync(jsonPath)) {
    try {
        allEvents = JSON.parse(fs.readFileSync(jsonPath, 'utf8'), (key, value) => {
            if (typeof value === "string" && isoDateRegex.exec(value)) { return new Date(value); } else if (key === 'timedList') {
                let transformedTimedList = {}
                for (const [key0, i] of Object.entries(value)) {
                    transformedTimedList[key0] = new timedMsg(i.message, i.time, i.chat, i.id, i.type);
                }
                return transformedTimedList
            } return value;
        });
    } catch (err) {
        console.log(err);
    }
} else {
    console.log('but why tho')
}

//allEvents = convertListToObj(allEvents)
//record(allEvents,jsonPath)
//console.log(allEvents)

client.on('qr', (qr) => {
    // NOTE: This event will not be fired if a session is specified.
    console.log('QR RECEIVED', qr);
    qrcode.generate(qr, { small: true });
});

client.on('authenticated', () => {
    console.log('AUTHENTICATED');
});

client.on('auth_failure', msg => {
    // Fired if session restore was unsuccessfull
    console.error('AUTHENTICATION FAILURE', msg);
});

client.on('ready', () => {
    console.log('\x1b[32m%s\x1b[0m', 'READY');
    console.log(`Logged in as ${client.info.pushname} (${client.info.wid._serialized})`);
    setInterval(timeTick, 1000 * intervalSize, client, allEvents)
});

//client.initialize();
let state = []


let is_global = false
client.on('message', async msg => {
    try {
        const contactname = (await msg.getContact()).pushname;
        const gabeName = (await msg.getContact()).name;
        const phoneNumber = (await msg.getContact()).id.user;
        let attendee = { id: contactname, number: phoneNumber, guests: 0, food: 'nothing' }
        const chatNum = msg.from;
        if (gabeName != undefined) {
            attendee.id = gabeName
        } else if (attendee.number == '972587120601') {
            attendee.id = 'God'
        }
        console.log(`${contactname} ${phoneNumber}`);
        console.log(chatNum)

        is_global = !(msg.body.match(/\s+-private\b/gi))
        console.log(is_global)
        /*allEvents['global'] = Object.entries(allEvents).filter(([key,value])=> key !== '120363029029121540@g.us' && key !== 'global').reduce((acc, item) => {
            return acc.concat(item[1].events.filter((event)=> (event.date) && !(event.isPrivate)))
            },[])*/
        if (allEvents['global'] === undefined) { allEvents['global'] = [] }
        if (is_global) {
            events = allEvents['global']
        } else if (allEvents[chatNum]) {
            events = allEvents[chatNum].events
        }
        if (!allEvents[chatNum]) {
            allEvents[chatNum] = { events: [], location: 'Haifa' }
            events = allEvents[chatNum].events
        }
        record(allEvents, jsonPath);

        let message = msg.body
        const now = new Date()
        if (chatNum == soccerChat) {
            soccerResponse(message, msg, attendee, allEvents[soccerChat])
            record(allEvents, jsonPath)
        } else {
            //saltyBot(msg,client,contactname)
            responder.response(client, msg, events.filter(event => event.date > now), attendee, is_global ? allEvents['global'] : allEvents[chatNum].events, allEvents)
            record(allEvents, jsonPath)
        }
    } catch (err) {
        console.log(err)
        msg.reply('uh oh sp! there was an error!')
    }
});


//upcomingEvents.at(eventsThis - 1).attendance = upcomingEvents.at(eventsThis - 1).attendance.filter(id => !id.startsWith(attendee));

client.initialize();
