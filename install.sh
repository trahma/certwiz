#!/usr/bin/env bash

# CertWiz Installer Script
# This script automatically detects your OS and architecture,
# downloads the appropriate binary, and installs it to your system.

set -e

# Configuration
REPO_OWNER="trahma"
REPO_NAME="certwiz"
BINARY_NAME="cert"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Detect OS
detect_os() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "${OS}" in
        linux*)     OS='linux';;
        darwin*)    OS='darwin';;
        freebsd*)   OS='freebsd';;
        mingw*|msys*|cygwin*)     
            error "Windows detected. Please download the Windows binary manually from:"
            error "https://github.com/${REPO_OWNER}/${REPO_NAME}/releases"
            exit 1
            ;;
        *)          
            error "Unsupported OS: ${OS}"
            exit 1
            ;;
    esac
    echo "${OS}"
}

# Detect architecture
detect_arch() {
    ARCH="$(uname -m)"
    case "${ARCH}" in
        x86_64|amd64)           ARCH='amd64';;
        aarch64|arm64)          ARCH='arm64';;
        i386|i686)              ARCH='386';;
        armv7l|armv7|arm)       ARCH='arm';;
        *)                      
            error "Unsupported architecture: ${ARCH}"
            exit 1
            ;;
    esac
    
    # Special case for macOS M1/M2/M3
    if [[ "${OS}" == "darwin" ]] && [[ "$(sysctl -n hw.optional.arm64 2>/dev/null)" == "1" ]]; then
        ARCH='arm64'
    fi
    
    echo "${ARCH}"
}

# Get the latest release version
get_latest_version() {
    local version
    version=$(curl -s "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest" | \
              grep '"tag_name":' | \
              sed -E 's/.*"([^"]+)".*/\1/')
    
    if [[ -z "${version}" ]]; then
        error "Failed to fetch latest version"
        exit 1
    fi
    
    echo "${version}"
}

# Download binary
download_binary() {
    local version="$1"
    local os="$2"
    local arch="$3"
    local temp_dir="$(mktemp -d)"
    
    # Construct download URL
    local binary_name="${BINARY_NAME}-${os}-${arch}"
    local archive_name="${binary_name}.tar.gz"
    local download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${version}/${archive_name}"
    
    info "Downloading ${BINARY_NAME} ${version} for ${os}/${arch}..."
    info "URL: ${download_url}"
    
    # Download the archive
    if ! curl -L --fail --progress-bar -o "${temp_dir}/${archive_name}" "${download_url}"; then
        error "Failed to download binary"
        error "Please check if the release exists for your platform: ${os}/${arch}"
        rm -rf "${temp_dir}"
        exit 1
    fi
    
    # Extract the archive
    info "Extracting archive..."
    if ! tar -xzf "${temp_dir}/${archive_name}" -C "${temp_dir}"; then
        error "Failed to extract archive"
        rm -rf "${temp_dir}"
        exit 1
    fi
    
    # Find the binary (it should be named cert-OS-ARCH in the archive)
    local binary_path="${temp_dir}/${binary_name}"
    if [[ ! -f "${binary_path}" ]]; then
        # Try just the binary name
        binary_path="${temp_dir}/${BINARY_NAME}"
        if [[ ! -f "${binary_path}" ]]; then
            # List what's in the temp dir for debugging
            info "Looking for binary in: ${temp_dir}"
            ls -la "${temp_dir}" >&2
            error "Binary not found in archive"
            rm -rf "${temp_dir}"
            exit 1
        fi
    fi
    
    echo "${binary_path}"
}

