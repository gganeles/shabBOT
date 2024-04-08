const { findIndicies, newestIndex, sumGuests } = require('./EventUtilities.js');
const dayjs = require('dayjs')
const dateFormat = 'ddd DD.MM.YYYY @ h:mm a';


function listEvents(events,requestedindex) {
    var messageContent = ``;
    if (requestedindex===-1) {
        messageContent+=`There are ${events.length} events planned.\n`
        events.forEach((event, index) => messageContent += `${index + 1}) ${event.eventName} on ${dayjs(event.date).format(dateFormat)}, ${event.attendance.length+sumGuests(event)} coming\n`);
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
                food = ` - ${event.attendance.at(j).food}`
            }
            messageContent += `  ${event.attendance.at(j).id}${num_of_guests}${food}\n`;
        }
    }
    return messageContent
}

function needsCMD(prompt,events) {
    const needs_by_amt = prompt.match(/(?<=\*)\d+/i);
    if (needs_by_amt) {
        return needs_from_amount(parseInt(needs_by_amt.at(0)));
    }
    let indexList = findIndicies(prompt,events)
    if (events.length == 0) {
        return 'No events found';
    }
    if (indexList.length == 0) {
        indexList.push(newestIndex(events));
    }
    let messageContent = ''
    indexList.forEach((index) => {
        if (events.length > index) {
            messageContent += `${events[index].eventName} still needs:\n`
            messageContent += needs(events[index])
        } else {
            messageContent += 'That event does not exist'
        }
    })
    return messageContent
}

function needs_from_amount(n) {
    const preset = {
        'main_dish': Math.ceil((n+1)/4),
        'side_dish': Math.ceil((n)/6),
        'plastics': n>5 ? Math.ceil(n/20) : 0,
        'drinks': n>5 ? Math.ceil(n/12) : 0,
        'wine': Math.ceil(n/6),
        'challah': Math.ceil(n/10),
        'dips': Math.ceil((n+1)/10),
        'dessert': Math.ceil((n-5)/8)
    }
    return Object.entries(preset).filter(([key, value]) => value > 0).map(([key, value]) => `${key}: ${value}`).join('\n').trim().replace(/main_dish/g, 'Main Dishes').replace(/side_dish/g, 'Side Dishes').replace(/plastics/g, 'Plastics').replace(/drinks/g, 'Drinks').replace(/wine/g, 'Wine').replace(/challah/g, 'Challah').replace(/dips/g, 'Dips').replace(/dessert/g,'Dessert')
}

function needs(event) {
    const n = event.attendance.length+sumGuests(event)
    if (n==0) {
        return 'theres nobody coming what could you need'
    }
    const preset = {
        'main_dish': Math.ceil((n+1)/4),
        'side_dish': Math.ceil((n)/6),
        'plastics': n>5 ? Math.ceil(n/20) : 0,
        'drinks': n>5 ? Math.ceil(n/12) : 0,
        'wine': event.eventName.match(/dinner/i) ? Math.ceil(n/6) : 1,
        'challah': Math.ceil(n/10),
        'dips': Math.ceil((n+1)/10),
        'dessert': Math.ceil((n-5)/8)
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
    return Object.entries(preset).filter(([key, value]) => value > 0).map(([key, value]) => `${key}: ${value}`).join('\n').trim().replace(/main_dish/g, 'Main Dishes').replace(/side_dish/g, 'Side Dishes').replace(/plastics/g, 'Plastics').replace(/drinks/g, 'Drinks').replace(/wine/g, 'Wine').replace(/challah/g, 'Challah').replace(/dips/g, 'Dips').replace(/dessert/g,'Dessert')
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

module.exports = { list, needsCMD };
