---
title: "Bold Brew v2.3: Homebrew 6, Vulnerability Scanning, and Multi-Source Support"
date: "2026-07-02"
description: "Bold Brew v2.3 brings full Homebrew 6.0 compatibility, vulnerability scanning with brew vulns, Mac App Store and Flatpak support, sort modes, Brewfile export, and a formulae-only filter."
keywords: "Bold Brew 2.3, Homebrew 6, vulnerability scanning, brew vulns, Mac App Store, Flatpak, Brewfile export, formulae filter, TUI, package manager"
tags:
  - Homebrew
  - Security
  - TUI
  - macOS
  - Linux
---

# Bold Brew v2.3: Homebrew 6, Vulnerability Scanning, and Multi-Source Support

We're excited to announce **Bold Brew v2.3** — the biggest feature release since 2.0. This update brings full Homebrew 6.0 compatibility, integrated vulnerability scanning, multi-source package support, and quality-of-life improvements that make managing your packages faster and more secure.

## Homebrew 6.0 Compatibility

Homebrew 6.0 introduced significant changes to how taps, installations, and the API work. Bold Brew v2.3 is fully compatible:

- **Tap trust** — Bold Brew respects the new trust model and prompts when needed
- **Ask mode** — interactive confirmation flows work seamlessly within the TUI
- **JSON v2 API** — faster metadata fetching with the updated API format

If you're already on Homebrew 6, Bold Brew will work out of the box with no configuration changes.

## Vulnerability Scanning

Security matters. Bold Brew now integrates with `brew vulns` to detect known CVEs in your installed packages:

- Press **`v`** anywhere to trigger a vulnerability scan
- Results are displayed inline with severity indicators
- If `brew vulns` isn't installed, Bold Brew offers to install it automatically

This makes it trivial to audit your development environment for known security issues without leaving the TUI.

## Mac App Store Support

Bold Brew now understands `mas` entries in Brewfiles. This means you can manage your Mac App Store applications alongside Homebrew packages in a unified workflow:

```
mas "Xcode", id: 497799835
mas "1Password", id: 1333542190
```

Load a Brewfile with mas entries using `bbrew -f ~/Brewfile` and see everything in one place.

## Flatpak Support (Linux)

On Linux, Bold Brew now supports `flatpak` entries in Brewfiles:

```
flatpak "org.gimp.GIMP"
flatpak "org.mozilla.firefox"
```

This is especially useful on Project Bluefin and Aurora where both Homebrew and Flatpak are part of the default setup.

## Formulae-Only Filter

A new filter mode lets you show only formulae (excluding casks):

- Press **`F`** to toggle the formulae-only filter
- Combines with search for focused formula discovery

This brings the total number of filter modes to six: all, installed, outdated, leaves, casks, and formulae.

## Sort Modes

You can now sort packages by popularity or name:

- Press **`s`** to cycle through sort modes
- Sort by **downloads** (most popular first) or **name** (alphabetical)

Sorting by downloads is great for discovering widely-used packages you might not know about.

## Brewfile Export

Export your current selection to a Brewfile:

- Press **`e`** to export
- Saves to `~/Brewfile` in the standard format
- Respects current filters — export only what you're viewing

This makes it easy to snapshot your setup for sharing or backup.

## CI Pipeline

Bold Brew now has a full CI pipeline with automated tests and coverage reporting. This ensures reliability across releases and makes contributing easier.

## Upgrade

Update Bold Brew to v2.3 with:

```
brew upgrade bbrew
```

Or install fresh:

```
brew install bbrew
```

Check out the [full changelog](/changelog/) for the complete list of changes.

---

*Bold Brew is the official Homebrew TUI for Project Bluefin, Aurora, and Universal Blue. Used by tens of thousands of developers on macOS and Linux.*
