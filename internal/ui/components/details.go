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
	installedIcon := "✗"
	if pkg.LocallyInstalled {
		installedStatus = "[green]Installed[-]"
		installedIcon = "✓"

		if pkg.Outdated {
			installedStatus = "[orange]Update available[-]"
			installedIcon = "⟳"
		}
	}

	// Type tag with escaped brackets
	typeTag := tview.Escape("[F]") // Formula
	typeLabel := "Formula"
	if pkg.Type == models.PackageTypeCask {
		typeTag = tview.Escape("[C]") // Cask
		typeLabel = "Cask"
	}

	// Basic information with status
	basicInfo := fmt.Sprintf(
		"[yellow::b]%s %s[-]\n\n"+
			"[blue]• Type:[-] %s %s\n"+
			"[blue]• Name:[-] %s\n"+
			"[blue]• Display Name:[-] %s\n"+
			"[blue]• Version:[-] %s\n"+
			"[blue]• Status:[-] %s\n\n"+
			"[yellow::b]Description[-]\n%s\n\n"+
			"[blue]• Homepage:[-] %s",
		pkg.Name, installedIcon,
		typeTag, typeLabel,
		pkg.Name,
		pkg.DisplayName,
		pkg.Version,
		installedStatus,
		pkg.Description,
		pkg.Homepage,
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
	if !pkg.LocallyInstalled {
		return "[yellow::b]Installation[-]\nNot installed"
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
			"[yellow::b]Installation Details[-]\n"+
				"[blue]• Path:[-] %s\n"+
				"[blue]• Installed on request:[-] %s\n"+
				"[blue]• Installed as dependency:[-] %s\n"+
				"[blue]• Installed version:[-] %s",
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
			"[yellow::b]Installation Details[-]\n"+
				"[blue]• Type:[-] macOS Application\n"+
				"[blue]• Installed version:[-] %s",
			installedVersion,
		)
	}

	return "[yellow::b]Installation[-]\nInstalled"
}

func (d *Details) getDependenciesInfo(info *models.Formula) string {
	title := "[yellow::b]Dependencies[-]\n"

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
	title := "[yellow::b]Analytics[-]\n"

	p := message.NewPrinter(language.English)

	title += fmt.Sprintf("[blue]• 90d Global Rank:[-] %s\n", p.Sprintf("%d", pkg.Analytics90dRank))
	title += fmt.Sprintf("[blue]• 90d   Downloads:[-] %s\n", p.Sprintf("%d", pkg.Analytics90dDownloads))

	return title
}

func (d *Details) View() *tview.TextView {
	return d.view
}

func (d *Details) Clear() {
	d.view.Clear()
}
