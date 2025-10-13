package commands

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	dateParse "shabBOT/chrono-node"
)

// GeoName represents a location from the geoNames CSV
type GeoName struct {
	Name        string
	Timezone    string
	Coordinates string
	CountryName string
	CountryCode string
}

type searchObject struct {
	id      string
	message string
	index   int
}

func whatKeys(keyword string, items []searchObject) []string {
	var keys []string
	keys = make([]string, 0, 3)

	for _, item := range items {
		if strings.Contains(strings.ToLower(item.message), strings.ToLower(keyword)) {
			keys = append(keys, item.id)
		}
	}

	if len(keys) == 0 {
		// Try to extract a number from the keyword if direct match fails
		r := regexp.MustCompile(`\b\d+\b`)
		matches := r.FindAllString(strings.ToLower(keyword), -1)
		if len(matches) > 0 {
			// Try to parse the first number found as an index
			index := parseInt(matches[0], -1)
			// Validate index is within bounds
			if index >= 0 && index < len(items) {
				keys = append(keys, items[index].id)
			}
		}
	}

	return keys
}

var geoNamesList []GeoName

// init loads the geoNames CSV file at package initialization
func init() {
	file, err := os.Open("commands/geoNamesList.csv")
	if err != nil {
		// Try alternative path
		file, err = os.Open("commands/geoNamesList.csv")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load geoNamesList.csv: %v\n", err)
			return
		}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not parse geoNamesList.csv: %v\n", err)
		return
	}

	// Skip header row
	if len(records) > 0 {
		records = records[1:]
	}

	geoNamesList = make([]GeoName, 0, len(records))
	for _, record := range records {
		if len(record) < 20 {
			continue
		}
		// Column 2 is "ASCII Name", column 7 is "Country name EN", column 16 is "Timezone", column 19 is "Coordinates"
		name := strings.TrimSpace(record[2])
		countryName := strings.TrimSpace(record[7])
		timezone := strings.TrimSpace(record[16])
		coordinates := strings.TrimSpace(record[19])
		countryCode := strings.TrimSpace(record[8]) // Country code
		if name != "" && timezone != "" && coordinates != "" {
			geoNamesList = append(geoNamesList, GeoName{
				Name:        name,
				Timezone:    timezone,
				Coordinates: coordinates,
				CountryName: countryName,
				CountryCode: countryCode,
			})
		}
	}
}

// parseArgs splits a prompt into arguments by whitespace
func parseArgs(prompt string) []string {
	return strings.Fields(prompt)
}

// parseInt safely parses an integer with default value
func parseInt(s string, defaultVal int) int {
	var result int
	if _, err := fmt.Sscanf(s, "%d", &result); err != nil {
		return defaultVal
	}
	return result
}

func parseTime(s string, timezone *time.Location) (int64, string, error) {
	// Use the when library to parse natural language dates
	loc := time.UTC
	if timezone != nil {
		loc = timezone
	}

	now := time.Now().In(loc)

	parsed, err := dateParse.New()
	if err != nil {
		return 0, "", fmt.Errorf("could not parse time")
	}

	parsedDate, err := parsed.Parse(s, now)

	if err != nil || len(parsedDate) == 0 {
		return 0, "", fmt.Errorf("could not parse time")
	}

	messageHalves := strings.Split(s, parsedDate[0].Text)
	for i := range messageHalves {
		messageHalves[i] = strings.TrimSpace(messageHalves[i])
	}
	timeStr := parsedDate[0].Time.Local().Format(time.DateTime)

	message := strings.Join(messageHalves, " ")
	cleanedTime, err := time.ParseInLocation(time.DateTime, timeStr, loc)
	if err != nil {
		return 0, "", fmt.Errorf("could not parse time")
	}
	return cleanedTime.Unix(), strings.TrimSpace(message), nil
}

// findGeoLocation searches for a location in the geoNamesList
func findGeoLocation(location string) *GeoName {
	if location == "" {
		return nil
	}

	for _, geo := range geoNamesList {
		if strings.EqualFold(geo.Name, location) {
			return &geo
		}
	}
	return nil
}
