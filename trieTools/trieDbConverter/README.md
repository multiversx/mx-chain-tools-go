

# trieDbConverter — usage

trieDbConverter is a command-line tool that copies a trie (and any related data tries) from a MultiversX trie database layout found under a working directory.

This README contains only usage information (flags, defaults and examples). It does not describe implementation details.

Prerequisites
- A working directory containing the database layout. The tool resolves DB locations relative to the working directory (see `--working-directory` and `--db-directory`).
- The `AccountsTrie` layout for the shard you intend to copy must exist under the resolved location.
- You must provide the main trie root hash as a 32-byte hex string via `--hex-roothash`.

Flags (common)
- `--working-directory` (string, required): Base path where the tool will look for DB directories and write logs.
- `--db-directory` (string, default: `db`): Subdirectory under the working directory that contains the chain DB layout.
- `--hex-roothash` (string, required): Hex-encoded 32-byte root hash of the main trie to copy.
- `--origin-db-type` (string, default: `static`): Type of the origin DB. Supported values: `static`, `pruning`.
- `--target-db-type` (string, default: `static`): Type of the target DB. Supported values: `static`, `pruning`.
- `--max-goroutines` (int, default: `100`): Maximum number of goroutines used for copying (main trie + data tries).
- Logging / runtime flags are also available (see `-h`): `--log-level`, `--log-save`, `--disable-ansi-color`, `--profile-mode`, `--log-logger-name`.

Safety notes
- The tool will refuse to open the same DB file path for both origin and target. If the resolved origin and target DB file paths are identical the tool exits with an error to avoid corrupting LevelDB files.

Examples

Run the converter from the repository folder using `go run`:

```bash
cd trieTools/trieDbConverter
go run . \
  --working-directory /data/trie-converter \
  --db-directory db \
  --hex-roothash 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
```

Specify a different goroutine limit:

```bash
go run . --working-directory /data/trie-converter --db-directory db --hex-roothash <root> --max-goroutines 50
```

Get full flag help from the binary:

```bash
go run . -h
```

Troubleshooting (brief)
- Verify the provided `--hex-roothash` is valid hex and exactly 32 bytes long.
- Ensure the working directory contains the expected `AccountsTrie` layout under the `--db-directory` path.
- If the tool errors about identical DB paths, provide distinct origin/target layouts under the working directory.

Exit codes
- 0 on success
- Non-zero on error; consult logs for details (use `--log-level` / `--log-save` to increase verbosity and persist logs).

For more examples or integration usage run the compiled binary with `-h` or inspect the `cmd/` directory.

