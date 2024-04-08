
const {HebrewCalendar, HDate, Location, Event} = require('@hebcal/core')
const date = require('date.js/lib/date')
const { removeAttendee } = require('../eventstuff/updateAttendeeCommands')
const locationList = ['Ashdod', 'Atlanta', 'Austin', 'Baghdad', 'Beer Sheva', 'Berlin', 'Baltimore', 'Bogota', 'Boston', 'Budapest', 'Buenos Aires', 'Buffalo', 'Chicago', 'Cincinnati', 'Cleveland', 'Dallas', 'Denver', 'Detroit', 'Eilat', 'Gibraltar', 'Haifa', 'Hawaii', 'Helsinki', 'Houston', 'Jerusalem', 'Johannesburg', 'Kiev', 'La Paz', 'Livingston', 'Las Vegas', 'London', 'Los Angeles', 'Marseilles', 'Miami', 'Minneapolis', 'Melbourne', 'Mexico City', 'Montreal', 'Moscow', 'New York', 'Omaha', 'Ottawa', 'Panama City', 'Paris', 'Pawtucket', 'Petach Tikvah', 'Philadelphia', 'Phoenix', 'Pittsburgh', 'Providence', 'Portland', 'Saint Louis', 'Saint Petersburg', 'San Diego', 'San Francisco', 'Sao Paulo', 'Seattle', 'Sydney', 'Tel Aviv', 'Tiberias', 'Toronto', 'Vancouver', 'White Plains', 'Washington DC', 'Worcester']
const dayjs = require('dayjs')
const dateFormat = 'h:mm a';
const d = require('dayjs/locale/de')


function LocationFromList(prompt) {
    for (location of locationList) {
        if (prompt.match(new RegExp('\\b'+location+'\\b','gi'))){
            return location
        }
    }
    return 0
}

function shabTimes(date, location1, prompt, attendee) {
    var location = location1
    for (item in locationList) {
        if (prompt.match(new RegExp(locationList[item],'i'))){
            location = locationList[item]
        }
    }
    date.setHours(0,0,0,0)
    const now = date
    const tomorrow = new Date(now)
    tomorrow.setDate(now.getDate() + (7- now.getDay()))
    const options = {
      start: now,
      end: tomorrow,
      candlelighting: true,
      location: Location.lookup(location),
    };
    const events = HebrewCalendar.calendar(options);
    if (!events.filter(item => item.getDesc() == 'Candle lighting').length && events.filter(item => item.getDesc() == 'Havdalah').length) {
        return `Havdalah time in ${location} is ${events.filter(item => item.getDesc() == 'Havdalah').at(0).eventTimeStr}`
    } else if (!events.filter(item => item.getDesc() == 'Havdalah').length) {
        return 'oh you poor thing its not shabbat...'
    }
    console.log(events.filter(item => item.getDesc() == 'Candle lighting'));
    let candleLighting = events.filter(item => item.getDesc() == 'Candle lighting').at(0).eventTimeStr
    let havdalahTime = events.filter(item => item.getDesc() == 'Havdalah').at(0).eventTimeStr
    const timez = events.filter(item => item.getDesc() == 'Candle lighting').at(0).location.tzid
    if (attendee.number == '12063214745') {
        candleLighting = dayjs(convertToTimeZone(events.filter(item => item.getDesc() == 'Candle lighting').at(0).eventTime,timez)).format(dateFormat)
        havdalahTime = dayjs(convertToTimeZone(events.filter(item => item.getDesc() == 'Havdalah').at(0).eventTime,timez)).format(dateFormat)
    }
    return `Shabbat in ${location} starts this week at ${candleLighting} and ends at ${havdalahTime}`
}

function shabLocation(prompt,chat) {
    const location = LocationFromList(prompt)
    if (location) {
        chat.location = location
        return `chat location was set to ${location}`
    } else {
        return "that location isn't in our database"
    }
}

function convertToTimeZone(inputDateString, targetTimeZone) {
    const inputDate = new Date(inputDateString);
  
    // Create a new Date object with the same timestamp but in the target time zone
    const targetDate = new Date(inputDate.toLocaleString('en-US', { timeZone: targetTimeZone }));
  
    return targetDate;
}
  
module.exports = { shabTimes, shabLocation }



//const that = new Date()
//console.log(dayjs(convertToTimeZone(that,"Asia/Jerusalem")).format(dateFormat))

//console.log(shabTimes(that,'New York','shabtimes',{number:'12063214745'}));