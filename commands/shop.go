package commands

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ShopCmd adds items to shopping list
func ShopCmd(db *sql.DB, prompt, chatID string) string {
	args := parseArgs(prompt)

	if len(args) < 2 {
		return "Usage: !shop <item> or !shop clear"
	}

	// Join remaining args as item text (may contain multiple items separated by commas or newlines)
	rawItems := strings.Join(args[1:], " ")

	// Check for clear command
	if strings.ToLower(strings.TrimSpace(rawItems)) == "clear" {
		_, err := db.Exec(`DELETE FROM shopping_list WHERE chat_id = ?`, chatID)
		if err != nil {
			return "Error clearing shopping list"
		}
		return "Shopping list cleared."
	}

	// Split on commas or newlines to allow adding multiple items at once
	// e.g. "apples, bananas\ncarrots"
	splitter := regexp.MustCompile(`[\r\n,]+`)
	parts := splitter.Split(rawItems, -1)

	if len(parts) == 0 {
		return "Usage: !shop <item>"
	}

	// Ensure chat exists
	GetOrCreateChat(db, chatID)

	// Build responses for each item
	var responses []string
	quantityRegex := regexp.MustCompile(`\d+`)

	for _, p := range parts {
		itemText := strings.TrimSpace(p)
		if itemText == "" {
			continue
		}

		// Extract quantity if present (e.g., "5 apples" or "apples 5")
		quantityMatch := quantityRegex.FindString(itemText)
		quantity := 0
		if quantityMatch != "" {
			quantity, _ = strconv.Atoi(quantityMatch)
			// Remove quantity from item text
			itemText = strings.TrimSpace(quantityRegex.ReplaceAllString(itemText, ""))
		}

		if itemText == "" {
			// Skip if nothing left after removing numbers
			continue
		}

		// Check if item already exists (case-insensitive)
		var existingID int
		var existingQty int
		err := db.QueryRow(`
			SELECT id, quantity FROM shopping_list
			WHERE chat_id = ? AND LOWER(item) = LOWER(?)
		`, chatID, itemText).Scan(&existingID, &existingQty)

		if err == nil {
			// Item exists, update quantity if provided
			if quantity > 0 {
				_, err = db.Exec(`
					UPDATE shopping_list SET quantity = ?
					WHERE id = ?
				`, quantity, existingID)
				if err != nil {
					responses = append(responses, fmt.Sprintf("Error updating %s", itemText))
					continue
				}
				responses = append(responses, fmt.Sprintf(`"%s" quantity updated to %d.`, itemText, quantity))
				continue
			}
			responses = append(responses, fmt.Sprintf(`"%s" is already in the shopping list.`, itemText))
			continue
		}

		// Add new item
		_, err = db.Exec(`
			INSERT INTO shopping_list (chat_id, item, quantity)
			VALUES (?, ?, ?)
		`, chatID, itemText, quantity)

		if err != nil {
			responses = append(responses, fmt.Sprintf("Error adding %s to shopping list", itemText))
			continue
		}

		quantityStr := ""
		if quantity > 0 {
			quantityStr = fmt.Sprintf("%d x ", quantity)
		}

		verb := "has"
		if quantity > 1 {
			verb = "have"
		}

		responses = append(responses, fmt.Sprintf(`%s"%s" %s been added to shopping list`, quantityStr, itemText, verb))
	}

	if len(responses) == 0 {
		return "Usage: !shop <item>"
	}

	return strings.Join(responses, "\n")
}

// ShopListCmd displays the shopping list
func ShopListCmd(db *sql.DB, chatID string) string {
	rows, err := db.Query(`
		SELECT item, quantity FROM shopping_list
		WHERE chat_id = ?
		ORDER BY id
	`, chatID)

	if err != nil {
		return "Error retrieving shopping list"
	}
	defer rows.Close()

	var items []string
	index := 1

	for rows.Next() {
		var item string
		var quantity int
		rows.Scan(&item, &quantity)

		quantityStr := ""
		if quantity > 0 {
			quantityStr = fmt.Sprintf("%d x ", quantity)
		}

		items = append(items, fmt.Sprintf("  %d.  %s%s", index, quantityStr, item))
		index++
	}

	if len(items) == 0 {
		return "Your shopping list is empty."
	}

	return "Shopping list:\n" + strings.Join(items, "\n")
}

// UnshopCmd removes items from the shopping list
func UnshopCmd(db *sql.DB, prompt, chatID string) string {
	args := parseArgs(prompt)

	if len(args) < 2 {
		return "Usage: !unshop <item>"
	}

	// Check if list is empty
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM shopping_list WHERE chat_id = ?`, chatID).Scan(&count)
	if count == 0 {
		return "Your shopping list is empty."
	}

	// Join remaining args as item text (may contain multiple items separated by commas or newlines)
	rawItems := strings.Join(args[1:], " ")

	// Split on commas or newlines to allow removing multiple items at once
	splitter := regexp.MustCompile(`[\r\n,]+`)
	parts := splitter.Split(rawItems, -1)

	var responses []string

	for _, p := range parts {
		itemToRemove := strings.TrimSpace(p)
		if itemToRemove == "" {
			continue
		}

		itemName, err := findItemToRemove(db, chatID, itemToRemove)
		if err != nil {
			responses = append(responses, err.Error())
			continue
		}

		if err := removeShoppingItem(db, chatID, itemName); err != nil {
			responses = append(responses, fmt.Sprintf("Error removing %s", itemName))
			continue
		}

		responses = append(responses, fmt.Sprintf(`"%s" has been removed from your shopping list.`, itemName))
	}

	if len(responses) == 0 {
		return "Usage: !unshop <item>"
	}

	return strings.Join(responses, "\n")
}

// findItemToRemove returns the item name to remove based on index or name search
func findItemToRemove(db *sql.DB, chatID, input string) (string, error) {
	// Check if it's a number (index)
	if index, err := strconv.Atoi(input); err == nil {
		var itemName string
		err := db.QueryRow(`
			SELECT item FROM shopping_list
			WHERE chat_id = ?
			ORDER BY id
			LIMIT 1 OFFSET ?
		`, chatID, index-1).Scan(&itemName)

		if err != nil {
			return "", fmt.Errorf("invalid item index")
		}
		return itemName, nil
	}

	// Search by name (partial match)
	var itemName string
	err := db.QueryRow(`
		SELECT item FROM shopping_list
		WHERE chat_id = ? AND item LIKE ?
		LIMIT 1
	`, chatID, "%"+input+"%").Scan(&itemName)

	if err != nil {
		return "", fmt.Errorf(`item "%s" not found in shopping list`, input)
	}

	return itemName, nil
}

// removeShoppingItem deletes a single item from the shopping list
func removeShoppingItem(db *sql.DB, chatID, itemName string) error {
	_, err := db.Exec(`
		DELETE FROM shopping_list
		WHERE chat_id = ? AND item = ?
		LIMIT 1
	`, chatID, itemName)
	return err
}
