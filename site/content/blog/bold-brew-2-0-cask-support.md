---
title: "Bold Brew 2.0: Complete Homebrew Management with Cask Support"
date: "2025-10-13"
description: "Bold Brew 2.0 brings major new features including full Cask support, Leaves filter, XDG compliance, and enhanced security. Manage both CLI tools and GUI applications seamlessly."
keywords: "Bold Brew 2.0, Homebrew casks, TUI package manager, leaves filter, XDG compliance, Homebrew GUI, terminal UI, package management, macOS apps, Linux apps"
---

# Bold Brew 2.0: Complete Homebrew Management with Cask Support

We're thrilled to announce **Bold Brew 2.0**, the biggest update since launch! This release transforms Bold Brew from a formula-only manager into a **complete Homebrew management solution** that handles both command-line tools and GUI applications.

## 🎉 What's New

### Full Homebrew Casks Support

The most requested feature is finally here! Bold Brew now provides **complete support for Homebrew Casks**, allowing you to manage GUI applications and binaries directly from the same intuitive interface you love.

**What are Casks?** Homebrew Casks extend Homebrew's package management to include macOS and Linux GUI applications. Instead of just installing command-line tools like `git` or `node`, you can now manage apps like Google Chrome, Visual Studio Code, Docker Desktop, and thousands more.

#### Visual Type Indicators

Never wonder what type of package you're looking at! Bold Brew 2.0 introduces clear visual indicators:
- `[F]` - Formula (command-line tool)
- `[C]` - Cask (GUI application or binary)

These tags appear throughout the interface, making it instantly clear whether you're working with a CLI tool or a desktop application.

### Smart Leaves Filter

Tired of seeing all those dependency packages cluttering your installed list? The new **Leaves Filter** (press `L`) shows only packages you explicitly installed, hiding all the dependencies that came along for the ride.

**Perfect for:**
- Cleaning up your system by identifying what you actually need
- Creating reproducible development environments
- Understanding your actual package footprint
- Selective updates of only your core tools

### XDG Base Directory Compliance

Bold Brew 2.0 now follows the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html), providing a cleaner, more standards-compliant cache management:

- **Linux**: `~/.cache/bbrew` or `$XDG_CACHE_HOME/bbrew`
- **macOS**: `~/Library/Caches/bbrew` (native macOS location!)
- **Windows** (WSL2): Windows Known Folders support

No more random dotfiles in your home directory—everything is where it should be.

## 🔧 Technical Improvements

### Go 1.25 and Modern Tooling

- **Updated to Go 1.25** for better performance and latest language features
- **Migrated to Podman** and OCI-compliant Containerfile for better security
- **Enhanced Makefile** with 15+ new targets including `make test`, `make security`, and `make install`
- **Improved build system** with local and containerized build options

### Enhanced Security

Security is a priority. Bold Brew 2.0 includes:
- **govulncheck** - Automated Go vulnerability scanning
- **gosec** - Static security analysis
- **GitHub Security integration** - SARIF reports uploaded to Security tab
- **Fixed memory aliasing issues** - Cleaner, safer code
- **Better permission handling** - Secure cache directory permissions (0750)

### Better User Experience

- **Enhanced keyboard shortcuts** - More intuitive navigation and filtering
- **Improved error messages** - Better debugging and user feedback
- **Analytics integration** - See popular packages based on 90-day download stats
- **Real-time feedback** - Live updates during package operations
- **Fixed rendering issues** - Proper display of all UI elements

## 🐛 Bug Fixes

- Fixed cask analytics endpoint (now correctly fetches download statistics)
- Corrected installed casks detection (properly identifies locally installed casks)
- Fixed tview special character rendering for type tags
- Improved directory permission handling for cache
- Enhanced error handling throughout the application

## 🚀 Getting Started

### For Existing Users

Update to the latest version:

```bash
brew update
brew upgrade bbrew
```

### For New Users

Install Bold Brew via Homebrew:

```bash
brew install Valkyrie00/homebrew-bbrew/bbrew
```

Or download from the [releases page](https://github.com/Valkyrie00/bold-brew/releases).

## 📖 Using the New Features

### Managing Casks

1. **Filter Casks Only**: Press `C` to show only Cask packages
2. **Search Casks**: Type `/` and search for your favorite GUI app (e.g., "chrome", "vscode", "docker")
3. **Install a Cask**: Select it and press `I`
4. **Update Casks**: Press `U` on any outdated Cask, or `Ctrl+U` to update all

### Using the Leaves Filter

1. Press `L` to activate the Leaves filter
2. Browse only the packages you explicitly installed
3. Identify packages you no longer need
4. Press `R` to remove unwanted packages

### Keyboard Shortcuts Reference

#### Filters
- `F` - Filter installed packages
- `O` - Filter outdated packages
- `L` - Filter leaves (explicitly installed, no dependencies)
- `C` - Filter casks only

#### Package Operations
- `I` - Install selected package
- `U` - Update selected package
- `R` - Remove selected package
- `Ctrl+U` - Update all outdated packages

## 🌐 Cross-Platform Support

Bold Brew 2.0 provides excellent support across platforms:

| Platform | Support | Notes |
|----------|---------|-------|
| 🍎 **macOS** | ✅ Full | Native Homebrew support with macOS-specific cache location |
| 🐧 **Linux** | ✅ Full | Linuxbrew/Homebrew support with XDG compliance |
| 🪟 **Windows** | ⚠️ Partial | Via WSL2 with Homebrew |

## 🎯 What's Next

We're not stopping here! Future plans include:
- **Tap management** - Add and manage custom Homebrew taps
- **Formulae pinning** - Pin specific package versions
- **Backup/restore** - Export and import your package lists
- **Themes** - Customizable color schemes
- **Plugin system** - Extend Bold Brew with custom functionality

## 🙏 Acknowledgments

A huge thank you to:
- The Homebrew team for the excellent package management system
- Project Bluefin for adopting Bold Brew as their official Homebrew TUI
- All contributors who submitted issues, PRs, and feature requests
- The community for their continued support and feedback

## 📣 Spread the Word

If you love Bold Brew 2.0:
- ⭐ [Star the project on GitHub](https://github.com/Valkyrie00/bold-brew)
- 🐦 Share on social media
- 📝 Write about it on your blog
- 💬 Tell your developer friends

## 🔗 Resources

- [GitHub Repository](https://github.com/Valkyrie00/bold-brew)
- [Documentation](https://bold-brew.com)
- [Release Notes](https://github.com/Valkyrie00/bold-brew/releases)
- [Report Issues](https://github.com/Valkyrie00/bold-brew/issues)

---

**Happy brewing! 🍺**

*The Bold Brew Team*

