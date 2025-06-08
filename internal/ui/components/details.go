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

func (d *Details) SetContent(info *models.Formula) {
	if info == nil {
		d.view.SetText("")
		return
	}

	// Installation status with colors
	installedStatus := "[red]Not installed[-]"
	installedIcon := "✗"
	if len(info.Installed) > 0 {
		installedStatus = "[green]Installed[-]"
		installedIcon = "✓"

		if info.Outdated {
			installedStatus = "[orange]Update available[-]"
			installedIcon = "⟳"
		}
	}

	// Basic information with icons
	basicInfo := fmt.Sprintf(
		"[yellow::b]%s %s[-]\n\n"+
			"[blue]• Name:[-] %s\n"+
			"[blue]• Version:[-] %s\n"+
			"[blue]• Status:[-] %s %s\n"+
			"[blue]• Tap:[-] %s\n"+
			"[blue]• License:[-] %s\n\n"+
			"[yellow::b]Description[-]\n%s\n\n"+
			"[blue]• Homepage:[-] %s",
		info.Name, installedIcon,
		info.FullName,
		info.Versions.Stable,
		installedStatus, d.getPackageVersionInfo(info),
		info.Tap,
		info.License,
		info.Description,
		info.Homepage,
	)

	// Installation details
	installDetails := d.getPackageInstallationDetails(info)

	// Dependencies with improved formatting
	dependenciesInfo := d.getDependenciesInfo(info)

	analyticsInfo := d.getAnalyticsInfo(info)

	d.view.SetText(strings.Join([]string{basicInfo, installDetails, dependenciesInfo, analyticsInfo}, "\n\n"))
}

func (d *Details) getPackageVersionInfo(info *models.Formula) string {
	if len(info.Installed) == 0 {
		return ""
	}

	installedVersion := info.Installed[0].Version
	stableVersion := info.Versions.Stable

	// Revision version
	if strings.HasPrefix(installedVersion, stableVersion+"_") {
		return fmt.Sprintf("([green]%s[-])", installedVersion)
	} else if installedVersion == stableVersion {
		return fmt.Sprintf("([green]%s[-])", installedVersion)
	} else if installedVersion < stableVersion || info.Outdated {
		return fmt.Sprintf("([orange]%s[-] → [green]%s[-])",
			installedVersion, stableVersion)
	}

	// Other cases
	return fmt.Sprintf("([green]%s[-])", installedVersion)
}

func (d *Details) getPackageInstallationDetails(info *models.Formula) string {
	if len(info.Installed) == 0 {
		return "[yellow::b]Installation[-]\nNot installed"
	}

	packagePrefix := info.LocalPath

	installedOnRequest := "No"
	if info.Installed[0].InstalledOnRequest {
		installedOnRequest = "Yes"
	}

	installedAsDependency := "No"
	if info.Installed[0].InstalledAsDependency {
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
		info.Installed[0].Version,
	)
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

func (d *Details) getAnalyticsInfo(info *models.Formula) string {
	title := "[yellow::b]Analytics[-]\n"

	p := message.NewPrinter(language.English)

	title += fmt.Sprintf("[blue]• 90d Global Rank:[-] %s\n", p.Sprintf("%d", info.Analytics90dRank))
	title += fmt.Sprintf("[blue]• 90d   Downloads:[-] %s\n", p.Sprintf("%d", info.Analytics90dDownloads))

	return title
}

func (d *Details) View() *tview.TextView {
	return d.view
}

func (d *Details) Clear() {
	d.view.Clear()
}
