# AutoNotes Tool

This directory contains the core logic and CLI for the AutoNotes system.

## Project Structure

- `cmd/autonotes/`: The main CLI entry point.
- `parser.go`: Hand-written XML-like parser for `.note` files.
- `build.go`: Logic for converting `.note` files to HTML and processing images.
- `status.go`: Logic for identifying unprocessed images in the `src/` directory.
- `query.go`: Logic for querying content from transcribed notes.
- `serve.go`: Local development server with live-reload.
- `sync.go`: Utility for syncing data (if implemented).
- `check.go`: Utility for checking note consistency.
- `config.go`: Global configuration and flags.

## Package Usage

The code in this directory is part of the `autonotes` package. It is designed to be used by the CLI in `cmd/autonotes/` or potentially other tools.

### Key Components

- **Parser**: Use `autonotes.NewParser(content)` to create a parser and `Parse()` to get an AST of `Node` elements.
- **Commands**: Exported cobra commands like `StatusCmd`, `BuildCmd`, etc. are available for integration.
