use log::{error, info, warn};
use std::process::Command;
use std::sync::Arc;
use whatsapp_rust::store::sqlite_store::SqliteStore;

/// Clean up corrupted sessions and identities for a specific user address
/// This is necessary when encountering "UntrustedIdentity" errors which indicate
/// that a contact has reinstalled WhatsApp or changed devices
///
/// This implementation uses the sqlite3 command-line tool to avoid Rust dependency conflicts
pub async fn cleanup_user_sessions(
    _backend: Arc<SqliteStore>,
    address: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    info!(
        "🧹 Attempting to clean up sessions and identities for user {}",
        address
    );

    // Use sqlite3 command to delete sessions
    let sessions_output = Command::new("sqlite3")
        .arg("db/whatsapp.db")
        .arg(format!(
            "DELETE FROM sessions WHERE address LIKE '{}%'; SELECT changes();",
            address
        ))
        .output()?;

    let sessions_deleted = if sessions_output.status.success() {
        String::from_utf8_lossy(&sessions_output.stdout)
            .trim()
            .parse::<i32>()
            .unwrap_or(0)
    } else {
        error!(
            "Failed to delete sessions: {}",
            String::from_utf8_lossy(&sessions_output.stderr)
        );
        0
    };

    // Use sqlite3 command to delete identities
    let identities_output = Command::new("sqlite3")
        .arg("db/whatsapp.db")
        .arg(format!(
            "DELETE FROM identities WHERE address LIKE '{}%'; SELECT changes();",
            address
        ))
        .output()?;

    let identities_deleted = if identities_output.status.success() {
        String::from_utf8_lossy(&identities_output.stdout)
            .trim()
            .parse::<i32>()
            .unwrap_or(0)
    } else {
        error!(
            "Failed to delete identities: {}",
            String::from_utf8_lossy(&identities_output.stderr)
        );
        0
    };

    if sessions_deleted > 0 || identities_deleted > 0 {
        info!(
            "✅ Cleaned up {} session(s) and {} identity/identities for user {}",
            sessions_deleted, identities_deleted, address
        );
        info!("   The library will automatically re-establish sessions with the new identity.");
    } else {
        warn!(
            "⚠️  No sessions or identities found for user {} to clean up",
            address
        );
    }

    Ok(())
}

/// Extract the phone number from a JID string
/// Examples: "972587120601.0" -> "972587120601"
///          "972587120601" -> "972587120601"
///          "972587120601:54@s.whatsapp.net" -> "972587120601"
///          "972587120601@s.whatsapp.net" -> "972587120601"
pub fn extract_phone_number(jid: &str) -> String {
    // First split on @ to remove the server part
    let user_part = jid.split('@').next().unwrap_or(jid);
    // Then split on : to remove the device ID
    let phone_part = user_part.split(':').next().unwrap_or(user_part);
    // Then split on . to handle old format
    phone_part
        .split('.')
        .next()
        .unwrap_or(phone_part)
        .to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_extract_phone_number() {
        assert_eq!(extract_phone_number("972587120601.0"), "972587120601");
        assert_eq!(extract_phone_number("972587120601"), "972587120601");
        assert_eq!(extract_phone_number("1234567890.54"), "1234567890");
        assert_eq!(
            extract_phone_number("972587120601:54@s.whatsapp.net"),
            "972587120601"
        );
        assert_eq!(
            extract_phone_number("972587120601@s.whatsapp.net"),
            "972587120601"
        );
        assert_eq!(
            extract_phone_number("1234567890:0@s.whatsapp.net"),
            "1234567890"
        );
    }
}
