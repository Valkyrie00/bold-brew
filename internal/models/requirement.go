package models

type Requirement struct {
	Name    string   `json:"name"`
	Version string   `json:"version"`
	Contexts []string `json:"contexts"`
}
