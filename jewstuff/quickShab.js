

const typeList = [
  "main",
  "side",
  "plastics",
  "drinks",
  "wine",
  "challah",
  "dips",
  "dessert"
];


function needsNumber(n) {
  const preset = {
      'main': Math.ceil((n+1)/4),
      'side': Math.ceil((n)/6),
      'plastics': n>5 ? Math.ceil(n/20) : 0,
      'drinks': n>5 ? Math.ceil(n/12) : 0,
      'wine': Math.ceil(n/8),
      'challah': Math.ceil(n/10),
      'dips': Math.ceil((n+1)/12),
      'dessert': Math.ceil((n-5)/8)
  }
  return preset;
}

function quickshab(eventsList, num_of_guests) {
  const newEvent = {number:num_of_guests};
  eventsList.push(newEvent);
  return formatString(newEvent);
}

function updateNumber(event, num_of_guests) { 
  event.number = num_of_guests;
  return formatString(event);
}

function show(event) {
  return formatString(event) + "\n\nExample Usage:\n   • !bring main\n   • !assign gabe main\n\nYou can also use !unbring and !unassign";
}

function formatString(data) {
  let result = [];
  needs = needsNumber(data.number);
  for (const category in needs) {
    upperLim = Math.max(needs[category],data[category]?data[category].length:0);
    for (let i=0; i<upperLim; i++) {
      result.push(`${capitalizeFirst(category)}: ${data[category]&&data[category].length>i?data[category][i]:""}`);
    }
  }

  Object.keys(data).filter(x=>!typeList.includes(x)&&x!="number").forEach((category) => {
  console.log("This category is not in the typeList: ", category);
    if (data[category] && data[category].length > 0) {
      for (let i = 0; i < data[category].length; i++) {
        result.push(`${capitalizeFirst(category)}: ${data[category]}`);
      }
    }
  });

  return result.join("\n");
}



const capitalizeFirst = str => str.charAt(0).toUpperCase() + str.slice(1);


function addToCategory(data, category, name) {
  if (!data[category]) {
      data[category] = [];
  }

  data[category].push(name);
}

const removeName = (data, name) => {
  for (const key in data) {
      if (key !== "Number" && Array.isArray(data[key])) {
          const index = data[key].indexOf(name);
          if (index !== -1) {
            if (data[key].length==1&&!typeList.includes(key)){
              return void delete data[key];
            } else {
              return void data[key].splice(index, 1);
            }
          }
        }
  }
};


function bringCmd(event, msg, name) {
  const category = msg.split(/\s/)[1].toLowerCase()
  addToCategory(event, category, name);
  return formatString(event);
}

function assignCmd(event, msg) {
  const category = msg.split(/\s/)[2].toLowerCase()
  const name = msg.split(/\s/)[1]
  addToCategory(event, category, name);
  return formatString(event);
}

function unbringCmd(event, name) {
  removeName(event, name);
  return formatString(event);
}

function unassignCmd(event, name) {
  removeName(event, name);
  return formatString(event);
}


module.exports = {show, updateNumber, quickshab, bringCmd, assignCmd, unbringCmd, unassignCmd};


if (false) {
  testList = [];
  console.log(quickshab(testList, 10));
  console.log("\n");
  console.log(bringCmd(testList[0], "!bring chair", "Gabe"));
  console.log("\n");
  console.log(unassignCmd(testList[0], "Gabe"));
  console.log("\n");
  console.log(show(testList[0]));
}