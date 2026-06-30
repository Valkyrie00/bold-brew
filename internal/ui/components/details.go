package components

import (
	"bbrew/internal/models"
	"bbrew/internal/ui/theme"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Details struct {
	view  *tview.TextView
	theme *theme.Theme
}

func NewDetails(theme *theme.Theme) *Details {
	details := &Details{
		view:  tview.NewTextView(),
		theme: theme,
	}

	details.view.SetDynamicColors(true)
	details.view.SetTextAlign(tview.AlignLeft)
	details.view.SetBorder(false)
	details.view.SetBorderPadding(0, 0, 3, 1)
	details.view.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		borderStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
		titleStyle := tcell.StyleDefault.Foreground(theme.TitleColor)
		for row := y; row < y+height; row++ {
			screen.SetContent(x, row, tview.Borders.Vertical, nil, borderStyle)
		}
		title := "Details"
		for i, ch := range title {
			screen.SetContent(x+2+i, y, ch, nil, titleStyle)
		}
		return x + 3, y + 2, width - 3, height - 2
	})
	return details
}

func (d *Details) SetContent(pkg *models.Package) {
	if pkg == nil {
		d.view.SetText("")
		return
	}

	// Installation status with colors
	installedStatus := "[red]Not installed[-]"
	if pkg.LocallyInstalled {
		installedStatus = "[green]Installed[-]"
		if pkg.Outdated {
			installedStatus = "[orange]Update available[-]"
		}
	}

	// Health warning inline (shown next to status)
	healthInline := ""
	if pkg.Disabled {
		healthInline = " [red::b]⚠ DISABLED[-]"
	} else if pkg.Deprecated {
		healthInline = " [yellow::b]⚠ DEPRECATED[-]"
	}

	// Type tag with escaped brackets
	var typeTag, typeLabel string
	switch pkg.Type {
	case models.PackageTypeCask:
		typeTag = tview.Escape("[C]")
		typeLabel = "Cask"
	case models.PackageTypeFlatpak:
		typeTag = tview.Escape("[P]")
		typeLabel = "Flatpak"
	default:
		typeTag = tview.Escape("[F]")
		typeLabel = "Formula"
	}

	// Section separator
	separator := "[dim]────────────────────────[-]"

	// Basic information with status
	basicInfo := fmt.Sprintf(
		"[yellow::b]%s[-]\n%s\n"+
			"[blue]• Type:[-] %s %s\n"+
			"[blue]• Name:[-] %s\n"+
			"[blue]• Display Name:[-] %s\n"+
			"[blue]• Version:[-] %s\n"+
			"[blue]• Status:[-] %s%s\n"+
			"[blue]• Homepage:[-] %s\n\n"+
			"[yellow::b]Description[-]\n%s\n%s",
		pkg.Name, separator,
		typeTag, typeLabel,
		pkg.Name,
		pkg.DisplayName,
		pkg.Version,
		installedStatus, healthInline,
		pkg.Homepage,
		separator,
		pkg.Description,
	)

	// Health section (only for deprecated/disabled)
	healthSection := d.getHealthInfo(pkg)

	// Installation details
	installDetails := d.getPackageInstallationDetails(pkg)

	// Dependencies (only for formulae)
	dependenciesInfo := ""
	if pkg.Type == models.PackageTypeFormula && pkg.Formula != nil {
		dependenciesInfo = d.getDependenciesInfo(pkg.Formula)
	}

	analyticsInfo := d.getAnalyticsInfo(pkg)

	parts := []string{basicInfo}
	if healthSection != "" {
		parts = append(parts, healthSection)
	}
	parts = append(parts, installDetails)
	if dependenciesInfo != "" {
		parts = append(parts, dependenciesInfo)
	}
	parts = append(parts, analyticsInfo)

	d.view.SetText(strings.Join(parts, "\n\n"))
}

