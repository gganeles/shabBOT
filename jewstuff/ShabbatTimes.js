
const {HebrewCalendar, HDate, Location, Event} = require('@hebcal/core');
const locationList = ['Ashdod', 'Atlanta', 'Austin', 'Baghdad', 'Beer Sheva', 'Berlin', 'Baltimore', 'Bogota', 'Boston', 'Budapest', 'Buenos Aires', 'Buffalo', 'Chicago', 'Cincinnati', 'Cleveland', 'Dallas', 'Denver', 'Detroit', 'Eilat', 'Gibraltar', 'Haifa', 'Hawaii', 'Helsinki', 'Houston', 'Jerusalem', 'Johannesburg', 'Kiev', 'La Paz', 'Livingston', 'Las Vegas', 'London', 'Los Angeles', 'Marseilles', 'Miami', 'Minneapolis', 'Melbourne', 'Mexico City', 'Montreal', 'Moscow', 'New York', 'Omaha', 'Ottawa', 'Panama City', 'Paris', 'Pawtucket', 'Petach Tikvah', 'Philadelphia', 'Phoenix', 'Pittsburgh', 'Providence', 'Portland', 'Saint Louis', 'Saint Petersburg', 'San Diego', 'San Francisco', 'Sao Paulo', 'Seattle', 'Sydney', 'Tel Aviv', 'Tiberias', 'Toronto', 'Vancouver', 'White Plains', 'Washington DC', 'Worcester'];
const dayjs = require('dayjs');
const dateFormat = 'h:mm a';
const fs = require('fs');
const Papa = require('papaparse');
const fileContent = fs.readFileSync('./jewstuff/geoNamesList.csv', 'utf8');

// Parse the CSV
const citiesObj = Papa.parse(fileContent, {
  header: true,
    skipEmptyLines: true
});

function LocationFromList(prompt) {
    for (const location of locationList) {
        if (prompt.match(new RegExp('\\b'+location+'\\b','gi'))){
        return location;
        }
    }
    return 0;
}

function shabTimes(date, location1, prompt, attendee) {
    var locationObj;
    var location = location1;
    var cityName = location1;
    const promptList = prompt.split(/\s/);
    let flag = false;
    for (const item in locationList) {
        if (prompt.match(new RegExp(locationList[item],'i'))){
            location = locationList[item];
            cityName = location;
            flag = true;
            break;
        }
    }

    if (!flag&&promptList.length > 1) {
        const name = promptList.slice(1).join(' ').toLowerCase();
        flag = false;
        for (const item in citiesObj.data) {
            if (name == citiesObj.data[item]["ASCII Name"].toLowerCase()) {
                const [lat,lng] = citiesObj.data[item]["Coordinates"].split(',');
                locationObj = new Location(parseFloat(lat), parseFloat(lng), citiesObj.data[item]["Country name EN"]=="Israel", citiesObj.data[item]["Timezone"]);
                cityName = citiesObj.data[item]["ASCII Name"];
                flag= true;
                break;
            }
        }
        if (!flag) {
            locationObj = Location.lookup(location1);
        }
    } else {
        locationObj = Location.lookup(location);
    }
    date.setHours(0,0,0,0);
    const now = date;
    const tomorrow = new Date(now);
    tomorrow.setDate(now.getDate() + (7- now.getDay()));
    const options = {
      start: now,
      end: tomorrow,
      candlelighting: true,
      location: locationObj,
    };
    const events = HebrewCalendar.calendar(options);
    //console.log(events)
    if (!events.filter(item => item.getDesc() == 'Candle lighting').length && events.filter(item => item.getDesc() == 'Havdalah').length) {
        return `Havdalah time in ${location} is ${events.filter(item => item.getDesc() == 'Havdalah').at(0).eventTimeStr}`;
    } else if (!events.filter(item => item.getDesc() == 'Havdalah').length) {
        return 'oh you poor thing its not shabbat...';
    }
    //console.log(events.filter(item => item.getDesc() == 'Candle lighting'));
    let candleLighting = events.filter(item => item.getDesc() == 'Candle lighting').at(0).eventTimeStr;
    let havdalahTime = events.filter(item => item.getDesc() == 'Havdalah').at(0).eventTimeStr;
    const timez = events.filter(item => item.getDesc() == 'Candle lighting').at(0).location.tzid;
    if (attendee.number == '12063214745') {
        candleLighting = dayjs(convertToTimeZone(events.filter(item => item.getDesc() == 'Candle lighting').at(0).eventTime,timez)).format(dateFormat);
        havdalahTime = dayjs(convertToTimeZone(events.filter(item => item.getDesc() == 'Havdalah').at(0).eventTime,timez)).format(dateFormat);
    }
    return `Shabbat in ${cityName} starts this week at ${candleLighting} and ends at ${havdalahTime}`;
}

function shabLocation(prompt,chat) {
    const name = prompt.split(/\s+/).slice(1).join(' ').trim().toLowerCase();
    for (const item in citiesObj.data) {
        if (name == citiesObj.data[item]["ASCII Name"].toLowerCase()) {
        chat.location = citiesObj.data[item]["ASCII Name"];
        return `chat location was set to ${chat.location}`;
        }
    }
    return "that location isn't in our database";
}

function convertToTimeZone(inputDateString, targetTimeZone) {
    const inputDate = new Date(inputDateString);
  
    // Create a new Date object with the same timestamp but in the target time zone
    const targetDate = new Date(inputDate.toLocaleString('en-US', { timeZone: targetTimeZone }));
  
    return targetDate;
}
  

module.exports = { shabTimes, shabLocation };


// tests:

function testShabbot() {
    const date = new Date('2023-10-20T00:00:00Z');
    const location = 'johannesburg';
    const prompt = '!shabtimes';
    const attendee = { number: '12063214745' };
    const chat = {location: "new york"};

    console.log(shabTimes(date, location, prompt, attendee));
    console.log(shabLocation("!shablocation johannesburg", chat));
}

//testShabbot()