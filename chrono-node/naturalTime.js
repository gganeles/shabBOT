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
    } else {
        return [];
    }
}

// CLI interface
if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args.length < 2) {
        console.error('Usage: bun naturalTime.js <expression> <baseDate>');
        process.exit(1);
    }
    
    const [expression, baseDate] = args;
    
    try {
        const result = parse(expression, baseDate);
        console.log(JSON.stringify(result));
    } catch (error) {
        console.error('Error:', error.message);
        process.exit(1);
    }
}

module.exports = { parse };
