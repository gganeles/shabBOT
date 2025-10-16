// @generated automatically by Diesel CLI.

diesel::table! {
    chats (chat_id) {
        chat_id -> Text,
        location -> Text,
        timezone -> Text,
    }
}

diesel::table! {
    quickshab (chat_id) {
        chat_id -> Text,
        number -> Integer,
    }
}

diesel::table! {
    quickshab_assignments (id) {
        id -> Integer,
        chat_id -> Text,
        category -> Text,
        name -> Text,
    }
}

diesel::table! {
    shopping_list (id) {
        id -> Integer,
        chat_id -> Text,
        item -> Text,
        quantity -> Integer,
    }
}

diesel::table! {
    reminders (id) {
        id -> Text,
        chat_id -> Text,
        message -> Text,
        time -> Integer,
        #[sql_name = "type"]
        type_ -> Text,
        snoozable -> Integer,
        sent_time -> Integer,
    }
}

diesel::joinable!(quickshab -> chats (chat_id));
diesel::joinable!(quickshab_assignments -> chats (chat_id));
diesel::joinable!(shopping_list -> chats (chat_id));
diesel::joinable!(reminders -> chats (chat_id));

diesel::allow_tables_to_appear_in_same_query!(
    chats,
    quickshab,
    quickshab_assignments,
    shopping_list,
    reminders,
);
