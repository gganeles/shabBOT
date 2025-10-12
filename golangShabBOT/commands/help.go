package commands

// HelpCmd returns the bot introduction and basic usage instructions
func HelpCmd(prompt string) string {
	// Check if help has additional parameters
	args := parseArgs(prompt)
	if len(args) > 1 {
		// Extended help could be implemented here if needed
		return ""
	}

	response := "Allow me to introduce myself!\n\n" +
		"i am the SHABbot!!\n" +
		"I can help you with you shabbat meals, as well as other things\n\n" +
		"The way it works:\n\n" +
		" - Use !quickshab to keep track of what everyone's bringing\n" +
		" - Use !shabtimes to find out when shabbat starts\n" +
		" - Use !remind to set reminders for yourself\n\n" +
		"For more information, type !docs to see the full documentation"

	return response
}
