package main

import (
	"fmt"
	"log"
	"time"

	chrononode "shabBOT/chrono-node"
)

func main() {
	parser, err := chrononode.New()
	if err != nil {
		log.Fatal("Failed to create parser:", err)
	}

	baseTime := time.Now()

	// Test Parse
	fmt.Println("Testing Parse...")
	results, err := parser.Parse("tomorrow at 3pm", baseTime)
	if err != nil {
		log.Fatal("Parse failed:", err)
	}
	fmt.Printf("Parse results: %+v\n\n", results)

	// Test ParseDate
	fmt.Println("Testing ParseDate...")
	date, err := parser.ParseDate("next friday", baseTime)
	if err != nil {
		log.Fatal("ParseDate failed:", err)
	}
	if date != nil {
		fmt.Printf("ParseDate result: %s\n", date.Format(time.RFC3339))
	} else {
		fmt.Println("ParseDate result: nil")
	}
}
