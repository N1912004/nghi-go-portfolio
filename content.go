package main

import (
	"embed"
	"encoding/json"
	"fmt"
)

// All public content and browser assets ship with the binary. No runtime CDN.
//
//go:embed web/* web/assets/* data/profile.json
var assets embed.FS

type Project struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Category   string   `json:"category"`
	Period     string   `json:"period"`
	Summary    string   `json:"summary"`
	Stack      []string `json:"stack"`
	Highlights []string `json:"highlights"`
	SourceURL  string   `json:"source_url,omitempty"`
}
type Experience struct {
	Company string `json:"company"`
	Role    string `json:"role"`
	Period  string `json:"period"`
	Summary string `json:"summary"`
}
type Training struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Date     string `json:"date"`
}
type Profile struct {
	Name            string       `json:"name"`
	Role            string       `json:"role"`
	Location        string       `json:"location"`
	Email           string       `json:"email"`
	Github          string       `json:"github"`
	Linkedin        string       `json:"linkedin"`
	Intro           string       `json:"intro"`
	About           string       `json:"about"`
	ProductionStack []string     `json:"production_stack"`
	GoStack         []string     `json:"go_stack"`
	Experience      []Experience `json:"experience"`
	Projects        []Project    `json:"projects"`
	Training        []Training   `json:"training"`
	Education       string       `json:"education"`
}

func loadProfile() (Profile, error) {
	var p Profile
	b, err := assets.ReadFile("data/profile.json")
	if err != nil {
		return p, err
	}
	if err = json.Unmarshal(b, &p); err != nil {
		return p, fmt.Errorf("parse profile: %w", err)
	}
	if p.Name == "" || p.Email == "" {
		return p, fmt.Errorf("profile requires name and email")
	}
	return p, nil
}
