# Installation Guide

certwiz can be installed in several ways depending on your needs and environment.

## Prerequisites

- Go 1.20 or higher (for building from source)
- Terminal with UTF-8 support (for emoji and special characters)
- macOS, Linux, or Windows

## Installation Methods

### 1. Install with Go (Recommended)

If you have Go installed, this is the easiest method:

```bash
go install github.com/certwiz/certwiz@latest
```

This will install certwiz to your `$GOPATH/bin` directory. Make sure this directory is in your PATH.

### 2. Download Pre-built Binary

Download the latest binary for your platform from the [releases page](https://github.com/certwiz/certwiz/releases).

#### macOS (Intel)
```bash
curl -L https://github.com/certwiz/certwiz/releases/latest/download/certwiz-darwin-amd64 -o certwiz
chmod +x certwiz
sudo mv certwiz /usr/local/bin/
```

#### macOS (Apple Silicon)
```bash
curl -L https://github.com/certwiz/certwiz/releases/latest/download/certwiz-darwin-arm64 -o certwiz
chmod +x certwiz
sudo mv certwiz /usr/local/bin/
```

#### Linux (x86_64)
```bash
curl -L https://github.com/certwiz/certwiz/releases/latest/download/certwiz-linux-amd64 -o certwiz
chmod +x certwiz
sudo mv certwiz /usr/local/bin/
```

#### Windows
Download the `.exe` file from the releases page and add it to your PATH.

### 3. Build from Source

Clone the repository and build:

```bash
git clone https://github.com/certwiz/certwiz
cd certwiz
make build
```

Or using Go directly:

```bash
git clone https://github.com/certwiz/certwiz
cd certwiz
go build -o certwiz .
```

### 4. Using Homebrew (macOS/Linux)

```bash
brew tap certwiz/tap
brew install certwiz
```

### 5. Using Docker

```bash
docker pull certwiz/certwiz:latest
docker run --rm certwiz/certwiz inspect google.com
```

Create an alias for convenience:
```bash
alias certwiz='docker run --rm -v $(pwd):/work -w /work certwiz/certwiz'
```

## Verification

After installation, verify certwiz is working:

```bash
# Check version
certwiz --version

# Test with a simple command
certwiz inspect google.com
```

## Shell Completion

certwiz supports shell completion for bash, zsh, fish, and PowerShell.

### Bash
```bash
certwiz completion bash > /etc/bash_completion.d/certwiz
```

### Zsh
```bash
certwiz completion zsh > "${fpath[1]}/_certwiz"
```

### Fish
```bash
certwiz completion fish > ~/.config/fish/completions/certwiz.fish
```

### PowerShell
```powershell
certwiz completion powershell | Out-String | Invoke-Expression
```

## Updating

### If installed with Go
```bash
go install github.com/certwiz/certwiz@latest
```

### If installed with Homebrew
```bash
brew upgrade certwiz
```

### If built from source
```bash
cd certwiz
git pull
make clean build
```

## Uninstalling

### If installed with Go
```bash
rm $(go env GOPATH)/bin/certwiz
```

### If installed with Homebrew
```bash
brew uninstall certwiz
```

### If installed manually
```bash
rm /usr/local/bin/certwiz
```

## Troubleshooting

### Command not found

If you get "command not found" after installation:

1. Check if certwiz is in your PATH:
   ```bash
   which certwiz
   ```

2. If using Go, ensure `$GOPATH/bin` is in your PATH:
   ```bash
   export PATH=$PATH:$(go env GOPATH)/bin
   ```
   Add this line to your shell configuration file (`~/.bashrc`, `~/.zshrc`, etc.)

### Permission denied

If you get permission errors when installing to `/usr/local/bin`:
```bash
sudo mv certwiz /usr/local/bin/
```

Or install to a user directory:
```bash
mkdir -p ~/.local/bin
mv certwiz ~/.local/bin/
export PATH=$PATH:~/.local/bin
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