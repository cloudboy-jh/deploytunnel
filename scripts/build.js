#!/usr/bin/env node

import { spawn } from "child_process";
import { platform } from "os";
import { join, dirname } from "path";
import { fileURLToPath } from "url";
import { mkdirSync, existsSync } from "fs";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const rootDir = join(__dirname, "..");
const binDir = join(rootDir, "bin");

// Determine output binary name based on platform
const getBinaryName = () => {
  const platformName = platform();
  if (platformName === "win32") {
    return "dt.exe";
  }
  return "dt";
};

// Check if Go is installed
const checkGo = () => {
  return new Promise((resolve) => {
    const child = spawn("go", ["version"], { stdio: "pipe" });
    child.on("error", () => resolve(false));
    child.on("exit", (code) => resolve(code === 0));
  });
};

// Build the Go binary
const build = async () => {
  console.log("Building Deploy Tunnel binary...");
  console.log("");

  // Check if Go is installed
  const hasGo = await checkGo();
  if (!hasGo) {
    console.error("Error: Go is not installed or not in PATH");
    console.error("");
    console.error("Deploy Tunnel requires Go to build the binary.");
    console.error("Please install Go from: https://go.dev/doc/install");
    console.error("");
    console.error("After installing Go, run: npm install -g deploy-tunnel");
    console.error("");
    process.exit(1);
  }

  // Ensure bin directory exists
  if (!existsSync(binDir)) {
    mkdirSync(binDir, { recursive: true });
  }

  const binaryName = getBinaryName();
  const outputPath = join(binDir, binaryName);

  console.log(`Platform: ${platform()}`);
  console.log(`Output: ${outputPath}`);
  console.log("");

  // Build the Go binary
  const buildArgs = ["build", "-o", outputPath, "./cmd/deploy-tunnel"];

  const child = spawn("go", buildArgs, {
    cwd: rootDir,
    stdio: "inherit",
    env: {
      ...process.env,
      CGO_ENABLED: "1", // Required for SQLite
    },
  });

  child.on("exit", (code) => {
    if (code === 0) {
      console.log("");
      console.log("Deploy Tunnel built successfully!");
      console.log("");
      console.log("Run: dt --help");
      console.log("");
      process.exit(0);
    } else {
      console.error("");
      console.error("Build failed with exit code:", code);
      console.error("");
      process.exit(code || 1);
    }
  });

  child.on("error", (err) => {
    console.error("");
    console.error("Build error:", err.message);
    console.error("");
    process.exit(1);
  });
};

build();
