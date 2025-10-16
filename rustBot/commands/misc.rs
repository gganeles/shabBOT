use crate::db::get_chat_location;
use diesel::sqlite::SqliteConnection;

pub fn count_cmd() -> String {
    let link = "https://www.chabad.org/holidays/sefirah/omer-count_cdo/jewish/Count-the-Omer.htm";
    format!(
        "you forgot that your suppose to count the omer!? ha! no, not tonight\n{}",
        link
    )
}

pub fn needs_cmd(_prompt: &str) -> String {
    "The needs command requires the events system which is being deprecated.\n\
     Please use !quickshab instead to track meal items."
        .to_string()
}

pub fn fast_cmd(db: &mut SqliteConnection, chat_id: &str) -> String {
    let location = get_chat_location(db, chat_id).unwrap_or_else(|_| "Haifa".to_string());
    format!(
        "Fast day checking for {}\n\n\
         To implement this properly, you need to:\n\
         1. Integrate with a Hebrew calendar library (e.g., hebcal.com API)\n\
         2. Get current Hebrew date\n\
         3. Check if current/upcoming date is a fast day\n\
         4. Calculate fast start/end times based on location",
        location
    )
}

pub fn start_cmd(prompt: &str) -> String {
    format!("Start: {}", prompt)
}

pub fn end_cmd() -> String {
    "Goodbye".to_string()
}
