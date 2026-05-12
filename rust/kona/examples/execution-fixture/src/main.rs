//! Example for creating a static test fixture for `kona-executor` from a live chain
//!
//! ## Usage
//!
//! ```sh
//! cargo run --release -p execution-fixture
//! ```
//!
//! ## Inputs
//!
//! The test fixture creator takes the following inputs:
//!
//! - `-v` or `--verbosity`: Verbosity level (0-2)
//! - `-r` or `--l2-rpc`: The L2 execution layer RPC URL to use. Must be archival.
//! - `-b` or `--block-number`: L2 block number to execute for the fixture.
//! - `-o` or `--output-dir`: (Optional) The output directory for the fixture. If not provided,
//!   defaults to `kona-executor`'s `testdata` directory.

use anyhow::{Result, anyhow};
use clap::Parser;
use kona_cli::{LogArgs, LogConfig};
use kona_executor::test_utils::ExecutorTestFixtureCreator;
use std::path::PathBuf;
use tracing::{error, info, warn};
use tracing_subscriber::EnvFilter;
use url::Url;

/// The execution fixture creation command.
#[derive(Parser, Debug, Clone)]
#[command(about = "Creates a static test fixture for `kona-executor` from a live chain")]
pub struct ExecutionFixtureCommand {
    #[command(flatten)]
    pub v: LogArgs,
    /// The L2 archive EL to use.
    #[arg(long, short = 'r')]
    pub l2_rpc: Url,
    /// L2 block number to execute.
    #[arg(long, short = 'b')]
    pub block_number: u64,
    /// The output directory for the fixture.
    #[arg(long, short = 'o')]
    pub output_dir: Option<PathBuf>,
    /// Number of blocks to process (default: 1)
    #[arg(long, default_value = "1")]
    pub block_count: u64,
    /// Skip saving data to disk (use temporary directory)
    #[arg(long, default_value = "false")]
    pub skip_save: bool,
}

/// Execution statistics tracker
#[derive(Debug, Default)]
struct BlockExecutionStats {
    success_count: u64,
    failure_count: u64,
    failed_blocks: Vec<u64>,
}

impl BlockExecutionStats {
    fn new() -> Self {
        Self::default()
    }

    fn record_success(&mut self) {
        self.success_count += 1;
    }

    fn record_failure(&mut self, block_number: u64) {
        self.failure_count += 1;
        self.failed_blocks.push(block_number);
    }

    fn print_summary(&self) {
        let total = self.success_count + self.failure_count;
        if total == 0 {
            info!("No blocks were processed");
            return;
        }

        let success_percent = (self.success_count as f64 / total as f64) * 100.0;
        let failure_percent = (self.failure_count as f64 / total as f64) * 100.0;

        // Print summary statistics
        println!("\n╔════════════════════════════════════════════════════════════════╗");
        println!("║                  📊 Block Execution Summary                   ║");
        println!("╠════════════════════════════════════════════════════════════════╣");
        println!("║  Total Blocks: {:<47}  ║", total);
        println!(
            "║  ✅ Success: {:<6} ({:.1}%)                                   ║",
            self.success_count, success_percent
        );
        println!(
            "║  ❌ Failed: {:<6} ({:.1}%)                                    ║",
            self.failure_count, failure_percent
        );
        println!("╚════════════════════════════════════════════════════════════════╝");

        // Print failed blocks
        if !self.failed_blocks.is_empty() {
            println!("\n╔════════════════════════════════════════════════════════════════╗");
            println!("║                    📋 Failed Block Details                    ║");
            println!("╠═══════════════════╦══════════════════════════════════════════╣");
            println!("║   Block Number    ║               Explorer Link               ║");
            println!("╠═══════════════════╬══════════════════════════════════════════╣");

            for block_num in &self.failed_blocks {
                println!(
                    "║  {:<16} ║  https://explorer.mantle.xyz/block/{}?tab=txs  ║",
                    block_num, block_num
                );
            }
            println!("╚═══════════════════╩══════════════════════════════════════════╝");
        }

        println!("\n🏁 Execution Completed!");
    }
}

#[tokio::main]
async fn main() -> Result<()> {
    let cli = ExecutionFixtureCommand::parse();
    LogConfig::new(cli.v).init_tracing_subscriber(None::<EnvFilter>)?;

    let output_dir = if let Some(output_dir) = cli.output_dir {
        output_dir
    } else {
        // Default to `crates/proof/executor/testdata`
        let output = std::process::Command::new(env!("CARGO"))
            .arg("locate-project")
            .arg("--workspace")
            .arg("--message-format=plain")
            .output()?
            .stdout;
        let workspace_root: PathBuf = String::from_utf8(output)?.trim().into();

        workspace_root
            .parent()
            .ok_or(anyhow!("Failed to locate workspace root"))?
            .join("crates/proof/executor/testdata")
    };

    let mut stats = BlockExecutionStats::new();

    info!(
        "Starting block processing from block {} for {} blocks",
        cli.block_number, cli.block_count
    );

    for i in 0..cli.block_count {
        let current_block = cli.block_number + i;
        let fixture_creator = ExecutorTestFixtureCreator::new_with_options(
            cli.l2_rpc.as_str(),
            current_block,
            output_dir.clone(),
            cli.skip_save,
        );

        info!(block_number = current_block, "Processing block");

        match fixture_creator.create_static_fixture().await {
            Ok(success) => {
                if success {
                    stats.record_success();
                    info!(block_number = current_block, "Block execution succeeded");
                } else {
                    stats.record_failure(current_block);
                    warn!(block_number = current_block, "Block execution failed");
                }
            }
            Err(_) => {
                stats.record_failure(current_block);
                error!(block_number = current_block, "Block execution error");
            }
        }
    }

    stats.print_summary();
    Ok(())
}
