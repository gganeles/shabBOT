package commands

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	when "github.com/olebedev/when"
	en "github.com/olebedev/when/rules/en"
)

// GeoName represents a location from the geoNames CSV
type GeoName struct {
	Name        string
	Timezone    string
	Coordinates string
	CountryName string
	CountryCode string
}

var geoNamesList []GeoName

// init loads the geoNames CSV file at package initialization
func init() {
	file, err := os.Open("golangShabBOT/commands/geoNamesList.csv")
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

func parseTime(s string, location string) (int64, error) {
	// Use the when library to parse natural language dates
	loc := time.UTC
	if location != "" {
		for _, geo := range geoNamesList {
			if strings.EqualFold(geo.Name, location) {
				if geo.Timezone != "" {
					var err error
					loc, err = time.LoadLocation(geo.Timezone)
					if err != nil {
						return 0, fmt.Errorf("invalid timezone for location: %s", geo.Timezone)
					}
				}
				break
			}
		}
	}

	w := when.New(nil)
	w.Add(en.All...)
	// Parse time in the correct timezone
	now := time.Now().In(loc)
	parsed, err := w.Parse(s, now)
	if err != nil || parsed == nil {
		return 0, fmt.Errorf("could not parse time")
	}

	return parsed.Time.Unix(), nil
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
