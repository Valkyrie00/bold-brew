#!/bin/bash
#
# Bold Brew Installer
# https://github.com/Valkyrie00/bold-brew
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Valkyrie00/bold-brew/main/install.sh | bash
#
# This script will:
#   1. Install Homebrew (if not already installed)
#   2. Install Bold Brew via Homebrew
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Logging functions
info() {
    echo -e "${BLUE}==>${NC} ${BOLD}$1${NC}"
}

success() {
    echo -e "${GREEN}==>${NC} ${BOLD}$1${NC}"
}

warn() {
    echo -e "${YELLOW}Warning:${NC} $1"
}

error() {
    echo -e "${RED}Error:${NC} $1" >&2
    exit 1
}

# Print banner
print_banner() {
    echo -e "${CYAN}"
    echo "  ____        _     _   ____                    "
    echo " | __ )  ___ | | __| | | __ ) _ __ _____      __"
    echo " |  _ \\ / _ \\| |/ _\` | |  _ \\| '__/ _ \\ \\ /\\ / /"
    echo " | |_) | (_) | | (_| | | |_) | | |  __/\\ V  V / "
    echo " |____/ \\___/|_|\\__,_| |____/|_|  \\___| \\_/\\_/  "
    echo -e "${NC}"
    echo -e "${BOLD}The Modern Homebrew TUI${NC}"
    echo ""
}

# Detect OS
detect_os() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"
    
    case "$OS" in
        Darwin)
            OS_TYPE="macos"
            if [ "$ARCH" = "arm64" ]; then
                BREW_PREFIX="/opt/homebrew"
            else
                BREW_PREFIX="/usr/local"
            fi
            ;;
        Linux)
            OS_TYPE="linux"
            BREW_PREFIX="/home/linuxbrew/.linuxbrew"
            ;;
        *)
            error "Unsupported operating system: $OS"
            ;;
    esac
    
    info "Detected: $OS ($ARCH)"
}

# Check if a command exists
command_exists() {
    command -v "$1" &> /dev/null
}

# Get Homebrew binary path
get_brew_path() {
    if [ -x "$BREW_PREFIX/bin/brew" ]; then
        echo "$BREW_PREFIX/bin/brew"
    elif command_exists brew; then
        command -v brew
    else
        echo ""
    fi
}

# Setup Homebrew environment
setup_brew_env() {
    local brew_bin="$1"
    if [ -n "$brew_bin" ] && [ -x "$brew_bin" ]; then
        eval "$("$brew_bin" shellenv)"
    fi
}

# Install Homebrew
install_homebrew() {
    info "Installing Homebrew..."
    echo ""
    
    # Use Homebrew's official installer
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    
    # Setup environment after installation
    local brew_bin
    brew_bin=$(get_brew_path)
    
    if [ -z "$brew_bin" ]; then
        error "Homebrew installation failed. Please install manually: https://brew.sh"
    fi
    
    setup_brew_env "$brew_bin"
    success "Homebrew installed successfully!"
}

# Install Bold Brew
install_boldbrew() {
    info "Installing Bold Brew..."
    
    brew install Valkyrie00/homebrew-bbrew/bbrew
    
    success "Bold Brew installed successfully!"
}

# Print post-install instructions
print_instructions() {
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║${NC}  ${BOLD}✅ Installation Complete!${NC}                                 ${GREEN}║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  Run ${CYAN}${BOLD}bbrew${NC} to start managing your Homebrew packages!"
    echo ""
    echo -e "  ${BOLD}Quick Start:${NC}"
    echo -e "    ${CYAN}bbrew${NC}                    # Browse all packages"
    echo -e "    ${CYAN}bbrew -f ~/Brewfile${NC}      # Use a Brewfile"
    echo -e "    ${CYAN}bbrew --help${NC}             # Show all options"
    echo ""
    echo -e "  ${BOLD}Documentation:${NC} ${BLUE}https://bold-brew.com${NC}"
    echo -e "  ${BOLD}GitHub:${NC}        ${BLUE}https://github.com/Valkyrie00/bold-brew${NC}"
    echo ""
    
    # Shell configuration reminder for Linux
    if [ "$OS_TYPE" = "linux" ]; then
        echo -e "  ${YELLOW}Note:${NC} You may need to restart your terminal or run:"
        echo -e "    ${CYAN}eval \"\$(${BREW_PREFIX}/bin/brew shellenv)\"${NC}"
        echo ""
    fi
}

# Main installation flow
main() {
    print_banner
    detect_os
    
    # Check for curl
    if ! command_exists curl; then
        error "curl is required but not installed. Please install curl first."
    fi
    
    # Check/Install Homebrew
    local brew_bin
    brew_bin=$(get_brew_path)
    
    if [ -n "$brew_bin" ]; then
        success "Homebrew is already installed"
        setup_brew_env "$brew_bin"
    else
        install_homebrew
    fi
    
    # Check if bbrew is already installed
    if command_exists bbrew; then
        warn "Bold Brew is already installed. Upgrading..."
        brew upgrade bbrew 2>/dev/null || brew install Valkyrie00/homebrew-bbrew/bbrew
    else
        install_boldbrew
    fi
    
    print_instructions
}

# Run main
main "$@"

