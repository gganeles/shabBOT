const chrono = require('chrono-node');
const msRecognizers = require('@microsoft/recognizers-text-date-time/dist/recognizers-text-date-time.umd.js');

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
    const x = new Date(date);
    // Use literal values instead of Culture/DateTimeOptions to avoid bundling issues
    // Culture.English = "en-us", DateTimeOptions.None = 0
    const recognizer = msRecognizers.recognizeDateTime(expression, "en-us", 0, x);
    const results = chrono.parse(expression, x, { forwardDate: true });
    if (results && results.length > 0) {
        return results.map(res => {
            const parsedDate = res.start ? res.start.date() : res.date();
            return {
                Text: res.text,
                Index: res.index,
                Time: parsedDate ? parsedDate.toISOString() : null,
                MicrosoftResults: JSON.stringify(recognizer)
            };
        });
    } else if (recognizer && recognizer.length > 0) {
        return recognizer.map(res => {
            let parsedDate = null;

            if (res.resolution && res.resolution.values && res.resolution.values.length > 0) {
                // Priority: datetime > time > date
                // Find the best resolution value based on type priority
                let selectedValue = null;

                // First, look for datetime
                selectedValue = res.resolution.values.find(v => v.type === 'datetime');

                // If no datetime, look for time
                if (!selectedValue) {
                    selectedValue = res.resolution.values.find(v => v.type === 'time');
                }

                // If no time, look for date
                if (!selectedValue) {
                    selectedValue = res.resolution.values.find(v => v.type === 'date');
                }

                // Fallback to first value if none matched
                if (!selectedValue) {
                    selectedValue = res.resolution.values[0];
                }

                const type = selectedValue.type;

                if (type === 'datetime') {
                    // Full datetime
                    parsedDate = new Date(selectedValue.value);
                } else if (type === 'time') {
                    // Time only - use current date
                    const timeStr = selectedValue.value;
                    const [hours, minutes, seconds] = timeStr.split(':').map(Number);
                    parsedDate = new Date(x);
                    parsedDate.setHours(hours, minutes || 0, seconds || 0, 0);
                } else if (type === 'date') {
                    // Date only - use current time
                    const dateStr = selectedValue.value;
                    parsedDate = new Date(dateStr);
                    parsedDate.setHours(x.getHours(), x.getMinutes(), x.getSeconds(), x.getMilliseconds());
                }
            }

            return {
                Text: res.text,
                Index: res.start,
                Time: parsedDate ? parsedDate.toISOString() : null,
                MicrosoftResults: JSON.stringify(recognizer)
            };
        });
    } else {
        return [];
    }


}

module.exports = {
    parseRange, parseDate, parse
};