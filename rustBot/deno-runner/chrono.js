const chrono = require('chrono-node');

function parse(expression, date) {
    const x = new Date(date);
    const results = chrono.parse(expression, x, { forwardDate: true });
    if (results && results.length > 0) {
        return results.map(res => {
            const parsedDate = res.start ? res.start.date() : res.date();
            return {
                Text: res.text,
                Index: res.index,
                Time: parsedDate ? parsedDate.toISOString() : null,
            };
        });
    }
    else return [];
}