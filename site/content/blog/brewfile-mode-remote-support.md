---
title: "Brewfile Mode: Curated Package Collections & Remote URL Support"
date: "2025-12-29"
description: "Bold Brew now supports Brewfile mode for curated package collections, plus the ability to load Brewfiles directly from HTTPS URLs. Perfect for team configurations and themed installers."
keywords: "Brewfile mode, remote Brewfile, HTTPS URL, curated packages, team configuration, Homebrew TUI, package collections, IDE installer, dev tools, Bold Brew"
---

# Brewfile Mode: Curated Package Collections & Remote URL Support

We're excited to announce a major new capability in Bold Brew: **Brewfile Mode** with full support for **remote Brewfiles via HTTPS URLs**. This feature transforms how teams and individuals manage their Homebrew package collections.

## 🎯 What is Brewfile Mode?

Brewfile Mode allows you to launch Bold Brew with a curated list of packages instead of the full Homebrew catalog. This is perfect for:

- **Themed installers** (IDE tools, AI/ML packages, DevOps tools)
- **Team onboarding** (share your team's essential tools)
- **Project-specific setups** (install only what a project needs)
- **Personal collections** (your favorite tools in one place)

### How It Works

```bash
# Local Brewfile
bbrew -f ~/Brewfile

# Remote Brewfile (NEW!)
bbrew -f https://raw.githubusercontent.com/your-org/configs/main/dev-tools.Brewfile
```

When you launch in Brewfile mode, Bold Brew shows **only** the packages defined in that Brewfile. You get the full Bold Brew experience—search, filters, install/remove—but focused on your curated selection.

## 🌐 Remote Brewfiles via HTTPS

The latest update adds support for loading Brewfiles directly from URLs. This opens up exciting possibilities:

### Share Team Configurations

```bash
# Everyone on the team uses the same dev tools
bbrew -f https://github.com/acme-corp/dotfiles/raw/main/Brewfile
```

### Create Themed Installers

```bash
# AI/ML development environment
bbrew -f https://example.com/brewfiles/ai-ml-toolkit.Brewfile

# Kubernetes tools collection
bbrew -f https://example.com/brewfiles/k8s-essentials.Brewfile

# Frontend development stack
bbrew -f https://example.com/brewfiles/frontend-dev.Brewfile
```

### Quick Setup for New Machines

Share a single URL with colleagues or include it in your README:

```markdown
## Development Setup

Install our recommended tools:
\`\`\`bash
brew install Valkyrie00/homebrew-bbrew/bbrew
bbrew -f https://our-company.com/dev-setup.Brewfile
\`\`\`
```

## 📦 Third-Party Tap Support

Brewfile Mode also includes full support for **third-party taps**. Your Brewfile can include packages from any Homebrew tap:

```ruby
# Brewfile example
tap "homebrew/cask-fonts"
tap "ublue-os/homebrew-tap"

# Core tools
brew "git"
brew "gh"
brew "jq"

# Fonts
cask "font-fira-code"
cask "font-jetbrains-mono"

# From third-party tap
cask "some-custom-package"
```

Bold Brew automatically:
1. **Installs missing taps** at startup
2. **Caches tap package info** for faster subsequent loads
3. **Shows real-time progress** during tap installation

## 🔒 Security First

Remote Brewfiles are loaded securely:

- **HTTPS only** — HTTP URLs are rejected for security
- **Temporary storage** — Downloaded files are automatically cleaned up
- **No persistence** — Remote content isn't cached between sessions

## 🎨 Use Cases

### 1. IDE Chooser for Teams

Create a Brewfile with all supported IDEs and let developers pick:

```ruby
# ide-chooser.Brewfile
cask "visual-studio-code"
cask "sublime-text"
cask "jetbrains-toolbox"
cask "zed"
cask "cursor"
```

```bash
bbrew -f https://team.example.com/ide-chooser.Brewfile
```

### 2. Project Onboarding

Include a Brewfile in your project repo:

```ruby
# project/.Brewfile
brew "node"
brew "pnpm"
brew "docker"
cask "docker"
```

New contributors run one command to get all dependencies.

### 3. Personal Dotfiles

Keep your Brewfile in your dotfiles repo and access it from anywhere:

```bash
bbrew -f https://github.com/username/dotfiles/raw/main/Brewfile
```

## 📋 Example Brewfiles

Bold Brew includes example Brewfiles in the `examples/` directory:

- **`all.brewfile`** — Comprehensive package collection
- **`ide.Brewfile`** — Popular IDEs and editors

Check them out for inspiration!

## 🚀 Getting Started

1. **Update Bold Brew** to the latest version:
   ```bash
   brew upgrade bbrew
   ```

2. **Try a local Brewfile**:
   ```bash
   bbrew -f ~/path/to/Brewfile
   ```

3. **Try a remote Brewfile**:
   ```bash
   bbrew -f https://raw.githubusercontent.com/Valkyrie00/bold-brew/main/examples/all.brewfile
   ```

## 🎹 Brewfile Mode Shortcuts

When running in Brewfile mode, you get additional keyboard shortcuts:

| Key | Action |
|-----|--------|
| `Ctrl+A` | Install all packages from Brewfile |
| `Ctrl+R` | Remove all packages from Brewfile |

All standard shortcuts (search, filters, individual install/remove) work exactly as expected.

## 💡 Tips & Best Practices

1. **Version your Brewfiles** — Keep them in git for change tracking
2. **Use comments** — Document why each package is included
3. **Organize by category** — Group related packages together
4. **Test before sharing** — Verify packages exist and install correctly

## 🔗 Resources

- [Homebrew Brewfile documentation](https://github.com/Homebrew/homebrew-bundle)
- [Bold Brew on GitHub](https://github.com/Valkyrie00/bold-brew)
- [Example Brewfiles](https://github.com/Valkyrie00/bold-brew/tree/main/examples)

---

**Happy brewing with curated collections! 🍺**

*The Bold Brew Team*