func (d *Details) getHealthInfo(pkg *models.Package) string {
	if !pkg.Deprecated && !pkg.Disabled {
		return ""
	}

	separator := "[dim]────────────────────────[-]"

	var title, reason, date, replacement string

	if pkg.Disabled {
		title = "[red::b]⚠ Package Disabled[-]"
		if pkg.Formula != nil {
			reason = interfaceToString(pkg.Formula.DisableReason)
			date = interfaceToString(pkg.Formula.DisableDate)
			replacement = interfaceToString(pkg.Formula.DisableReplacement)
		} else if pkg.Cask != nil {
			reason = interfaceToString(pkg.Cask.DisableReason)
			date = interfaceToString(pkg.Cask.DisableDate)
		}
	} else {
		title = "[yellow::b]⚠ Package Deprecated[-]"
		if pkg.Formula != nil {
			reason = interfaceToString(pkg.Formula.DeprecationReason)
			date = interfaceToString(pkg.Formula.DeprecationDate)
			replacement = interfaceToString(pkg.Formula.DeprecationReplacement)
		} else if pkg.Cask != nil {
			reason = interfaceToString(pkg.Cask.DeprecationReason)
			date = interfaceToString(pkg.Cask.DeprecationDate)
		}
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n%s\n", title, separator)

	if reason != "" {
		fmt.Fprintf(&sb, "[blue]• Reason:[-] %s\n", reason)
	}
	if date != "" {
		fmt.Fprintf(&sb, "[blue]• Since:[-] %s\n", date)
	}
	if replacement != "" {
		fmt.Fprintf(&sb, "[blue]• Replacement:[-] [green]%s[-]\n", replacement)
	}

	if pkg.Disabled {
		sb.WriteString("\n[dim]This package can no longer be installed.[-]")
	} else {
		sb.WriteString("\n[dim]Consider migrating to the replacement before this package is removed.[-]")
	}

	return sb.String()
}

func interfaceToString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func (d *Details) getPackageInstallationDetails(pkg *models.Package) string {
	separator := "[dim]────────────────────────[-]"

	if !pkg.LocallyInstalled {
		return fmt.Sprintf("[yellow::b]Installation[-]\n%s\nNot installed", separator)
	}

	// For formulae, show detailed installation info
	if pkg.Type == models.PackageTypeFormula && pkg.Formula != nil && len(pkg.Formula.Installed) > 0 {
		packagePrefix := pkg.Formula.LocalPath

		installedOnRequest := "No"
		if pkg.Formula.Installed[0].InstalledOnRequest {
			installedOnRequest = "Yes"
		}

		installedAsDependency := "No"
		if pkg.Formula.Installed[0].InstalledAsDependency {
			installedAsDependency = "Yes"
		}

		return fmt.Sprintf(
			"[yellow::b]Installation Details[-]\n%s\n"+
				"[blue]• Path:[-] %s\n"+
				"[blue]• Installed on request:[-] %s\n"+
				"[blue]• Installed as dependency:[-] %s\n"+
				"[blue]• Installed version:[-] %s",
			separator,
			packagePrefix,
			installedOnRequest,
			installedAsDependency,
			pkg.Formula.Installed[0].Version,
		)
	}

	// For casks, show simpler installation info
	if pkg.Type == models.PackageTypeCask && pkg.Cask != nil {
		installedVersion := "Unknown"
		if pkg.Cask.Installed != nil {
			installedVersion = *pkg.Cask.Installed
		}

		return fmt.Sprintf(
			"[yellow::b]Installation Details[-]\n%s\n"+
				"[blue]• Type:[-] Desktop Application\n"+
				"[blue]• Installed version:[-] %s",
			separator,
			installedVersion,
		)
	}

	return fmt.Sprintf("[yellow::b]Installation[-]\n%s\nInstalled", separator)
}

func (d *Details) getDependenciesInfo(info *models.Formula) string {
	separator := "[dim]────────────────────────[-]"
	title := fmt.Sprintf("[yellow::b]Dependencies[-]\n%s\n", separator)

	if len(info.Dependencies) == 0 {
		return title + "No dependencies"
	}

	// Format dependencies in multiple columns or with separators
	deps := ""
	for i, dep := range info.Dependencies {
		deps += dep
		if i < len(info.Dependencies)-1 {
			if (i+1)%3 == 0 {
				deps += "\n"
			} else {
				deps += ", "
			}
		}
	}

	return title + deps
}

func (d *Details) getAnalyticsInfo(pkg *models.Package) string {
	separator := "[dim]────────────────────────[-]"
	p := message.NewPrinter(language.English)

	return fmt.Sprintf(
		"[yellow::b]Analytics[-]\n%s\n"+
			"[blue]• 90d Global Rank:[-] %s\n"+
			"[blue]• 90d Downloads:[-] %s",
		separator,
		p.Sprintf("%d", pkg.Analytics90dRank),
		p.Sprintf("%d", pkg.Analytics90dDownloads),
	)
}

func (d *Details) View() *tview.TextView {
	return d.view
}

func (d *Details) Clear() {
	d.view.Clear()
}
