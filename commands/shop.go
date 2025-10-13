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

	// Join remaining args as item name
	itemText := strings.Join(args[1:], " ")
	
	// Check for clear command
	if strings.ToLower(itemText) == "clear" {
		_, err := db.Exec(`DELETE FROM shopping_list WHERE chat_id = ?`, chatID)
		if err != nil {
			return "Error clearing shopping list"
		}
		return "Shopping list cleared."
	}

	// Extract quantity if present (e.g., "5 apples" or "apples 5")
	quantityRegex := regexp.MustCompile(`\d+`)
	quantityMatch := quantityRegex.FindString(itemText)
	
	quantity := 0
	if quantityMatch != "" {
		quantity, _ = strconv.Atoi(quantityMatch)
		// Remove quantity from item text
		itemText = strings.TrimSpace(quantityRegex.ReplaceAllString(itemText, ""))
	}

	if itemText == "" {
		return "Usage: !shop <item>"
	}

	// Ensure chat exists
	GetOrCreateChat(db, chatID)

	// Check if item already exists
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
				return "Error updating item"
			}
			return fmt.Sprintf(`"%s" quantity updated to %d.`, itemText, quantity)
		}
		return fmt.Sprintf(`"%s" is already in the shopping list.`, itemText)
	}

	// Add new item
	_, err = db.Exec(`
		INSERT INTO shopping_list (chat_id, item, quantity) 
		VALUES (?, ?, ?)
	`, chatID, itemText, quantity)

	if err != nil {
		return "Error adding item to shopping list"
	}

	quantityStr := ""
	if quantity > 0 {
		quantityStr = fmt.Sprintf("%d x ", quantity)
	}
	
	verb := "has"
	if quantity > 1 {
		verb = "have"
	}

	return fmt.Sprintf(`%s"%s" %s been added to shopping list`, quantityStr, itemText, verb)
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

	itemToRemove := strings.Join(args[1:], " ")

	// Check if list is empty
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM shopping_list WHERE chat_id = ?`, chatID).Scan(&count)
	if count == 0 {
		return "Your shopping list is empty."
	}

	// Check if it's a number (index)
	if index, err := strconv.Atoi(itemToRemove); err == nil {
		// Remove by index
		var itemName string
		err := db.QueryRow(`
			SELECT item FROM shopping_list 
			WHERE chat_id = ? 
			ORDER BY id 
			LIMIT 1 OFFSET ?
		`, chatID, index-1).Scan(&itemName)

		if err != nil {
			return "Invalid item index."
		}

		_, err = db.Exec(`
			DELETE FROM shopping_list 
			WHERE chat_id = ? AND item = ?
			AND id = (
				SELECT id FROM shopping_list 
				WHERE chat_id = ? AND item = ?
				LIMIT 1
			)
		`, chatID, itemName, chatID, itemName)

		if err != nil {
			return "Error removing item"
		}

		return fmt.Sprintf(`"%s" has been removed from your shopping list.`, itemName)
	}

	// Remove by name (partial match)
	var itemName string
	err := db.QueryRow(`
		SELECT item FROM shopping_list 
		WHERE chat_id = ? AND item LIKE ?
		LIMIT 1
	`, chatID, "%"+itemToRemove+"%").Scan(&itemName)

	if err != nil {
		return fmt.Sprintf(`Item "%s" not found in shopping list.`, itemToRemove)
	}

	_, err = db.Exec(`
		DELETE FROM shopping_list 
		WHERE chat_id = ? AND item = ?
		AND id = (
			SELECT id FROM shopping_list 
			WHERE chat_id = ? AND item = ?
			LIMIT 1
		)
	`, chatID, itemName, chatID, itemName)

	if err != nil {
		return "Error removing item"
	}

	return fmt.Sprintf(`"%s" has been removed from your shopping list.`, itemName)
}
