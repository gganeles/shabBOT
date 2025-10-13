package commands

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hebcal/hebcal-go/zmanim"
)

// ShabTimesCmd calculates and returns Shabbat times for a given location and date
func ShabTimesCmd(db *sql.DB, prompt, chatID string) string {
	// Get chat's default location
	location, err := GetChatLocation(db, chatID)
	if err != nil {
		location = "Haifa"
	}

	// Parse prompt for alternate location
	args := parseArgs(prompt)
	if len(args) > 1 {
		// Check if a location is specified in the prompt
		potentialLocation := strings.Join(args[1:], " ")
		if potentialLocation != "" {
			location = potentialLocation
		}
	}

	// Find location data from geoNamesList
	geoData := findGeoLocation(location)
	if geoData == nil {
		return fmt.Sprintf("Location '%s' not found in database. Try !shablocation to set a valid location.", location)
	}

	// Parse coordinates (format: "lat, lng")
	coords := strings.Split(geoData.Coordinates, ",")
	if len(coords) != 2 {
		return fmt.Sprintf("Invalid coordinates for location '%s'", location)
	}

	lat, err := strconv.ParseFloat(strings.TrimSpace(coords[0]), 64)
	if err != nil {
		return fmt.Sprintf("Error parsing latitude for '%s'", location)
	}

	lng, err := strconv.ParseFloat(strings.TrimSpace(coords[1]), 64)
	if err != nil {
		return fmt.Sprintf("Error parsing longitude for '%s'", location)
	}

	// Load timezone
	loc, err := time.LoadLocation(geoData.Timezone)
	if err != nil {
		loc = time.UTC
	}

	// Find next Friday (or today if it's Friday)
	now := time.Now().In(loc)
	daysUntilFriday := (5 - int(now.Weekday()) + 7) % 7
	if daysUntilFriday == 0 && now.Weekday() == time.Friday {
		// It's Friday - check if candle lighting has passed
		// If it's past candle lighting time, show next week
		testTime := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, loc)
		if now.After(testTime) {
			daysUntilFriday = 7
		}
	} else if daysUntilFriday == 0 {
		daysUntilFriday = 7
	}

	fridayDate := now.AddDate(0, 0, daysUntilFriday)
	saturdayDate := fridayDate.AddDate(0, 0, 1)

	zman_loc := zmanim.Location{
		Name:        geoData.Name,
		Latitude:    lat,
		Longitude:   lng,
		CountryCode: geoData.CountryCode,
		TimeZoneId:  geoData.Timezone,
	}

	zman := zmanim.New(&zman_loc, fridayDate)

	offset := 0
	switch geoData.Name {
	case "Jerusalem":
		offset = -40
	case "Haifa", "Zichron Yaakov":
		offset = -30
	default:
		offset = -18
	}

	// Calculate sunset times
	fridaySunset := zman.SunsetOffset(offset, true)

	zman = zmanim.New(&zman_loc, saturdayDate)

	saturdaySunset := zman.Tzeit(8.5) // 8.5 degrees for nightfall

	havdalah := saturdaySunset.In(loc)

	// Check if we're currently in Shabbat
	if now.After(fridaySunset) && now.Before(havdalah) && daysUntilFriday == 0 {
		return fmt.Sprintf("Shabbat Shalom! 🕯️\nHavdalah in %s is at %s",
			geoData.Name,
			havdalah.Format("3:04 PM"))
	}

	return fmt.Sprintf("Shabbat in %s starts this week at %s and ends at %s",
		geoData.Name,
		fridaySunset.Format("3:04 PM"),
		havdalah.Format("3:04 PM"))
}

// ShabTimesAPICmd is an alternative that uses an external API
// Example using hebcal.com API
func ShabTimesAPICmd(db *sql.DB, prompt, chatID, location string) string {
	// This would make an HTTP request to https://www.hebcal.com/shabbat/
	// Example: https://www.hebcal.com/shabbat?cfg=json&geonameid=294801&M=on

	return fmt.Sprintf(
		"To implement this properly:\n"+
			"1. Use net/http to call hebcal.com API\n"+
			"2. Parse JSON response\n"+
			"3. Format and return candle lighting and Havdalah times\n\n"+
			"API endpoint: https://www.hebcal.com/shabbat?cfg=json&geo=geoname&city=%s",
		location,
	)
}

