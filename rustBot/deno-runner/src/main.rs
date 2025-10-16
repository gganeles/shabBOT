use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::io::{self, BufRead, Write};

#[derive(Debug, Deserialize)]
#[serde(tag = "type")]
enum Request {
    #[serde(rename = "eval")]
    Eval { code: String },
    #[serde(rename = "run_file")]
    RunFile { path: String },
    #[serde(rename = "parse_date")]
    ParseDate {
        expression: String,
        reference_date: Option<String>,
    },
    #[serde(rename = "ping")]
    Ping,
}

#[derive(Debug, Serialize)]
#[serde(tag = "status")]
enum Response {
    #[serde(rename = "success")]
    Success { result: String },
    #[serde(rename = "error")]
    Error { message: String },
}

fn main() -> Result<()> {
    // Simple REPL-style communication via stdin/stdout
    // The WhatsApp bot will send JSON requests and receive JSON responses

    eprintln!("Deno runner started. Waiting for commands...");

    let stdin = io::stdin();
    let mut stdout = io::stdout();

    for line in stdin.lock().lines() {
        let line = line?;

        if line.trim().is_empty() {
            continue;
        }

        let response = match serde_json::from_str::<Request>(&line) {
            Ok(request) => handle_request(request),
            Err(e) => Response::Error {
                message: format!("Invalid JSON: {}", e),
            },
        };

        // Send response as JSON line
        serde_json::to_writer(&mut stdout, &response)?;
        stdout.write_all(b"\n")?;
        stdout.flush()?;
    }

    Ok(())
}

fn handle_request(request: Request) -> Response {
    match request {
        Request::Ping => Response::Success {
            result: "pong".to_string(),
        },
        Request::Eval { code } => {
            // TODO: Implement Deno evaluation
            // For now, just echo back
            Response::Success {
                result: format!("Would evaluate: {}", code),
            }
        }
        Request::RunFile { path } => {
            // TODO: Implement Deno file execution
            Response::Success {
                result: format!("Would run file: {}", path),
            }
        }
        Request::ParseDate {
            expression,
            reference_date,
        } => {
            // Call the chrono.ts module using deno command
            match parse_date_with_deno(&expression, reference_date.as_deref()) {
                Ok(result) => Response::Success { result },
                Err(e) => Response::Error {
                    message: format!("Date parsing error: {}", e),
                },
            }
        }
    }
}

fn parse_date_with_deno(expression: &str, reference_date: Option<&str>) -> Result<String> {
    use std::process::Command as StdCommand;

    // Build the inline JavaScript that imports and uses chrono.ts
    let ref_date_arg = reference_date.unwrap_or("null");
    let code = format!(
        r#"
import {{ parse }} from "./chrono.ts";
const results = parse("{}", {});
console.log(JSON.stringify(results));
"#,
        expression.replace('"', "\\\""),
        if ref_date_arg == "null" {
            "null".to_string()
        } else {
            format!("\"{}\"", ref_date_arg.replace('"', "\\\""))
        }
    );

    // Run deno with the inline code
    // Note: For npm imports, we need to use a temp file instead of eval
    use std::fs;
    use std::path::Path;

    let temp_file = Path::new("temp_parse.ts");
    fs::write(temp_file, &code)?;

    // Run deno with the temp file
    // Note: npm packages are cached after first download, so we only need --allow-read
    // To cache initially, run: deno cache chrono.ts
    let output = StdCommand::new("deno")
        .arg("run")
        .arg("--allow-read")
        .arg("temp_parse.ts")
        .current_dir(".")
        .output()?;

    // Clean up temp file
    let _ = fs::remove_file(temp_file);

    if output.status.success() {
        let result = String::from_utf8_lossy(&output.stdout).trim().to_string();
        Ok(result)
    } else {
        let error = String::from_utf8_lossy(&output.stderr);
        Err(anyhow::anyhow!("Deno execution failed: {}", error))
    }
}
