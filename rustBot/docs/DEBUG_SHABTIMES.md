# Debugging Shabbat Times API

## Issue
The `!shabtimes` command is returning the fallback message instead of actual times from the Hebcal API.

## Expected Behavior
```
User: !shabtimes
Bot: Shabbat in Haifa starts this week at 17:46 and ends at 18:53
```

## Current Behavior (Fallback)
```
User: !shabtimes
Bot: Shabbat times for Haifa - Candle lighting approximately 18 minutes before sunset on Friday
```

## Debugging Steps

### 1. Check if API is Reachable
```bash
curl -s "https://www.hebcal.com/shabbat?cfg=json&geo=pos&latitude=32.8191&longitude=34.9983&tzid=Asia%2FJerusalem&m=50" | jq '.items[] | {category, title}'
```

Expected output:
```json
{"category":"candles","title":"Candle lighting: 17:46"}
{"category":"havdalah","title":"Havdalah (50 min): 18:53"}
```

### 2. Check Bot Logs
When you run the bot and send `!shabtimes`, you should see log output like:

```
[INFO] Fetching Shabbat times from: https://www.hebcal.com/shabbat?...
[INFO] API response status: 200
[INFO] Found 4 items in response
[DEBUG] Item category: candles, title: "Candle lighting: 17:46"
[DEBUG] Item category: havdalah, title: "Havdalah (50 min): 18:53"
[INFO] Successfully parsed times - Candles: 17:46, Havdalah: 18:53
```

If you see:
- `[ERROR] Error fetching Shabbat times:` - Network/HTTP error
- `[ERROR] Failed to parse JSON:` - JSON parsing error
- `[WARN] No 'items' array in response` - API changed format
- `[WARN] Could not find candle lighting or havdalah` - Parsing error

### 3. Enable Debug Logging
Make sure your env_logger is set to show all logs:

```bash
RUST_LOG=debug cargo run
```

Or in your terminal before running the bot:
```fish
set -x RUST_LOG debug
cargo run
```

### 4. Common Issues and Fixes

#### Issue: Network Error
**Symptoms**: `[ERROR] Error fetching Shabbat times: timed out`

**Fix**: Check internet connection, firewall rules, or add a timeout:
```rust
let response = ureq::get(&api_url)
    .timeout(std::time::Duration::from_secs(10))
    .call()?;
```

#### Issue: SSL/TLS Error  
**Symptoms**: `[ERROR] Error fetching Shabbat times: certificate verify failed`

**Fix**: Add `native-tls` feature to ureq in Cargo.toml:
```toml
ureq = { version = "2.10", features = ["json", "native-tls"] }
```

#### Issue: Parsing Error
**Symptoms**: `[WARN] Could not find candle lighting or havdalah in response`

**Fix**: The parsing logic has been updated to handle:
- "Candle lighting: HH:MM"
- "Havdalah (X min): HH:MM"

If you see this warning, check the actual `title` fields in the logs.

#### Issue: Wrong Location
**Symptoms**: Fallback message but logs show success

**Fix**: Check that:
1. Location is set: `!shablocation`
2. Location exists in geoNamesList.csv
3. Coordinates are valid

### 5. Test API Directly in Code

Add this test to verify the parsing:

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_api_parse() {
        let json = r#"{"items":[
            {"title":"Candle lighting: 17:46","category":"candles"},
            {"title":"Havdalah (50 min): 18:53","category":"havdalah"}
        ]}"#;
        
        let result: serde_json::Value = serde_json::from_str(json).unwrap();
        let items = result["items"].as_array().unwrap();
        
        let mut candles = None;
        let mut havdalah = None;
        
        for item in items {
            let category = item["category"].as_str().unwrap();
            if category == "candles" {
                candles = item["title"].as_str()
                    .and_then(|s| s.split("Candle lighting: ").nth(1));
            } else if category == "havdalah" {
                havdalah = item["title"].as_str()
                    .and_then(|s| s.split(": ").nth(1));
            }
        }
        
        assert_eq!(candles, Some("17:46"));
        assert_eq!(havdalah, Some("18:53"));
    }
}
```

Run with: `cargo test test_api_parse`

### 6. Manual Test Commands

```bash
# In the bot terminal, send these WhatsApp messages:
!shablocation Haifa          # Set location
!shablocation                # Verify location
!shabtimes                   # Test Shabbat times
!shabtimes Jerusalem         # Test with different location
!shabtimes New York          # Test with US location
```

## Current Code Status

✅ **Fixed Issues:**
- Replaced `reqwest::blocking` with `ureq` (no more runtime errors)
- Updated Havdalah parsing to handle "(X min)" format
- Added comprehensive logging at INFO and DEBUG levels

✅ **Code Compiles:** 0 errors

⚠️ **Next Steps:**
1. Run bot with `RUST_LOG=debug`
2. Send `!shabtimes` command
3. Check logs for specific error
4. Report exact error message for further debugging

## Expected Log Output

### Success Case:
```
[INFO] Fetching Shabbat times from: https://www.hebcal.com/shabbat?cfg=json&geo=pos&latitude=32.8191&longitude=34.9983&tzid=Asia%2FJerusalem&m=50
[INFO] API response status: 200
[DEBUG] API response: {"title":"Hebcal...","items":[...]}
[INFO] Found 4 items in response
[DEBUG] Item category: candles, title: "Candle lighting: 17:46"
[DEBUG] Item category: parashat, title: "Parashat Bereshit"
[DEBUG] Item category: mevarchim, title: "Mevarchim Chodesh Cheshvan"
[DEBUG] Item category: havdalah, title: "Havdalah (50 min): 18:53"
[INFO] Successfully parsed times - Candles: 17:46, Havdalah: 18:53
```

### Failure Case (what to look for):
```
[ERROR] Error fetching Shabbat times: <specific error here>
```

The logs will tell us exactly what's failing!
