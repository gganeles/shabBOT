const chrono = require('chrono-node');

function parseDate(expression, date) {
    return chrono.parseDate(expression, new Date(date))
}

function parseRange(expression, date) {
    return chrono.parse(expression, new Date(date)).map(res => {
        let result = {};

        if (res.start) result.start = res.start.date();
        if (res.end) result.end = res.end.date();

        result.date = res.date();

        return result;
    });
}

function parse(expression, date) {
    const results = chrono.parse(expression, new Date(date), { forwardDate: true });
    return results.map(res => {
        const parsedDate = res.start ? res.start.date() : res.date();
        return {
            Text: res.text,
            Index: res.index,
            Time: parsedDate ? parsedDate.toISOString() : null
        };
    });
}

module.exports = {
    parseRange, parseDate, parse
};