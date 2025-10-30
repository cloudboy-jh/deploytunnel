#!/usr/bin/env node

import { spawn } from 'child_process';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { platform } from 'os';
import { existsSync } from 'fs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Determine the binary name based on platform
const getBinaryName = () => {
  const platformName = platform();
  if (platformName === 'win32') {
    return 'dt.exe';
  }
  return 'dt';
};

// Find the compiled Go binary
const findBinary = () => {
  const binaryName = getBinaryName();

  // Try multiple possible locations
  const possiblePaths = [
    join(__dirname, '..', 'bin', binaryName),           // npm global install
    join(__dirname, '..', 'cmd', 'deploy-tunnel', binaryName), // development
    join(__dirname, '..', binaryName),                  // root of package
  ];

  for (const path of possiblePaths) {
    if (existsSync(path)) {
      return path;
    }
  }

  return null;
};

// Main execution
const main = () => {
  const binaryPath = findBinary();

  if (!binaryPath) {
    console.error('Error: Deploy Tunnel binary not found.');
    console.error('');
    console.error('This usually means the Go binary was not compiled during installation.');
    console.error('');
    console.error('To fix this, try:');
    console.error('  1. Ensure Go is installed: https://go.dev/doc/install');
    console.error('  2. Reinstall: npm install -g deploy-tunnel');
    console.error('  3. Or build manually: go build -o bin/dt ./cmd/deploy-tunnel');
    console.error('');
    process.exit(1);
  }

  // Spawn the Go binary with all arguments
  const child = spawn(binaryPath, process.argv.slice(2), {
    stdio: 'inherit',
    env: process.env,
  });

  // Handle exit
  child.on('exit', (code) => {
    process.exit(code || 0);
  });

  // Handle errors
  child.on('error', (err) => {
    console.error('Error executing Deploy Tunnel:', err.message);
    process.exit(1);
  });
};

main();
