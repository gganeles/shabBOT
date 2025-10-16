// Test for clean_reminder_text utility function
// Run with: cargo test --test test_clean_reminder_text
// Tests match the Go implementation behavior exactly

use rustWABot::utils::clean_reminder_text;

#[test]
fn test_clean_basic() {
    assert_eq!(clean_reminder_text("buy milk"), "buy milk");
}

#[test]
fn test_clean_me_to_prefix() {
    assert_eq!(clean_reminder_text("me to buy milk"), "buy milk");
}

#[test]
fn test_clean_me_prefix() {
    assert_eq!(clean_reminder_text("me buy milk"), "buy milk");
}

#[test]
fn test_clean_in_suffix() {
    assert_eq!(clean_reminder_text("buy milk in"), "buy milk");
}

#[test]
fn test_clean_at_suffix() {
    assert_eq!(clean_reminder_text("buy milk at"), "buy milk");
}

#[test]
fn test_clean_whitespace() {
    assert_eq!(clean_reminder_text("  buy milk  "), "buy milk");
}

#[test]
fn test_clean_empty() {
    // Go version returns empty string, not "nothing"
    assert_eq!(clean_reminder_text(""), "");
}

#[test]
fn test_clean_only_prefix() {
    // Go version: "me to " trims to "me to", doesn't match "me to " prefix,
    // then matches "me " and removes it, leaving "to"
    assert_eq!(clean_reminder_text("me to "), "to");
}

#[test]
fn test_clean_only_me_prefix() {
    // Go version: "me " trims to "me", doesn't match "me " prefix (with space)
    // so it returns "me" unchanged
    assert_eq!(clean_reminder_text("me "), "me");
}

#[test]
fn test_clean_only_spaces() {
    // Go version returns empty string
    assert_eq!(clean_reminder_text("   "), "");
}

#[test]
fn test_clean_combined() {
    assert_eq!(clean_reminder_text("me to buy milk at"), "buy milk");
}

#[test]
fn test_clean_case_sensitive() {
    // Go version uses case-sensitive comparison (strings.TrimPrefix/TrimSuffix)
    // So uppercase prefixes/suffixes are NOT removed
    assert_eq!(clean_reminder_text("ME TO buy milk"), "ME TO buy milk");
    assert_eq!(clean_reminder_text("buy milk AT"), "buy milk AT");
}
