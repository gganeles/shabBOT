// Library interface for rust_shabbot
// This allows tests to access internal modules

pub mod commands;
pub mod db;
pub mod deno_client;
pub mod router;
pub mod utils;

// Re-export database and schema for backward compatibility
pub use db::{database, schema};
