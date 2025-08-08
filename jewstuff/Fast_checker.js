const {HebrewCalendar, Location} = require('@hebcal/core')


function fastLister (date, location = 'Haifa') {
    let tomorrow = new Date(date)
    tomorrow = new Date(tomorrow.setDate(date.getDate() + 1))
    date.setHours(1,0,0,0)

    const options = {
        start: date,
        end: tomorrow,
        candlelighting: true,
        location: Location.lookup(location),
    };

    const events = HebrewCalendar.calendar(options);

    let fast_start;
    let fast_end;
    [ fast_start, fast_end ] = events.filter(event => event.constructor.name === 'TimedEvent' || (event.linkedEvent && event.linkedEvent.desc.match(/yom\skippur/gi)))
    if (fast_start && fast_end) {
        return `${fast_start.linkedEvent.desc} will start in ${location} at ${fast_end.eventTime.getDate()==fast_start.eventTime.getDate()?'tomorrow at '+fast_start.eventTimeStr:fast_start.eventTimeStr} and will end tomorrow at ${fast_end.eventTimeStr}`
    } else if (fast_start) {
        return `${fast_start.linkedEvent.desc} will end in ${location} at ${fast_start.eventTimeStr}`
    } else {
        return false
    }
}

// let yesterday = new Date()
// yesterday = new Date(yesterday.setDate(yesterday.getDate()-1))
//console.log(fastLister())

module.exports = { fastLister }

//tests