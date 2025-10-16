# Go -> Rust Migration Progress

This file tracks the progress of porting the Go `shabBOT` commands to Rust.

Status:

- [ ] Inventory Go commands & router
- [ ] Scaffold Rust project & deps
- [ ] Port router to Rust
- [ ] Write utilities to parse datetime from command with chrono-node binaries
- [ ] Write utility function
- [ ] Port `send` command with &str.parse<DateTime>()
- [ ] Port `remind` command with &str.parse<DateTime>()
- [ ] Port remaining commands
- [ ] Integrate hebcal-rust
- [ ] Tests, build & smoke test

Mapping (initial):

- `commands/router.go` -> `rustBot/src/router.rs`
- `commands/*.go` -> `rustBot/src/commands/*.rs`
- Database schema -> `rustBot/migrations` or embedded setup in `main.rs`

Notes:

- Use `whatsapp-rust` evaluate available crates.
- `chrono-node` binaries will be called as child processes and JSON parsed via `serde_json`.
- `hebcal-rust` will be used for holiday/shabbat calculations.
