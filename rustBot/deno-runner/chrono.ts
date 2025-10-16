// Deno-compatible chrono parser
// Using npm:chrono-node which works with Deno's npm specifiers

import * as chrono from "npm:chrono-node@2.7.6";

export function parse(expression: string, referenceDate: string | null): any[] {
    const refDate = referenceDate ? new Date(referenceDate) : new Date();
    const results = chrono.parse(expression, refDate, { forwardDate: true });

    if (results && results.length > 0) {
        return results.map((res: any) => {
            const parsedDate = res.start ? res.start.date() : res.date();
            return {
                Text: res.text,
                Index: res.index,
                Time: parsedDate ? parsedDate.toISOString() : null,
            };
        });
    }
    return [];
}

// For direct CLI testing
if (import.meta.main) {
    const expression = Deno.args[0] || "tomorrow at 3pm";
    const referenceDate = Deno.args[1] || null;

    console.log("Parsing:", expression);
    const results = parse(expression, referenceDate);
    console.log(JSON.stringify(results, null, 2));
}
