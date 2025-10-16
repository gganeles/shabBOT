pub fn help_cmd(_prompt: &str) -> String {
    "Allow me to introduce myself!\n\n\
     i am the SHABbot!!\n\
     I can help you with you shabbat meals, as well as other things\n\n\
     The way it works:\n\n\
      - Use !quickshab to keep track of what everyone's bringing\n\
      - Use !shabtimes to find out when shabbat starts\n\
      - Use !remind to set reminders for yourself\n\n\
     For more information, type !docs to see the full documentation"
        .to_string()
}

pub fn docs_cmd() -> String {
    "github.com/gganeles/shabBOT".to_string()
}
