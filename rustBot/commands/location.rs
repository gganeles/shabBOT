use crate::db::*;
use crate::utils::*;
use diesel::prelude::*;

pub fn shablocation_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
    let args = parse_args(prompt);

    if args.len() < 2 {
        let location = get_chat_location(db, chat_id).unwrap_or_else(|_| "Haifa".to_string());
        return format!("Current location: {}", location);
    }

    let location = args[1..].join(" ");
    let location = location.trim();

    if location.is_empty() {
        return "Please provide a valid location".to_string();
    }

    let geo = find_geo_location(location);
    if let Some(geo_data) = geo {
        if let Err(e) = set_chat_location(db, chat_id, &geo_data.name, &geo_data.timezone) {
            return format!("Error setting location: {}", e);
        }
        format!("Chat location was set to {}", geo_data.name)
    } else {
        format!(
            "Location '{}' not found in database. Try a different location name.",
            location
        )
    }
}

pub fn shabtimes_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
    // Get chat's default location
    let mut location = get_chat_location(db, chat_id).unwrap_or_else(|_| "Haifa".to_string());

    // Parse prompt for alternate location
    let args = parse_args(prompt);
    if args.len() > 1 {
        let potential_location = args[1..].join(" ");
        if !potential_location.is_empty() {
            location = potential_location;
        }
    }

    // Find location data from geoNamesList
    let geo_data = match find_geo_location(&location) {
        Some(data) => data,
        None => {
            return format!(
                "Location '{}' not found in database. Try !shablocation to set a valid location.",
                location
            );
        }
    };

    // Calculate candle lighting offset based on location
    let offset_minutes = match geo_data.name.as_str() {
        "Jerusalem" => 40,
        "Haifa" | "Zichron Yaakov" => 30,
        _ => 18,
    };

    // Use Hebcal API with geoname ID
    let api_url = format!(
        "https://www.hebcal.com/shabbat?cfg=json&geonameid={}&m=on&b={}",
        geo_data.geoname_id, offset_minutes
    );

    log::info!("Fetching Shabbat times from: {}", api_url);

    match ureq::get(&api_url).call() {
        Ok(response) => {
            log::info!("API response status: {}", response.status());

            match response.into_string() {
                Ok(text) => {
                    log::debug!("API response: {}", text);

                    match serde_json::from_str::<serde_json::Value>(&text) {
                        Ok(result) => {
                            if let Some(items) = result["items"].as_array() {
                                log::info!("Found {} items in response", items.len());

                                let mut candle_lighting = None;
                                let mut havdalah = None;

                                for item in items {
                                    if let Some(category) = item["category"].as_str() {
                                        log::debug!(
                                            "Item category: {}, title: {:?}",
                                            category,
                                            item["title"]
                                        );

                                        if category == "candles" {
                                            candle_lighting = item["title"]
                                                .as_str()
                                                .and_then(|s| s.split("Candle lighting: ").nth(1));
                                        } else if category == "havdalah" {
                                            havdalah = item["title"].as_str().and_then(|s| {
                                                // Handle "Havdalah (X min): HH:MM" format
                                                if let Some(after_colon) = s.split(": ").nth(1) {
                                                    Some(after_colon)
                                                } else {
                                                    s.split("Havdalah: ").nth(1)
                                                }
                                            });
                                        }
                                    }
                                }

                                if let (Some(candles), Some(havd)) = (candle_lighting, havdalah) {
                                    log::info!(
                                        "Successfully parsed times - Candles: {}, Havdalah: {}",
                                        candles,
                                        havd
                                    );

                                    return format!(
                                        "Shabbat in {} starts this week at {} and ends at {}",
                                        geo_data.name, candles, havd
                                    );
                                } else {
                                    log::warn!(
                                        "Could not find candle lighting or havdalah in response"
                                    );
                                }
                            } else {
                                log::warn!("No 'items' array in response");
                            }
                        }
                        Err(e) => {
                            log::error!("Failed to parse JSON: {}", e);
                        }
                    }
                }
                Err(e) => {
                    log::error!("Failed to read response text: {}", e);
                }
            }
        }
        Err(e) => {
            log::error!("Error fetching Shabbat times: {}", e);
        }
    }

    // Fallback if API fails
    format!(
        "Shabbat times for {} - Candle lighting approximately {} minutes before sunset on Friday",
        geo_data.name, offset_minutes
    )
}

pub fn chagtimes_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
    // Get chat's default location
    let mut location = get_chat_location(db, chat_id).unwrap_or_else(|_| "Haifa".to_string());

    // Parse prompt for alternate location
    let args = parse_args(prompt);
    if args.len() > 1 {
        let potential_location = args[1..].join(" ");
        if !potential_location.is_empty() {
            location = potential_location;
        }
    }

    // Find location data from geoNamesList
    let geo_data = match find_geo_location(&location) {
        Some(data) => data,
        None => {
            return format!(
                "Location '{}' not found in database. Try !shablocation to set a valid location.",
                location
            );
        }
    };

    // Call Hebcal REST API with geoname ID
    let api_url = format!(
        "https://www.hebcal.com/shabbat?cfg=json&geonameid={}&M=on",
        geo_data.geoname_id
    );

    match ureq::get(&api_url).call() {
        Ok(response) => {
            if let Ok(text) = response.into_string() {
                if let Ok(result) = serde_json::from_str::<serde_json::Value>(&text) {
                    if let Some(items) = result["items"].as_array() {
                        // Find the next major holiday (excluding Shabbat)
                        for item in items {
                            let title = item["title"].as_str().unwrap_or("");
                            let category = item["category"].as_str().unwrap_or("");
                            let date = item["date"].as_str().unwrap_or("");
                            let hebrew = item["hebrew"].as_str().unwrap_or("");

                            // Skip Shabbat and candle lighting
                            if category.contains("candles")
                                || category.contains("havdalah")
                                || title == "Parashat"
                                || title.starts_with("Parashat ")
                            {
                                continue;
                            }

                            // Parse item date
                            if let Ok(item_date) =
                                chrono::NaiveDate::parse_from_str(date, "%Y-%m-%d")
                            {
                                let today = chrono::Utc::now().date_naive();

                                // Check if this is in the future or today
                                if item_date >= today {
                                    let mut response = format!("Next Holiday: {} 🕎\n", title);

                                    if !hebrew.is_empty() {
                                        response.push_str(&format!("Hebrew: {}\n", hebrew));
                                    }

                                    response.push_str(&format!("Date: {}\n", date));

                                    // Calculate days until
                                    let days_until = (item_date - today).num_days();
                                    match days_until {
                                        0 => response.push_str("That's today!\n"),
                                        1 => response.push_str("That's tomorrow!\n"),
                                        _ => response
                                            .push_str(&format!("That's in {} days\n", days_until)),
                                    }

                                    // Note about candle lighting for major holidays
                                    if category.contains("holiday")
                                        && !title.to_lowercase().contains("purim")
                                        && !title.to_lowercase().contains("tu b")
                                    {
                                        let offset = match geo_data.name.as_str() {
                                            "Jerusalem" => 40,
                                            "Haifa" | "Zichron Yaakov" => 30,
                                            _ => 18,
                                        };
                                        response.push_str(&format!("\n(Candle lighting approximately {} minutes before sunset)", offset));
                                    }

                                    return response;
                                }
                            }
                        }
                    }
                }
            }
        }
        Err(e) => {
            log::error!("Error fetching holiday data: {}", e);
            return format!("Error fetching holiday data. Please try again later.");
        }
    }

    "No upcoming holidays found in the next 2 months.".to_string()
}