# Install binary
install_binary() {
    local binary_path="$1"
    local install_path="${INSTALL_DIR}/${BINARY_NAME}"
    
    # Check if we need sudo
    local sudo_cmd=""
    if [[ ! -w "${INSTALL_DIR}" ]]; then
        if command -v sudo >/dev/null 2>&1; then
            sudo_cmd="sudo"
            info "Installation requires sudo privileges..."
        else
            error "Cannot write to ${INSTALL_DIR} and sudo is not available"
            exit 1
        fi
    fi
    
    # Create install directory if it doesn't exist
    if [[ ! -d "${INSTALL_DIR}" ]]; then
        info "Creating installation directory: ${INSTALL_DIR}"
        ${sudo_cmd} mkdir -p "${INSTALL_DIR}"
    fi
    
    # Copy binary to installation directory
    info "Installing ${BINARY_NAME} to ${install_path}..."
    ${sudo_cmd} cp "${binary_path}" "${install_path}"
    
    # Make binary executable
    ${sudo_cmd} chmod +x "${install_path}"
    
    # Clean up temp directory
    rm -rf "$(dirname "${binary_path}")"
}

# Verify installation
verify_installation() {
    local install_path="${INSTALL_DIR}/${BINARY_NAME}"
    
    if [[ ! -f "${install_path}" ]]; then
        error "Installation verification failed: binary not found"
        exit 1
    fi
    
    # Check if binary is in PATH
    if command -v "${BINARY_NAME}" >/dev/null 2>&1; then
        local installed_version="$(${BINARY_NAME} version 2>/dev/null || echo "unknown")"
        success "${BINARY_NAME} has been installed successfully!"
        info "Version: ${installed_version}"
        info "Location: $(which ${BINARY_NAME})"
    else
        warning "${BINARY_NAME} has been installed to ${install_path}"
        warning "However, ${INSTALL_DIR} is not in your PATH"
        echo ""
        info "Add ${INSTALL_DIR} to your PATH by adding this to your shell profile:"
        echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        echo ""
        info "Then reload your shell configuration or start a new terminal session."
    fi
}

# Main installation flow
main() {
    echo "================================================"
    echo "     CertWiz Installation Script"
    echo "================================================"
    echo ""
    
    # Check dependencies
    info "Checking dependencies..."
    for cmd in curl tar; do
        if ! command -v "${cmd}" >/dev/null 2>&1; then
            error "Required command '${cmd}' not found. Please install it first."
            exit 1
        fi
    done
    
    # Detect system
    info "Detecting system..."
    OS="$(detect_os)"
    ARCH="$(detect_arch)"
    success "Detected: ${OS}/${ARCH}"
    
    # Get version
    if [[ -n "${VERSION}" ]]; then
        info "Using specified version: ${VERSION}"
    else
        info "Fetching latest version..."
        VERSION="$(get_latest_version)"
        success "Latest version: ${VERSION}"
    fi
    
    # Download binary
    BINARY_PATH="$(download_binary "${VERSION}" "${OS}" "${ARCH}")"
    success "Download complete!"
    
    # Install binary
    install_binary "${BINARY_PATH}"
    
    # Verify installation
    echo ""
    verify_installation
    
    echo ""
    echo "================================================"
    echo "     Installation Complete!"
    echo "================================================"
    echo ""
    echo "Get started with:"
    echo "  ${BINARY_NAME} help"
    echo "  ${BINARY_NAME} version"
    echo "  ${BINARY_NAME} inspect google.com"
    echo ""
}

# Handle script arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --help)
            echo "CertWiz Installer"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --version VERSION     Install specific version (default: latest)"
            echo "  --install-dir DIR     Installation directory (default: /usr/local/bin)"
            echo "  --help               Show this help message"
            echo ""
            echo "Environment Variables:"
            echo "  INSTALL_DIR          Alternative to --install-dir flag"
            echo "  VERSION              Alternative to --version flag"
            echo ""
            echo "Examples:"
            echo "  # Install latest version"
            echo "  curl -sSL https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/main/install.sh | bash"
            echo ""
            echo "  # Install specific version"
            echo "  curl -sSL https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/main/install.sh | bash -s -- --version v0.1.0"
            echo ""
            echo "  # Install to custom directory"
            echo "  curl -sSL https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/main/install.sh | bash -s -- --install-dir \$HOME/.local/bin"
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Run main installation
main