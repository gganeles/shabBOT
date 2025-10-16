// Example: How to integrate chrono date parsing into a command
// Add this to router.rs

// Example command that uses natural time parsing
pub fn example_remind_cmd(prompt: &str) -> String {
    // Parse the natural language time expression
    match parse_natural_time(prompt, None) {
        Ok(results) if !results.is_empty() => {
            let mut response = String::from("Parsed dates:\n");
            for parsed in results {
                response.push_str(&format!(
                    "  • '{}' → {}\n",
                    parsed.text,
                    parsed.time.as_ref().unwrap_or(&"unknown".to_string())
                ));
            }
            response
        }
        Ok(_) => "❌ No date/time found in your message".to_string(),
        Err(e) => format!("❌ Error: {}", e),
    }
}

// To use in the router's match statement:
//
// match cmd.as_str() {
//     "testdate" => example_remind_cmd(prompt_trimmed),
//     // ... other commands
// }
