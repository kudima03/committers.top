package merge

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"most-active-github-users-counter/github"
)

// Slug is the location name under which the merged rating is published.
const Slug = "global"

// MaxAmount is the largest rating size which can be computed exactly: every
// regional list contains only the top 256 users of its region.
const MaxAmount = 256

// LocationUser mirrors a user entry written by output.YamlOutput.
type LocationUser struct {
	Rank          int    `yaml:"rank"`
	Name          string `yaml:"name"`
	Login         string `yaml:"login"`
	AvatarURL     string `yaml:"avatarUrl"`
	Contributions int    `yaml:"contributions"`
	Company       string `yaml:"company"`
	Organizations string `yaml:"organizations"`
}

// Location is the subset of a regional data file needed for merging.
type Location struct {
	Slug      string         `yaml:"-"`
	Title     string         `yaml:"title"`
	Generated time.Time      `yaml:"generated"`
	Users     []LocationUser `yaml:"users"`
	Public    []LocationUser `yaml:"users_public_contributions"`
	Private   []LocationUser `yaml:"private_users"`
}

// User is a ranked user of the merged rating.
type User struct {
	github.User
	Contributions int
	Regions       []string
	generated     time.Time
}

type Result struct {
	Users           []User // public commits
	Public          []User // public contributions
	Private         []User // all contributions
	RegionCount     int
	ConsideredUsers int
	OldestData      time.Time
	NewestData      time.Time
}

// LoadDir parses all regional data files in dir, except for the given slugs.
func LoadDir(dir string, skip ...string) ([]Location, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return nil, err
	}

	locations := []Location{}
Files:
	for _, file := range files {
		slug := strings.TrimSuffix(filepath.Base(file), ".yml")
		for _, s := range skip {
			if slug == s {
				continue Files
			}
		}

		content, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		location := Location{}
		if err := yaml.Unmarshal(content, &location); err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		location.Slug = slug
		locations = append(locations, location)
	}

	if len(locations) == 0 {
		return nil, errors.New("no regional data found in " + dir)
	}
	return locations, nil
}

// Merge combines regional ratings into a single one, keeping at most amount users per list.
func Merge(locations []Location, amount int) Result {
	result := Result{
		Users:       mergeList(locations, func(l Location) []LocationUser { return l.Users }, amount),
		Public:      mergeList(locations, func(l Location) []LocationUser { return l.Public }, amount),
		Private:     mergeList(locations, func(l Location) []LocationUser { return l.Private }, amount),
		RegionCount: len(locations),
	}

	logins := map[string]bool{}
	for _, l := range locations {
		for _, list := range [][]LocationUser{l.Users, l.Public, l.Private} {
			for _, u := range list {
				logins[strings.ToLower(u.Login)] = true
			}
		}
		if l.Generated.IsZero() {
			continue
		}
		if result.OldestData.IsZero() || l.Generated.Before(result.OldestData) {
			result.OldestData = l.Generated
		}
		if l.Generated.After(result.NewestData) {
			result.NewestData = l.Generated
		}
	}
	result.ConsideredUsers = len(logins)

	return result
}

func mergeList(locations []Location, list func(Location) []LocationUser, amount int) []User {
	byLogin := map[string]*User{}
	for _, l := range locations {
		for _, u := range list(l) {
			key := strings.ToLower(u.Login)
			existing, found := byLogin[key]
			if !found {
				existing = &User{}
				byLogin[key] = existing
			}
			existing.Regions = append(existing.Regions, l.Slug)
			// the same user can be listed in several regions: trust the freshest data
			if !found || l.Generated.After(existing.generated) {
				existing.User = toGithubUser(u)
				existing.Contributions = u.Contributions
				existing.generated = l.Generated
			}
		}
	}

	users := make([]User, 0, len(byLogin))
	for _, u := range byLogin {
		sort.Strings(u.Regions)
		users = append(users, *u)
	}
	sort.Slice(users, func(i, j int) bool {
		if users[i].Contributions != users[j].Contributions {
			return users[i].Contributions > users[j].Contributions
		}
		return strings.ToLower(users[i].Login) < strings.ToLower(users[j].Login)
	})

	if len(users) > amount {
		users = users[:amount]
	}
	return users
}

func toGithubUser(u LocationUser) github.User {
	organizations := []string{}
	for _, org := range strings.Split(u.Organizations, ",") {
		if org != "" {
			organizations = append(organizations, org)
		}
	}
	return github.User{
		Login:         u.Login,
		AvatarURL:     u.AvatarURL,
		Name:          u.Name,
		Company:       u.Company,
		Organizations: organizations,
	}
}
