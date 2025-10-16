use chrono::Local;
use log::{error, info};
use qr2term::print_qr;
use wacore::proto_helpers::MessageExt;
use wacore::types::events::Event;
use waproto::whatsapp as wa;
use whatsapp_rust::bot::MessageContext;
// This is a demo of a simple ping-pong bot with every type of media.

mod commands;
mod db;
mod deno_client;
mod reminder_routine;
mod router;
mod session_cleanup;
mod utils;
fn main() {
    use std::sync::Arc;
    use whatsapp_rust::bot::Bot;
    use whatsapp_rust::store::sqlite_store::SqliteStore;

    env_logger::Builder::from_env(env_logger::Env::default().default_filter_or("error"))
        .format(|buf, record| {
            use std::io::Write;
            writeln!(
                buf,
                "{} [{:<5}] [{}] - {}",
                Local::now().format("%H:%M:%S"),
                record.level(),
                record.target(),
                record.args()
            )
        })
        .init();

    let rt = tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()
        .unwrap();

    rt.block_on(async {
        let backend = match SqliteStore::new("db/whatsapp.db").await {
            Ok(store) => Arc::new(store),
            Err(e) => {
                error!("Failed to create SQLite backend: {}", e);
                return;
            }
        };
        info!("SQLite backend initialized successfully.");

        let mut bot = Bot::builder()
            .with_backend(backend)
            // Optional: Override the WhatsApp version (normally auto-fetched)
            // .with_version((2, 3000, 1027868167))
            .on_event(move |event, client| {
                async move {
                    match event {
                        Event::PairingQrCode { code, timeout } => {
                            info!("----------------------------------------");
                            info!(
                                "New pairing code received (valid for {} seconds):",
                                timeout.as_secs()
                            );
                            print_qr(&code).unwrap_or_else(|e| {
                                error!("Failed to print QR code: {}", e);
                            });
                            info!("----------------------------------------");
                        }

                        Event::Message(msg, info) => {
                            let ctx = MessageContext {
                                message: msg,
                                info,
                                client,
                            };

                            if let Some(text) = ctx.message.text_content() {
                                if ctx.info.source.is_from_me {
                                    return;
                                }
                                let mut router =
                                    router::CommandRouter::new("db/shabbot.db").unwrap();
                                let replies = router.parse_multiple_commands(
                                    text,
                                    &ctx.info.source.chat.to_ad_string(),
                                    &ctx.info.source.sender.to_ad_string(),
                                );
                                for x in replies {
                                    let ext = wa::message::ExtendedTextMessage {
                                        text: Some(x.clone()),
                                        context_info: Some(Box::new(wa::ContextInfo {
                                            stanza_id: Some(ctx.info.id.clone()),
                                            participant: Some(
                                                ctx.info.source.sender.to_ad_string(),
                                            ),
                                            quoted_message: Some(Box::new((*ctx.message).clone())),
                                            ..Default::default()
                                        })),
                                        ..Default::default()
                                    };
                                    if let Err(e) = ctx
                                        .send_message(wa::Message {
                                            extended_text_message: Some(Box::new(ext)),
                                            ..Default::default()
                                        })
                                        .await
                                    {
                                        error!("Failed to send message: {}", e);
                                    }
                                }
                            }
                        }

                        Event::UndecryptableMessage(u) => {
                            error!("Decryption error from: {:?}", u);
                        }
                        Event::Connected(_) => {
                            info!("✅ Bot connected successfully!");
                        }
                        Event::Receipt(receipt) => {
                            info!(
                                "Got receipt for message(s) {:?}, type: {:?}",
                                receipt.message_ids, receipt.r#type
                            );
                        }
                        Event::LoggedOut(_) => {
                            error!("❌ Bot was logged out!");
                        }
                        _ => {
                            // debug!("Received unhandled event: {:?}", event);
                        }
                    }
                }
            })
            .build()
            .await
            .expect("Failed to build bot");

        // Get the client for the reminder routine before running the bot
        let client_for_reminders = bot.client().clone();

        // Spawn reminder routine in background with the client
        tokio::spawn(async move {
            reminder_routine::start_reminder_routine(
                client_for_reminders,
                "db/shabbot.db".to_string(),
            )
            .await;
        });
        info!("Reminder routine started in background");

        // Auto-reconnect loop with exponential backoff
        let mut reconnect_attempts = 0;
        const MAX_RECONNECT_ATTEMPTS: u32 = 10;
        const STABLE_CONNECTION_TIME: u64 = 300; // 5 minutes - reset counter after this

        loop {
            info!(
                "Starting bot connection (attempt {})",
                reconnect_attempts + 1
            );

            let connection_start = tokio::time::Instant::now();

            // Run the bot
            let bot_handle = match bot.run().await {
                Ok(handle) => {
                    info!("✅ Bot connected successfully!");
                    handle
                }
                Err(e) => {
                    error!("❌ Bot failed to start: {}", e);

                    // Exponential backoff: 2s, 4s, 8s, 16s, ... up to 60s
                    reconnect_attempts += 1;
                    if reconnect_attempts >= MAX_RECONNECT_ATTEMPTS {
                        error!(
                            "Max reconnection attempts ({}) reached. Giving up.",
                            MAX_RECONNECT_ATTEMPTS
                        );
                        return;
                    }

                    let delay = std::cmp::min(2u64.pow(reconnect_attempts), 60);
                    error!("Reconnecting in {} seconds...", delay);
                    tokio::time::sleep(tokio::time::Duration::from_secs(delay)).await;
                    continue;
                }
            };

            // Wait for the bot to disconnect
            match bot_handle.await {
                Ok(_) => {
                    info!("Bot disconnected normally");
                }
                Err(e) => {
                    error!("Bot connection error: {}", e);
                }
            }

            // Check connection uptime - reset counter if connection was stable
            let uptime = connection_start.elapsed().as_secs();
            if uptime >= STABLE_CONNECTION_TIME {
                info!(
                    "Connection was stable for {} seconds, resetting reconnect counter",
                    uptime
                );
                reconnect_attempts = 0;
            } else {
                reconnect_attempts += 1;
                info!(
                    "Connection only lasted {} seconds (needed {} for reset)",
                    uptime, STABLE_CONNECTION_TIME
                );
            }

            // Check if we've exceeded max attempts
            if reconnect_attempts >= MAX_RECONNECT_ATTEMPTS {
                error!(
                    "Max reconnection attempts ({}) reached. Giving up.",
                    MAX_RECONNECT_ATTEMPTS
                );
                return;
            }

            // Exponential backoff before reconnecting
            let delay = std::cmp::min(2u64.pow(reconnect_attempts), 60);
            info!(
                "Connection lost. Reconnecting in {} seconds (attempt {}/{})...",
                delay,
                reconnect_attempts + 1,
                MAX_RECONNECT_ATTEMPTS
            );
            tokio::time::sleep(tokio::time::Duration::from_secs(delay)).await;
        }
    });
}
