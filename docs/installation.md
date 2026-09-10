# Installation Guide

certwiz can be installed in several ways depending on your needs and environment. The installed command is named `cert`.

## Prerequisites

- macOS, Linux, FreeBSD, or Windows
- A terminal with UTF-8 support for the default bordered output (use `--plain` otherwise)
- `curl` and `tar` (for the installer script and manual downloads)
- Go 1.20 or higher (only for building from source)

## Installation Methods

### 1. Quick Install Script (Recommended)

The installer detects your OS and architecture, downloads the matching release archive, and installs the `cert` binary.

```bash
# Install the latest version
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash
```

#### Installation Options

```bash
# Install a specific version
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash -s -- --version v0.4.1

# Install to a custom directory (e.g., for non-root users)
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash -s -- --install-dir $HOME/.local/bin

# See all options
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash -s -- --help
```

The installer:
- Detects your OS (macOS, Linux, FreeBSD)
- Detects your architecture (x86_64, arm64, armv7, i386)
- Downloads the matching archive from GitHub releases
- Installs to `/usr/local/bin` (or the directory given with `--install-dir`)
- Verifies the installation
- Explains how to update your PATH if needed

### 2. Manual Download

Download a pre-built archive from the [releases page](https://github.com/trahma/certwiz/releases). Archives are named `cert-<os>-<arch>.tar.gz` (`.zip` on Windows) and each contains a single binary named `cert` (or `cert.exe`) alongside the README, CHANGELOG, and LICENSE.

Available builds:

| OS | Architectures |
|----|---------------|
| macOS (`darwin`) | `arm64` (Apple Silicon), `x86_64` (Intel) |
| Linux | `x86_64`, `arm64`, `armv7`, `i386` |
| FreeBSD | `x86_64`, `arm64` |
| Windows | `x86_64`, `arm64`, `i386` |

#### macOS (Apple Silicon)
```bash
curl -L https://github.com/trahma/certwiz/releases/latest/download/cert-darwin-arm64.tar.gz | tar xz cert
sudo mv cert /usr/local/bin/cert
```

#### macOS (Intel)
```bash
curl -L https://github.com/trahma/certwiz/releases/latest/download/cert-darwin-x86_64.tar.gz | tar xz cert
sudo mv cert /usr/local/bin/cert
```

#### Linux (x86_64)
```bash
curl -L https://github.com/trahma/certwiz/releases/latest/download/cert-linux-x86_64.tar.gz | tar xz cert
sudo mv cert /usr/local/bin/cert
```

#### Linux (arm64)
```bash
curl -L https://github.com/trahma/certwiz/releases/latest/download/cert-linux-arm64.tar.gz | tar xz cert
sudo mv cert /usr/local/bin/cert
```

For `armv7` or `i386`, substitute the architecture in the file name.

#### FreeBSD
```bash
curl -L https://github.com/trahma/certwiz/releases/latest/download/cert-freebsd-x86_64.tar.gz | tar xz cert
sudo mv cert /usr/local/bin/cert
```

#### Windows
Download the appropriate `.zip` file from the [releases page](https://github.com/trahma/certwiz/releases):
- `cert-windows-x86_64.zip` for 64-bit systems
- `cert-windows-arm64.zip` for ARM64 systems
- `cert-windows-i386.zip` for 32-bit systems

Extract `cert.exe` and add its directory to your PATH. Note that `cert update` is not available on Windows; download a new release to upgrade.

#### Verifying a Download

Every release publishes a `checksums.txt` with SHA-256 sums for all archives:

```bash
VERSION=v0.4.1
curl -LO https://github.com/trahma/certwiz/releases/download/$VERSION/cert-darwin-arm64.tar.gz
curl -LO https://github.com/trahma/certwiz/releases/download/$VERSION/checksums.txt
shasum -a 256 --ignore-missing -c checksums.txt   # on Linux: sha256sum --ignore-missing -c checksums.txt
```

### 3. Build from Source

`go install` is not supported: the module is declared as `certwiz` in `go.mod`, not as its GitHub path. Clone and build instead:

```bash
git clone https://github.com/trahma/certwiz
cd certwiz
make build          # produces ./cert
make install        # installs to $GOPATH/bin
```

Or with Go directly:

```bash
go build -o cert .
```

## Verification

After installation, check that `cert` is on your PATH and reports its version:

```bash
cert version
# cert version 0.4.1

# Try a simple command
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

### Automatic Update (macOS, Linux, FreeBSD)

```bash
# Download the installer and upgrade in place if a newer release exists
cert update

# Reinstall even if already on the latest version
cert update --force
```

The command downloads the installer script to a temporary file, runs it, and removes the file when the installer exits. The installer compares the latest release with your current version and upgrades your existing installation in place.

### Using the installer script
```bash
# Update to latest version
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash

# Update to a specific version
curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash -s -- --version v0.4.1
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
# Or from a custom location:
rm $HOME/.local/bin/cert
```

### If installed with `make install`
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

3. If installed with `make install`, ensure `$GOPATH/bin` is in your PATH:
   ```bash
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

### Permission denied

If you get permission errors when installing to `/usr/local/bin`:
- Use the installer script, which handles sudo automatically
- Or use a user directory:
  ```bash
  curl -sSL https://raw.githubusercontent.com/trahma/certwiz/main/install.sh | bash -s -- --install-dir $HOME/.local/bin
  ```

### Colors or borders not displaying

- Ensure your terminal supports 256 colors and UTF-8
- Try setting: `export TERM=xterm-256color`
- Use `--plain` for output without borders, colors, or symbols
- On Windows, use Windows Terminal or PowerShell 7+

## Next Steps

- Read the [Usage Guide](usage.md) to learn basic commands
- Check out [Examples](examples.md) for real-world scenarios
- See the [Command Reference](commands.md) for detailed options
