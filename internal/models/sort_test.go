package models

import "testing"

func TestSortMode_String(t *testing.T) {
	tests := []struct {
		mode SortMode
		want string
	}{
		{SortNone, "None"},
		{SortByDownloads, "Downloads"},
		{SortByName, "Name"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("SortMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestSortMode_Next(t *testing.T) {
	tests := []struct {
		mode SortMode
		want SortMode
	}{
		{SortNone, SortByDownloads},
		{SortByDownloads, SortByName},
		{SortByName, SortNone},
	}

	for _, tt := range tests {
		if got := tt.mode.Next(); got != tt.want {
			t.Errorf("SortMode(%d).Next() = %d, want %d", tt.mode, got, tt.want)
		}
	}
}
