use crate::router::parse_natural_time;
use chrono::{DateTime, Utc};
use regex::Regex;
use serde::Deserialize;
use std::fs::File;
use std::io::Read;
use std::sync::OnceLock;

#[derive(Debug, Clone, Deserialize)]
pub struct GeoName {
    pub geoname_id: String,
    pub name: String,
    pub timezone: String,
    pub coordinates: String,
    pub country_name: String,
    pub country_code: String,
}

/// Search object for whatKeys function
#[derive(Clone)]
pub struct SearchObject {
    pub id: String,
    pub message: String,
    pub index: usize,
}

// Global lazy-loaded geonames list
static GEONAMES_LIST: OnceLock<Vec<GeoName>> = OnceLock::new();

/// Initialize and load the geonames list from CSV
fn init_geonames() -> Vec<GeoName> {
    // Try multiple possible paths for the CSV file
    let possible_paths = vec![
        "commands/geoNamesList.csv",
        "rustBot/commands/geoNamesList.csv",
        "../commands/geoNamesList.csv",
        "geoNamesList.csv",
    ];

    for path in &possible_paths {
        match load_geonames_from_file(path) {
            Ok(geonames) => {
                log::info!("Loaded {} geonames from {}", geonames.len(), path);
                return geonames;
            }
            Err(e) => {
                log::debug!("Could not load geonames from {}: {}", path, e);
            }
        }
    }

    log::error!(
        "Could not load geoNamesList.csv from any known location. Tried: {:?}",
        possible_paths
    );
    Vec::new()
}

/// Load geonames from a specific CSV file path
fn load_geonames_from_file(file_path: &str) -> Result<Vec<GeoName>, Box<dyn std::error::Error>> {
    let mut file = File::open(file_path)?;
    let mut contents = String::new();
    file.read_to_string(&mut contents)?;

    let mut geonames = Vec::new();
    let mut rdr = csv::ReaderBuilder::new()
        .delimiter(b';')
        .has_headers(true)
        .flexible(true)
        .from_reader(contents.as_bytes());

    for result in rdr.records() {
        let record = result?;
        if record.len() < 20 {
            continue;
        }

        // Column 0 is "Geoname ID", Column 2 is "ASCII Name", column 7 is "Country name EN",
        // column 8 is "Country code", column 16 is "Timezone", column 19 is "Coordinates"
        let geoname_id = record.get(0).unwrap_or("").trim().to_string();
        let name = record.get(2).unwrap_or("").trim().to_string();
        let country_name = record.get(7).unwrap_or("").trim().to_string();
        let country_code = record.get(8).unwrap_or("").trim().to_string();
        let timezone = record.get(16).unwrap_or("").trim().to_string();
        let coordinates = record.get(19).unwrap_or("").trim().to_string();

        if !geoname_id.is_empty()
            && !name.is_empty()
            && !timezone.is_empty()
            && !coordinates.is_empty()
        {
            geonames.push(GeoName {
                geoname_id,
                name,
                timezone,
                coordinates,
                country_name,
                country_code,
            });
        }
    }

    Ok(geonames)
}

/// Parse command arguments from a string
pub fn parse_args(prompt: &str) -> Vec<String> {
    prompt.split_whitespace().map(|s| s.to_string()).collect()
}

/// Parse integer with default value
pub fn parse_int(s: &str, default: i32) -> i32 {
    s.parse::<i32>().unwrap_or(default)
}

/// Find keys in items that match the keyword
pub fn what_keys(keyword: &str, items: &[SearchObject]) -> Vec<String> {
    let mut keys = Vec::new();
    let keyword_lower = keyword.to_lowercase();

    // First try direct message matching
    for item in items {
        if item.message.to_lowercase().contains(&keyword_lower) {
            keys.push(item.id.clone());
        }
    }

    // If no matches, try extracting a number and using it as an index
    if keys.is_empty() {
        let re = Regex::new(r"\b\d+\b").unwrap();
        if let Some(mat) = re.find(&keyword_lower) {
            if let Ok(index) = mat.as_str().parse::<usize>() {
                if index < items.len() {
                    keys.push(items[index].id.clone());
                }
            }
        }
    }

    keys
}

/// Find a geo location from a name
pub fn find_geo_location(location: &str) -> Option<GeoName> {
    if location.is_empty() {
        return None;
    }

    // Get or initialize the geonames list
    let geonames = GEONAMES_LIST.get_or_init(init_geonames);

    let location_lower = location.to_lowercase();

    // Search for exact match (case-insensitive)
    for geo in geonames {
        if geo.name.to_lowercase() == location_lower {
            return Some(geo.clone());
        }
    }

    None
}

