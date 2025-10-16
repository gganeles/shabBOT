// Test the Hebcal API parsing logic

use serde_json;

fn main() {
    let json_str = r#"{"title":"Hebcal","items":[{"title":"Candle lighting: 17:46","category":"candles"},{"title":"Havdalah (50 min): 18:53","category":"havdalah"}]}"#;

    let result: serde_json::Value = serde_json::from_str(json_str).unwrap();

    if let Some(items) = result["items"].as_array() {
        println!("Found {} items", items.len());

        let mut candle_lighting = None;
        let mut havdalah = None;

        for item in items {
            if let Some(category) = item["category"].as_str() {
                println!("Category: {}, Title: {:?}", category, item["title"]);

                if category == "candles" {
                    candle_lighting = item["title"]
                        .as_str()
                        .and_then(|s| s.split("Candle lighting: ").nth(1));
                    println!("  Parsed candles: {:?}", candle_lighting);
                } else if category == "havdalah" {
                    havdalah = item["title"].as_str().and_then(|s| {
                        if let Some(after_colon) = s.split(": ").nth(1) {
                            Some(after_colon)
                        } else {
                            s.split("Havdalah: ").nth(1)
                        }
                    });
                    println!("  Parsed havdalah: {:?}", havdalah);
                }
            }
        }

        if let (Some(candles), Some(havd)) = (candle_lighting, havdalah) {
            println!("\nSUCCESS! Candles: {}, Havdalah: {}", candles, havd);
        } else {
            println!(
                "\nFAILED! Candles: {:?}, Havdalah: {:?}",
                candle_lighting, havdalah
            );
        }
    }
}
