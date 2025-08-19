# Installation Guide

certwiz can be installed in several ways depending on your needs and environment.

## Prerequisites

- Terminal with UTF-8 support (for emoji and special characters)
- macOS, Linux, FreeBSD, or Windows
- `curl` and `tar` (for automatic installation)
- Go 1.20 or higher (only for building from source)

## Installation Methods

### 1. Quick Install Script (Recommended) 🚀

The easiest way to install certwiz is using our automatic installer script. It detects your OS and architecture, downloads the appropriate binary, and installs it for you.

```bash
# Install latest version
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash
```

#### Installation Options

```bash
# Install a specific version
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash -s -- --version v0.1.0

# Install to a custom directory (e.g., for non-root users)
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash -s -- --install-dir $HOME/.local/bin

# See all options
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash -s -- --help
```

The installer:
- ✅ Automatically detects your OS (macOS, Linux, FreeBSD)
- ✅ Automatically detects your architecture (amd64, arm64, 386)
- ✅ Downloads the correct binary from GitHub releases
- ✅ Installs to `/usr/local/bin` (or custom directory)
- ✅ Verifies the installation
- ✅ Provides PATH configuration help if needed

### 2. Manual Download

Download pre-built binaries from the [releases page](https://github.com/trahma/certwiz/releases).

#### macOS (Apple Silicon - M1/M2/M3)
```bash
curl -L https://github.com/trahma/certwiz/releases/latest/download/cert-darwin-arm64.tar.gz | tar xz
sudo mv cert-darwin-arm64 /usr/local/bin/cert
chmod +x /usr/local/bin/cert
```

#### macOS (Intel)
```bash
curl -L https://github.com/trahma/certwiz/releases/latest/download/cert-darwin-amd64.tar.gz | tar xz
sudo mv cert-darwin-amd64 /usr/local/bin/cert
chmod +x /usr/local/bin/cert
```

#### Linux (x86_64)
```bash
curl -L https://github.com/trahma/certwiz/releases/latest/download/cert-linux-amd64.tar.gz | tar xz
sudo mv cert-linux-amd64 /usr/local/bin/cert
chmod +x /usr/local/bin/cert
```

#### Linux (ARM64)
```bash
curl -L https://github.com/trahma/certwiz/releases/latest/download/cert-linux-arm64.tar.gz | tar xz
sudo mv cert-linux-arm64 /usr/local/bin/cert
chmod +x /usr/local/bin/cert
```

#### Windows
Download the appropriate `.zip` file from the [releases page](https://github.com/trahma/certwiz/releases):
- `cert-windows-amd64.zip` for 64-bit systems
- `cert-windows-arm64.zip` for ARM64 systems
- `cert-windows-386.zip` for 32-bit systems

Extract and add `cert.exe` to your PATH.

### 3. Install with Go

If you have Go 1.20+ installed:

```bash
go install github.com/trahma/certwiz@latest
```

This installs the `cert` binary to your `$GOPATH/bin` directory.

### 4. Build from Source

Clone the repository and build:

```bash
git clone https://github.com/trahma/certwiz
cd certwiz
make build
```

Or using Go directly:

```bash
git clone https://github.com/trahma/certwiz
cd certwiz
go build -o cert .
```

## Verification

After installation, verify certwiz is working:

```bash
# Check version
cert version

# Test with a simple command
cert inspect google.com
```

## Shell Completion

cert supports shell completion for bash, zsh, fish, and PowerShell.

### Bash
```bash
cert completion bash > /etc/bash_completion.d/cert
# Or for user installation:
cert completion bash > ~/.bash_completion
```

### Zsh
```bash
cert completion zsh > "${fpath[1]}/_cert"
# Or add to your .zshrc:
source <(cert completion zsh)
```

### Fish
```bash
cert completion fish > ~/.config/fish/completions/cert.fish
```

### PowerShell
```powershell
cert completion powershell | Out-String | Invoke-Expression
# Or add to your profile:
cert completion powershell >> $PROFILE
```

## Updating

### Using the installer script
```bash
# Update to latest version
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash

# Update to specific version
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash -s -- --version v0.2.0
```

### If installed with Go
```bash
go install github.com/trahma/certwiz@latest
```

### If built from source
```bash
cd certwiz
git pull
make clean build
```

## Uninstalling

### If installed with the installer script or manually
```bash
sudo rm /usr/local/bin/cert
# Or from custom location:
rm $HOME/.local/bin/cert
```

### If installed with Go
```bash
rm $(go env GOPATH)/bin/cert
```

## Troubleshooting

### Command not found

If you get "command not found" after installation:

1. Check if cert is in your PATH:
   ```bash
   which cert
   ```

2. If installed to a custom directory, add it to PATH:
   ```bash
   export PATH=$PATH:$HOME/.local/bin
   ```
   Add this line to your shell configuration file (`~/.bashrc`, `~/.zshrc`, etc.)

3. If using Go, ensure `$GOPATH/bin` is in your PATH:
   ```bash
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

### Permission denied

If you get permission errors when installing to `/usr/local/bin`:
- Use the installer script which handles sudo automatically
- Or use a user directory:
  ```bash
  curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash -s -- --install-dir $HOME/.local/bin
  ```

### Colors not displaying

If colors aren't showing properly:
- Ensure your terminal supports 256 colors
- Try setting: `export TERM=xterm-256color`
- On Windows, use Windows Terminal or PowerShell 7+

## Next Steps

- Read the [Usage Guide](usage.md) to learn basic commands
- Check out [Examples](examples.md) for real-world scenarios
- See the [Command Reference](commands.md) for detailed options