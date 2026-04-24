# envoy-cli

> A CLI tool for managing and syncing `.env` files across local, staging, and production environments using encrypted vaults.

---

## Installation

**Using Go:**
```bash
go install github.com/yourname/envoy-cli@latest
```

**Using Homebrew:**
```bash
brew install yourname/tap/envoy-cli
```

---

## Usage

Initialize a new encrypted vault in your project:
```bash
envoy init
```

Push your local `.env` file to the staging environment:
```bash
envoy push --env staging
```

Pull the production environment variables to your local machine:
```bash
envoy pull --env production
```

Sync variables across all environments:
```bash
envoy sync --all
```

List available environments and their sync status:
```bash
envoy status
```

---

## How It Works

`envoy-cli` encrypts your `.env` files using AES-256 before storing them in a remote vault. Each environment (`local`, `staging`, `production`) is managed as a separate encrypted snapshot, allowing teams to safely share and sync secrets without exposing plaintext credentials.

---

## Requirements

- Go 1.21+
- A configured vault backend (local filesystem, S3, or Vault by HashiCorp)

---

## License

[MIT](LICENSE) © 2024 yourname