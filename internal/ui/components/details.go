package components

import (
	"bbrew/internal/models"
	"bbrew/internal/ui/theme"
	"fmt"
	"strings"

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
	details.view.SetTitle("Details")
	details.view.SetTitleColor(theme.TitleColor)
	details.view.SetTitleAlign(tview.AlignLeft)
	details.view.SetBorder(true)
	details.view.SetBorderPadding(0, 0, 1, 1)
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

	// Type tag with escaped brackets
	typeTag := tview.Escape("[F]") // Formula
	typeLabel := "Formula"
	if pkg.Type == models.PackageTypeCask {
		typeTag = tview.Escape("[C]") // Cask
		typeLabel = "Cask"
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
			"[blue]• Status:[-] %s\n"+
			"[blue]• Homepage:[-] %s\n\n"+
			"[yellow::b]Description[-]\n%s\n%s",
		pkg.Name, separator,
		typeTag, typeLabel,
		pkg.Name,
		pkg.DisplayName,
		pkg.Version,
		installedStatus,
		pkg.Homepage,
		separator,
		pkg.Description,
	)

	// Installation details
	installDetails := d.getPackageInstallationDetails(pkg)

	// Dependencies (only for formulae)
	dependenciesInfo := ""
	if pkg.Type == models.PackageTypeFormula && pkg.Formula != nil {
		dependenciesInfo = d.getDependenciesInfo(pkg.Formula)
	}

	analyticsInfo := d.getAnalyticsInfo(pkg)

	parts := []string{basicInfo, installDetails}
	if dependenciesInfo != "" {
		parts = append(parts, dependenciesInfo)
	}
	parts = append(parts, analyticsInfo)

	d.view.SetText(strings.Join(parts, "\n\n"))
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
