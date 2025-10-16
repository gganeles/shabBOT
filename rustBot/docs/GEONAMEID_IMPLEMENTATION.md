# Using Geoname ID in Hebcal API Requests

## Changes Made

Updated the Shabbat times and holiday times commands to use Geoname ID instead of latitude/longitude coordinates when making Hebcal API requests.

## Why This Change?

### Before (Using lat/lng)
```rust
let api_url = format!(
    "https://www.hebcal.com/shabbat?cfg=json&geo=pos&latitude={}&longitude={}&tzid={}&m={}",
    lat, lng, urlencoding::encode(&geo_data.timezone), offset_minutes
);
```

**Issues:**
- More complex URL construction
- Requires parsing and validating coordinates
- Timezone must be URL-encoded
- More error-prone

### After (Using geonameid)
```rust
let api_url = format!(
    "https://www.hebcal.com/shabbat?cfg=json&geonameid={}&m={}",
    geo_data.geoname_id, offset_minutes
);
```

**Benefits:**
- ✅ Simpler URL construction
- ✅ No coordinate parsing needed
- ✅ No timezone encoding needed
- ✅ More reliable (uses Hebcal's built-in geoname database)
- ✅ Hebcal automatically provides correct timezone
- ✅ Shorter URLs

## Implementation Details

### 1. Updated GeoName Struct

Added `geoname_id` field:

```rust
#[derive(Debug, Clone, Deserialize)]
pub struct GeoName {
    pub geoname_id: String,    // NEW: Geoname ID from column 0
    pub name: String,
    pub timezone: String,
    pub coordinates: String,
    pub country_name: String,
    pub country_code: String,
}
```

### 2. Updated CSV Loading

Now extracts Geoname ID from column 0:

```rust
let geoname_id = record.get(0).unwrap_or("").trim().to_string();
let name = record.get(2).unwrap_or("").trim().to_string();
// ...

if !geoname_id.is_empty() && !name.is_empty() && !timezone.is_empty() {
    geonames.push(GeoName {
        geoname_id,
        name,
        // ...
    });
}
```

### 3. Updated API Calls

#### `shabtimes_cmd`
```rust
// OLD
let api_url = format!(
    "https://www.hebcal.com/shabbat?cfg=json&geo=pos&latitude={}&longitude={}&tzid={}&m=on",
    lat, lng, urlencoding::encode(&geo_data.timezone)
);

// NEW
let api_url = format!(
    "https://www.hebcal.com/shabbat?cfg=json&geonameid={}&m={}",
    geo_data.geoname_id, offset_minutes
);
```

**Note:** The `m` parameter (candle lighting offset in minutes) now uses the numeric value instead of "on".

#### `chagtimes_cmd`
```rust
// OLD
let api_url = format!(
    "https://www.hebcal.com/hebcal?v=1&cfg=json&maj=on&min=on&mod=on&nx=on&year=now&month=x&ss=on&mf=on&c=on&geo=pos&latitude={}&longitude={}&tzid={}&m=50",
    lat, lng, urlencoding::encode(&geo_data.timezone)
);

// NEW
let api_url = format!(
    "https://www.hebcal.com/hebcal?v=1&cfg=json&maj=on&min=on&mod=on&nx=on&year=now&month=x&ss=on&mf=on&c=on&geonameid={}&m=50",
    geo_data.geoname_id
);
```

## Example API Requests

### Old Format
```
https://www.hebcal.com/shabbat?cfg=json&geo=pos&latitude=32.8191&longitude=34.9983&tzid=Asia%2FJerusalem&m=on
```

### New Format
```
https://www.hebcal.com/shabbat?cfg=json&geonameid=294801&m=18
```

Much cleaner!

## Testing

### Test Haifa (Geoname ID: 294801)
```bash
curl "https://www.hebcal.com/shabbat?cfg=json&geonameid=294801&m=18"
```

### Test Jerusalem (Geoname ID: 281184)
```bash
curl "https://www.hebcal.com/shabbat?cfg=json&geonameid=281184&m=40"
```

### Test New York (Geoname ID: 5128581)
```bash
curl "https://www.hebcal.com/shabbat?cfg=json&geonameid=5128581&m=18"
```

## Candle Lighting Offsets

The `m` parameter specifies minutes before sunset for candle lighting:

```rust
let offset_minutes = match geo_data.name.as_str() {
    "Jerusalem" => 40,            // 40 minutes before sunset
    "Haifa" | "Zichron Yaakov" => 30,  // 30 minutes before sunset
    _ => 18,                      // 18 minutes (standard)
};
```

## Benefits Summary

1. **Simpler Code**: Removed coordinate parsing logic
2. **More Reliable**: Hebcal's geoname database is authoritative
3. **Better URLs**: No URL encoding needed for timezones
4. **Consistent**: Both Shabbat and Holiday APIs now use same format
5. **Efficient**: Shorter URLs, faster parsing

## Compilation Status

✅ **Successfully compiles**
- 0 errors
- Only warnings about unused functions (expected)

## CSV Data Structure

The geoNamesList.csv contains:
- **Column 0**: Geoname ID (e.g., "294801" for Haifa)
- **Column 2**: ASCII Name (e.g., "Haifa")
- **Column 7**: Country Name EN (e.g., "Israel")
- **Column 8**: Country Code (e.g., "IL")
- **Column 16**: Timezone (e.g., "Asia/Jerusalem")
- **Column 19**: Coordinates (e.g., "32.8191, 34.9983")

All fields are now properly extracted and used appropriately.
