package models

// SortMode represents how the package list is sorted.
type SortMode int

const (
	SortByDownloads SortMode = iota
	SortByName
	SortByInstalled
)

func (s SortMode) String() string {
	switch s {
	case SortByName:
		return "Name"
	case SortByInstalled:
		return "Installed"
	default:
		return "Downloads"
	}
}

// Next cycles to the next sort mode.
func (s SortMode) Next() SortMode {
	return (s + 1) % 3
}
