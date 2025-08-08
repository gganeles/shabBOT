const fetch = require('node-fetch');


function isComing(phoneNumber, event) {
    for (let p = 0; p < event.attendance.length; p++) {
        if (event.attendance.at(p).number === phoneNumber) {
            return true;
        }
    }
    return false
}

function countAst(str) {
    let count = 0
    for (let i = 0; i < str.length; i++) {
        if (str.at(i) === '*') {
            count += 1;
        }
    }
    return count;
} 

function find_event_index_by_name(string, eventList) {
    for (let i = 0; i < eventList.length; i++) {
        if (string.match(new RegExp('\\b' + eventList.at(i).eventName + '\\b', 'i'))) {
            return i
        }
    }
    return -1
}

function findIndicies(string, eventList) {
    const eventByName = find_event_index_by_name(string, eventList)
    if (eventByName != -1) {
        return [eventByName]
    } else {
        let indicies = []
        if (!string.match(/\b\d\b/g)) {return []}
        string.match(/\b\d\b/g).forEach(item => indicies.push(parseInt(item)-1))
        return indicies
    }
}

function removeCmdandIndicies(string, eventList) {
    for (let i = 0; i < eventList.length; i++) {
        string = string.replace(new RegExp('\\s' + eventList.at(i).eventName + '\\b', 'i'),'')
    }
    return string.split(/\s*,\s*|\s/).slice(1,).filter(item => isNaN(parseInt(item))).join(' ').trim()
}

function sumGuests(event) {
    let summ = 0;
    for (let i=0; i<event.attendance.length; i++) {
        summ += event.attendance.at(i).guests;
    }
    return summ;
}

function newestIndex(eventList) {
    for (let i = 0; i<eventList.length; i++) {
        if (eventList.at(i).isNewest) {
            return i
        }
    }
    return -1
}

async function itemTyper(item) {
    item = item.replace(/dips/gi,'salatim')
	const foodtypes = item.match(/(drinks|wine|challah|salatim|plastics)/gi)
	if (foodtypes) {
        return foodtypes;
    } else {
        const type = await GPTparse(item)
        const typeList = type.match(/(main_dish|side_dish|dessert)/gi)
        if (typeList) {
            return typeList
        } else {
            return ["other"]
        }
	}
}

async function GPTparse(item) {
    try {
	const apiUrl = 'https://api.openai.com/v1/chat/completions'
	//const dishList = 'main_dish, side_dish, disposables, drinks, wine, challah, dips'
	const gptPrompt = `I input phrase you output type: main_dish, side_dish, or dessert. ONLY respond with type. if you are VERY SURE, you MAY respond with multiple types. The first prompt is "${item}"`
	const request = {
		model:'gpt-3.5-turbo',
		messages:[
			{role:'user',content: gptPrompt}
		],
		temperature: 0,
		max_tokens: 8,
	}
	const headers = {
			'Content-Type': 'application/json',
			'Authorization': `Bearer ${process.env.OPENAI_KEY}`,
		}

	const response = await fetch(apiUrl, {
		method: 'POST',
		headers: headers,
		body: JSON.stringify(request)
	});
		//.then(response => response.json()).then(item=> console.log(item.choices.at(0).message.content))
	
	const result = await response.json();
	return await result.choices.at(0).message.content
    }catch(err){}
}



function convertListToObj(list) {
    const result = {};
    
    for (let i = 0; i < list.length; i++) {
      const { chatID } = list[i];
      result[chatID] = list[i];
      delete list[i].chatID;
    }
    
    return result;
}

module.exports = {findIndicies, newestIndex, itemTyper, countAst, isComing, sumGuests, removeCmdandIndicies, convertListToObj};
