// Integration test for extract_datetime_from_text utility function
// Run with: cargo test --test test_datetime_extract

use rustWABot::utils::extract_datetime_from_text;

#[test]
fn test_simple_future_time() {
    let input = "remind me tomorrow at 3pm to buy milk";
    match extract_datetime_from_text(input, None) {
        Ok((remaining, _datetime)) => {
            println!("Remaining text: \"{}\"", remaining);
            assert!(remaining.contains("buy milk") || remaining.contains("remind me"));
            assert!(!remaining.contains("tomorrow at 3pm"));
        }
        Err(e) => panic!("Failed to parse: {}", e),
    }
}

#[test]
fn test_relative_time() {
    let input = "send this message in 2 hours";
    match extract_datetime_from_text(input, None) {
        Ok((remaining, _datetime)) => {
            println!("Remaining text: \"{}\"", remaining);
            assert!(remaining.contains("send") || remaining.contains("message"));
            assert!(!remaining.contains("in 2 hours"));
        }
        Err(e) => panic!("Failed to parse: {}", e),
    }
}

#[test]
fn test_specific_date() {
    let input = "schedule meeting next Friday at 2pm";
    match extract_datetime_from_text(input, None) {
        Ok((remaining, _datetime)) => {
            println!("Remaining text: \"{}\"", remaining);
            assert!(remaining.contains("meeting") || remaining.contains("schedule"));
            assert!(!remaining.contains("next Friday at 2pm"));
        }
        Err(e) => panic!("Failed to parse: {}", e),
    }
}

#[test]
fn test_no_datetime() {
    let input = "just a regular reminder";
    match extract_datetime_from_text(input, None) {
        Ok(_) => panic!("Should have failed - no datetime in input"),
        Err(e) => {
            println!("Expected error: {}", e);
            assert!(e.contains("No datetime found"));
        }
    }
}

#[test]
fn test_multiple_words_after_datetime() {
    let input = "meet with John tomorrow at 10am to discuss the project";
    match extract_datetime_from_text(input, None) {
        Ok((remaining, _datetime)) => {
            println!("Remaining text: \"{}\"", remaining);
            assert!(remaining.contains("John") || remaining.contains("project"));
            assert!(!remaining.contains("tomorrow at 10am"));
        }
        Err(e) => panic!("Failed to parse: {}", e),
    }
}