// ChagTimesCmd calculates and returns the next Jewish holiday with times for a given location
func ChagTimesCmd(db *sql.DB, prompt, chatID string) string {
	// Get chat's default location
	location, err := GetChatLocation(db, chatID)
	if err != nil {
		location = "Haifa"
	}

	// Parse prompt for alternate location
	args := parseArgs(prompt)
	if len(args) > 1 {
		// Check if a location is specified in the prompt
		potentialLocation := strings.Join(args[1:], " ")
		if potentialLocation != "" {
			location = potentialLocation
		}
	}

	// Find location data from geoNamesList
	geoData := findGeoLocation(location)
	if geoData == nil {
		return fmt.Sprintf("Location '%s' not found in database. Try !shablocation to set a valid location.", location)
	}

	// Parse coordinates (format: "lat, lng")
	coords := strings.Split(geoData.Coordinates, ",")
	if len(coords) != 2 {
		return fmt.Sprintf("Invalid coordinates for location '%s'", location)
	}

	lat, err := strconv.ParseFloat(strings.TrimSpace(coords[0]), 64)
	if err != nil {
		return fmt.Sprintf("Error parsing latitude for '%s'", location)
	}

	lng, err := strconv.ParseFloat(strings.TrimSpace(coords[1]), 64)
	if err != nil {
		return fmt.Sprintf("Error parsing longitude for '%s'", location)
	}

	// Load timezone
	loc, err := time.LoadLocation(geoData.Timezone)
	if err != nil {
		loc = time.UTC
	}

	// Call Hebcal REST API to get upcoming holidays
	// API: https://www.hebcal.com/hebcal?v=1&cfg=json&maj=on&min=on&mod=on&nx=on&year=now&month=x&ss=on&mf=on&c=on&geo=pos&latitude=LAT&longitude=LONG&tzid=TIMEZONE&m=50
	apiURL := fmt.Sprintf(
		"https://www.hebcal.com/hebcal?v=1&cfg=json&maj=on&min=on&mod=on&nx=on&year=now&month=x&ss=on&mf=on&c=on&geo=pos&latitude=%f&longitude=%f&tzid=%s&m=50",
		lat, lng, url.QueryEscape(geoData.Timezone),
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Sprintf("Error fetching holiday data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Error: Hebcal API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Error reading API response: %v", err)
	}

	// Parse JSON response
	var result HebcalResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Sprintf("Error parsing API response: %v", err)
	}

	// Find the next major holiday (excluding Shabbat)
	now := time.Now().In(loc)
	var nextHoliday *HebcalItem

	for i := range result.Items {
		fmt.Println("Item:", result.Items[i])
		item := &result.Items[i]

		// Skip Shabbat and candle lighting
		if strings.Contains(item.Category, "candles") ||
			strings.Contains(item.Category, "havdalah") ||
			item.Title == "Parashat" ||
			strings.HasPrefix(item.Title, "Parashat ") {
			continue
		}

		// Parse item date
		itemDate, err := time.Parse("2006-01-02", item.Date)
		if err != nil {
			continue
		}

		// Convert to local timezone
		itemDate = time.Date(itemDate.Year(), itemDate.Month(), itemDate.Day(), 0, 0, 0, 0, loc)

		// Check if this is in the future
		if itemDate.After(now) || itemDate.Equal(now.Truncate(24*time.Hour)) {
			nextHoliday = item
			break
		}
	}

	if nextHoliday == nil {
		return "No upcoming holidays found in the next 2 months."
	}

	// Format the response
	response := fmt.Sprintf("Next Holiday: %s 🕎\n", nextHoliday.Title)

	if nextHoliday.Hebrew != "" {
		response += fmt.Sprintf("Hebrew: %s\n", nextHoliday.Hebrew)
	}

	// Parse and format the date
	holidayDate, _ := time.Parse("2006-01-02", nextHoliday.Date)
	holidayDate = time.Date(holidayDate.Year(), holidayDate.Month(), holidayDate.Day(), 0, 0, 0, 0, loc)
	response += fmt.Sprintf("Date: %s\n", holidayDate.Format("Monday, January 2, 2006"))

	// Calculate days until
	daysUntil := int(holidayDate.Sub(now.Truncate(24*time.Hour)).Hours() / 24)
	switch daysUntil {
	case 0:
		response += "That's today!\n"
	case 1:
		response += "That's tomorrow!\n"
	default:
		response += fmt.Sprintf("That's in %d days\n", daysUntil)
	}

	// If there's a time associated with the holiday (candle lighting, etc.)
	if nextHoliday.Date != "" {
		// Check if it's a holiday that requires candle lighting
		if strings.Contains(nextHoliday.Category, "holiday") &&
			!strings.Contains(strings.ToLower(nextHoliday.Title), "purim") &&
			!strings.Contains(strings.ToLower(nextHoliday.Title), "tu b") {

			// Calculate candle lighting for the holiday
			zman_loc := zmanim.Location{
				Name:        geoData.Name,
				Latitude:    lat,
				Longitude:   lng,
				CountryCode: geoData.CountryCode,
				TimeZoneId:  geoData.Timezone,
			}

			zman := zmanim.New(&zman_loc, holidayDate)

			offset := -18
			switch geoData.Name {
			case "Jerusalem":
				offset = -40
			case "Haifa", "Zichron Yaakov":
				offset = -30
			}

			candleLighting := zman.SunsetOffset(offset, true)
			response += fmt.Sprintf("\nCandle lighting: %s", candleLighting.Format("3:04 PM"))
		}
	}

	return response
}

// HebcalResponse represents the JSON response from Hebcal API
type HebcalResponse struct {
	Title string       `json:"title"`
	Date  string       `json:"date"`
	Items []HebcalItem `json:"items"`
}

// HebcalItem represents a single event from the Hebcal API
type HebcalItem struct {
	Title    string `json:"title"`
	Date     string `json:"date"`
	Hebrew   string `json:"hebrew"`
	Category string `json:"category"`
	Subcat   string `json:"subcat"`
	Yomtov   bool   `json:"yomtov"`
	Link     string `json:"link"`
}
