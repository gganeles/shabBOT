const {HebrewCalendar, Location} = require('@hebcal/core')
const date = require('date.js/lib/date')


function fastLister (date_for_testing, location = 'Haifa') {
    let date = new Date()
    let farEndDate = new Date(date.getDate+100)
    let tomorrow = new Date(date)
    if (date_for_testing) {
        date = date_for_testing
        tomorrow = farEndDate
    }
    date.setHours(12,0,0,0)

    tomorrow.setDate(date.getDate() + 1)
    const options = {
        start: date,
        end: tomorrow,
        candlelighting: true,
        location: Location.lookup(location),
    };

    const events = HebrewCalendar.calendar(options);

    [ fast_start, fast_end ] = events.filter(event => event.constructor.name === 'TimedEvent' || (event.linkedEvent && event.linkedEvent.desc.match(/yom\skippur/gi)))
    //console.log(events)
    if (fast_start && fast_end) {
        return `${fast_end.linkedEvent.desc} will start in ${location} at ${fast_end.eventTime.getDate()==fast_start.eventTime.getDate()?'tomorrow at '+fast_start.eventTimeStr:fast_start.eventTimeStr} and will end tomorrow at ${fast_end.eventTimeStr}`
    } else {
        return false
    }
}

module.exports = { fastLister }

//tests

//fastLister(new Date())