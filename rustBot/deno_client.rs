use anyhow::Result;
use log::{error, info};
use serde::{Deserialize, Serialize};
use std::io::{BufRead, BufReader, Write};
use std::process::{Child, Command, Stdio};

#[derive(Debug, Deserialize, Serialize)]
pub struct ParsedDate {
    #[serde(rename = "Text")]
    pub text: String,
    #[serde(rename = "Index")]
    pub index: i32,
    #[serde(rename = "Time")]
    pub time: Option<String>,
}

#[derive(Debug, Serialize)]
#[serde(tag = "type")]
pub enum DenoRequest {
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

#[derive(Debug, Deserialize)]
#[serde(tag = "status")]
pub enum DenoResponse {
    #[serde(rename = "success")]
    Success { result: String },
    #[serde(rename = "error")]
    Error { message: String },
}

pub struct DenoRunner {
    process: Child,
}

impl DenoRunner {
    /// Start the deno-runner binary as a child process
    pub fn new() -> Result<Self> {
        info!("Starting deno-runner subprocess...");

        let process = Command::new("./target/debug/deno-runner")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::inherit())
            .spawn()?;

        info!("Deno runner subprocess started with PID {}", process.id());

        Ok(Self { process })
    }

    /// Send a request to the deno-runner and get the response
    pub fn send_request(&mut self, request: DenoRequest) -> Result<DenoResponse> {
        let stdin = self
            .process
            .stdin
            .as_mut()
            .ok_or_else(|| anyhow::anyhow!("Failed to get stdin"))?;

        let stdout = self
            .process
            .stdout
            .as_mut()
            .ok_or_else(|| anyhow::anyhow!("Failed to get stdout"))?;

        // Send request as JSON line
        let request_json = serde_json::to_string(&request)?;
        writeln!(stdin, "{}", request_json)?;
        stdin.flush()?;

        // Read response as JSON line
        let mut reader = BufReader::new(stdout);
        let mut response_line = String::new();
        reader.read_line(&mut response_line)?;

        let response: DenoResponse = serde_json::from_str(&response_line)?;

        Ok(response)
    }

    /// Convenience method to evaluate JavaScript code
    pub fn eval(&mut self, code: &str) -> Result<String> {
        let request = DenoRequest::Eval {
            code: code.to_string(),
        };

        match self.send_request(request)? {
            DenoResponse::Success { result } => Ok(result),
            DenoResponse::Error { message } => Err(anyhow::anyhow!("Deno error: {}", message)),
        }
    }

    /// Convenience method to run a JavaScript file
    pub fn run_file(&mut self, path: &str) -> Result<String> {
        let request = DenoRequest::RunFile {
            path: path.to_string(),
        };

        match self.send_request(request)? {
            DenoResponse::Success { result } => Ok(result),
            DenoResponse::Error { message } => Err(anyhow::anyhow!("Deno error: {}", message)),
        }
    }

    /// Test if the runner is responsive
    pub fn ping(&mut self) -> Result<bool> {
        let request = DenoRequest::Ping;

        match self.send_request(request)? {
            DenoResponse::Success { result } => Ok(result == "pong"),
            DenoResponse::Error { message } => {
                error!("Ping failed: {}", message);
                Ok(false)
            }
        }
    }

    /// Parse natural language date/time expressions using chrono-node
    pub fn parse_date(
        &mut self,
        expression: &str,
        reference_date: Option<&str>,
    ) -> Result<Vec<ParsedDate>> {
        let request = DenoRequest::ParseDate {
            expression: expression.to_string(),
            reference_date: reference_date.map(|s| s.to_string()),
        };

        match self.send_request(request)? {
            DenoResponse::Success { result } => {
                // Parse the JSON result as Vec<ParsedDate>
                let parsed: Vec<ParsedDate> = serde_json::from_str(&result)?;
                Ok(parsed)
            }
            DenoResponse::Error { message } => Err(anyhow::anyhow!("Deno error: {}", message)),
        }
    }
}

impl Drop for DenoRunner {
    fn drop(&mut self) {
        info!("Shutting down deno-runner subprocess...");
        let _ = self.process.kill();
    }
}
