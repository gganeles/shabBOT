use diesel::prelude::*;
use regex::Regex;
use crate::db::*;
use crate::utils::*;

pub fn shop_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
    let args = parse_args(prompt);

    if args.len() < 2 {
        return "Usage: !shop <item> or !shop clear".to_string();
    }

    let item_text = args[1..].join(" ");

    use crate::db::schema::shopping_list::dsl::{
        chat_id as sl_chat_id, id as sl_id, item as sl_item, quantity as sl_quantity, shopping_list,
    };

    if item_text.to_lowercase() == "clear" {
        if let Err(_) = diesel::delete(shopping_list.filter(sl_chat_id.eq(chat_id))).execute(db) {
            return "Error clearing shopping list".to_string();
        }
        return "Shopping list cleared.".to_string();
    }

    let quantity_regex = Regex::new(r"\d+").unwrap();
    let quantity: i32 = quantity_regex
        .find(&item_text)
        .and_then(|m| m.as_str().parse().ok())
        .unwrap_or(0);

    let item_text = quantity_regex
        .replace_all(&item_text, "")
        .trim()
        .to_string();

    if item_text.is_empty() {
        return "Usage: !shop <item>".to_string();
    }

    let _ = get_or_create_chat(db, chat_id);

    let all_items: Vec<(i32, String, i32)> = shopping_list
        .filter(sl_chat_id.eq(chat_id))
        .select((sl_id, sl_item, sl_quantity))
        .load::<(i32, String, i32)>(db)
        .unwrap_or_default();

    let existing = all_items
        .iter()
        .find(|(_, item, _)| item.to_lowercase() == item_text.to_lowercase())
        .map(|(id, _, qty)| (*id, *qty));

    if let Some((id, _)) = existing {
        if quantity > 0 {
            if let Err(_) = diesel::update(shopping_list.filter(sl_id.eq(id)))
                .set(sl_quantity.eq(quantity))
                .execute(db)
            {
                return "Error updating item".to_string();
            }
            return format!("\"{}\" quantity updated to {}.", item_text, quantity);
        }
        return format!("\"{}\" is already in the shopping list.", item_text);
    }

    if let Err(_) = diesel::insert_into(shopping_list)
        .values((
            sl_chat_id.eq(chat_id),
            sl_item.eq(&item_text),
            sl_quantity.eq(quantity),
        ))
        .execute(db)
    {
        return "Error adding item to shopping list".to_string();
    }

    let quantity_str = if quantity > 0 {
        format!("{} x ", quantity)
    } else {
        String::new()
    };

    let verb = if quantity > 1 { "have" } else { "has" };

    format!(
        "{}\"{}\" {} been added to shopping list",
        quantity_str, item_text, verb
    )
}

pub fn shoplist_cmd(db: &mut SqliteConnection, chat_id: &str) -> String {
    use crate::db::schema::shopping_list::dsl::{
        chat_id as sl_chat_id, item as sl_item, quantity as sl_quantity, shopping_list,
    };

    let items: Vec<(String, i32)> = shopping_list
        .filter(sl_chat_id.eq(chat_id))
        .select((sl_item, sl_quantity))
        .order(crate::db::schema::shopping_list::id)
        .load::<(String, i32)>(db)
        .unwrap_or_default();

    if items.is_empty() {
        return "Your shopping list is empty.".to_string();
    }

    let formatted_items: Vec<String> = items
        .iter()
        .enumerate()
        .map(|(i, (item, quantity))| {
            let quantity_str = if *quantity > 0 {
                format!("{} x ", quantity)
            } else {
                String::new()
            };
            format!("  {}.  {}{}", i + 1, quantity_str, item)
        })
        .collect();

    format!("Shopping list:\n{}", formatted_items.join("\n"))
}

pub fn unshop_cmd(db: &mut SqliteConnection, prompt: &str, chat_id: &str) -> String {
    let args = parse_args(prompt);

    if args.len() < 2 {
        return "Usage: !unshop <item>".to_string();
    }

    let item_to_remove = args[1..].join(" ");

    use crate::db::schema::shopping_list::dsl::{
        chat_id as sl_chat_id, id as sl_id, item as sl_item, shopping_list,
    };
    use diesel::dsl::count_star;

    let count: i64 = shopping_list
        .filter(sl_chat_id.eq(chat_id))
        .select(count_star())
        .first(db)
        .unwrap_or(0);

    if count == 0 {
        return "Your shopping list is empty.".to_string();
    }

    if let Ok(index) = item_to_remove.parse::<usize>() {
        let items: Vec<(i32, String)> = shopping_list
            .filter(sl_chat_id.eq(chat_id))
            .select((sl_id, sl_item))
            .order(sl_id)
            .load::<(i32, String)>(db)
            .unwrap_or_default();

        if index == 0 || index > items.len() {
            return "Invalid item index.".to_string();
        }

        let (item_id, item_name) = &items[index - 1];

        if let Err(_) = diesel::delete(shopping_list.filter(sl_id.eq(item_id))).execute(db) {
            return "Error removing item".to_string();
        }
        return format!(
            "\"{}\" has been removed from your shopping list.",
            item_name
        );
    }

    let all_items: Vec<(i32, String)> = shopping_list
        .filter(sl_chat_id.eq(chat_id))
        .select((sl_id, sl_item))
        .load::<(i32, String)>(db)
        .unwrap_or_default();

    let matching_item = all_items
        .iter()
        .find(|(_, item)| item.to_lowercase().contains(&item_to_remove.to_lowercase()))
        .map(|(id, item)| (*id, item.clone()));

    if let Some((item_id, item_name)) = matching_item {
        if let Err(_) = diesel::delete(shopping_list.filter(sl_id.eq(item_id))).execute(db) {
            return "Error removing item".to_string();
        }
        format!(
            "\"{}\" has been removed from your shopping list.",
            item_name
        )
    } else {
        format!("Item \"{}\" not found in shopping list.", item_to_remove)
    }
}