/// Clean reminder text by removing common filler words and phrases
///
/// This function removes common prefixes and suffixes that users often include
/// in reminder text but aren't part of the actual reminder message.
/// Matches the Go implementation logic exactly (case-sensitive, single pass).
///
/// # Arguments
/// * `text` - The text to clean
///
/// # Returns
/// * Cleaned text (trimmed)
///
/// # Examples
/// ```
/// use rustWABot::utils::clean_reminder_text;
///
/// assert_eq!(clean_reminder_text("me to buy milk"), "buy milk");
/// assert_eq!(clean_reminder_text("me "), "");
/// assert_eq!(clean_reminder_text("  "), "");
/// ```
pub fn clean_reminder_text(text: &str) -> String {
    let mut cleaned = text.trim().to_string();

    // Remove common prefixes - single pass only (matches Go implementation)
    // Case-sensitive comparison
    if cleaned.starts_with("me to ") {
        cleaned = cleaned[6..].to_string();
    } else if cleaned.starts_with("me ") {
        cleaned = cleaned[3..].to_string();
    }

    // Remove common suffixes - single pass only (matches Go implementation)
    // Case-sensitive comparison
    if cleaned.ends_with(" in") {
        cleaned = cleaned[..cleaned.len() - 3].to_string();
    } else if cleaned.ends_with(" at") {
        cleaned = cleaned[..cleaned.len() - 3].to_string();
    }

    cleaned.trim().to_string()
}

/// Extract datetime from text and return (remaining_text, datetime)
///
/// This function parses natural language datetime expressions from text
/// and returns the text with the datetime portion removed, along with the parsed DateTime.
/// Matches the Go implementation logic exactly with timezone support.
///
/// # Arguments
/// * `text` - The full text containing a datetime expression
/// * `timezone_name` - IANA timezone name (e.g., "Asia/Jerusalem", "America/New_York")
///
/// # Returns
/// * `Ok((remaining_text, datetime))` - The text without the datetime part and the parsed DateTime
/// * `Err(String)` - Error message if parsing fails or no datetime found
///
/// # Examples
/// ```no_run
/// use rustWABot::utils::extract_datetime_from_text;
///
/// let (remaining, dt) = extract_datetime_from_text("remind me tomorrow at 3pm to buy milk", "Asia/Jerusalem")?;
/// // remaining = "buy milk"
/// // dt = DateTime representing tomorrow at 3pm in the specified timezone
/// # Ok::<(), String>(())
/// ```
pub fn extract_datetime_from_text(
    text: &str,
    timezone_name: &str,
) -> Result<(String, DateTime<Utc>), String> {
    use chrono_tz::Tz;

    // Parse the timezone
    let tz: Tz = timezone_name
        .parse()
        .map_err(|_| format!("Invalid timezone: {}", timezone_name))?;

    // Get current time in the target timezone as reference
    let now_in_tz = chrono::Utc::now().with_timezone(&tz);
    let reference_date = Some(now_in_tz.to_rfc3339());

    // Parse using chrono-node with timezone-aware reference date
    let parsed_dates = parse_natural_time(text, reference_date.as_deref())?;

    if parsed_dates.is_empty() {
        return Err("No datetime found in text".to_string());
    }

    // Get the first parsed date
    let parsed = &parsed_dates[0];

    // Extract the datetime
    let time_str = parsed
        .time
        .as_ref()
        .ok_or_else(|| "No time information in parsed result".to_string())?;

    // Parse ISO 8601 datetime string and convert to UTC
    let datetime = DateTime::parse_from_rfc3339(time_str)
        .map_err(|e| format!("Failed to parse datetime: {}", e))?
        .with_timezone(&Utc);

    // Split the text by the datetime text (matches Go logic)
    let datetime_text = &parsed.text;
    let parts: Vec<&str> = text.split(datetime_text).collect();

    // Join the parts back together with a space and trim each part
    let remaining_text = parts
        .iter()
        .map(|s| s.trim())
        .filter(|s| !s.is_empty())
        .collect::<Vec<&str>>()
        .join(" ");

    // Clean the remaining text using the same logic as Go
    let cleaned_text = clean_reminder_text(&remaining_text);

    Ok((cleaned_text, datetime))
}
