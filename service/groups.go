package service

import (
	"encoding/json"
	"fmt"
	"os"
)

// Group represents a mutual-block group: only the "Active" domain is allowed;
// all other members are always blocked.
type Group struct {
	// Name is the human-readable identifier for this group.
	Name string `json:"name"`

	// Active is the one domain in this group that is currently allowed ("" = all blocked).
	Active string `json:"active"`

	// Domains lists every domain belonging to this group.
	Domains []string `json:"domains"`
}

// groupsFile is the on-disk JSON envelope.
type groupsFile struct {
	Groups []Group `json:"groups"`
}

// ReadGroups reads groups.json and returns the list of groups.
// If the file does not exist, it returns an empty slice (not an error).
func ReadGroups(path string) ([]Group, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Group{}, nil
		}
		return nil, fmt.Errorf("could not read groups file: %w", err)
	}
	var gf groupsFile
	if err := json.Unmarshal(data, &gf); err != nil {
		return nil, fmt.Errorf("could not parse groups file: %w", err)
	}
	return gf.Groups, nil
}

// WriteGroups writes the groups slice to groups.json, creating the file if needed.
func WriteGroups(path string, groups []Group) error {
	gf := groupsFile{Groups: groups}
	data, err := json.MarshalIndent(gf, "", "  ")
	if err != nil {
		return fmt.Errorf("could not encode groups: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("could not write groups file: %w", err)
	}
	return nil
}

// AddGroup creates a new group with the given name and domains.
// Returns an error if:
//   - a group with that name already exists
//   - any domain is already a member of another group
//   - fewer than 2 domains are provided
func AddGroup(path string, name string, domains []string) error {
	if len(domains) < 2 {
		return fmt.Errorf("a group must have at least 2 domains")
	}

	groups, err := ReadGroups(path)
	if err != nil {
		return err
	}

	// Check for duplicate group name
	for _, g := range groups {
		if g.Name == name {
			return fmt.Errorf("group '%s' already exists", name)
		}
	}

	// Check that no domain belongs to another group already
	for _, domain := range domains {
		domain = normalizeDomain(domain)
		for _, g := range groups {
			for _, d := range g.Domains {
				if d == domain {
					return fmt.Errorf("domain '%s' already belongs to group '%s'", domain, g.Name)
				}
			}
		}
	}

	// Normalize all domains and deduplicate
	normalized := make([]string, 0, len(domains))
	seen := map[string]bool{}
	for _, d := range domains {
		d = normalizeDomain(d)
		if seen[d] {
			continue
		}
		seen[d] = true
		normalized = append(normalized, d)
	}

	groups = append(groups, Group{
		Name:    name,
		Active:  "",
		Domains: normalized,
	})
	return WriteGroups(path, groups)
}

// RemoveGroup deletes the named group. Returns an error if the group does not exist.
func RemoveGroup(path, name string) error {
	groups, err := ReadGroups(path)
	if err != nil {
		return err
	}

	found := false
	var updated []Group
	for _, g := range groups {
		if g.Name == name {
			found = true
			continue
		}
		updated = append(updated, g)
	}
	if !found {
		return fmt.Errorf("group '%s' not found", name)
	}
	if updated == nil {
		updated = []Group{}
	}
	return WriteGroups(path, updated)
}

// GroupSetActive sets the active (allowed) domain for the named group.
// Pass domain="" to reset (block all group members again).
// Returns an error if the domain is not a member of the group.
func GroupSetActive(path, name, domain string) error {
	groups, err := ReadGroups(path)
	if err != nil {
		return err
	}

	domain = normalizeDomain(domain)

	found := false
	for i, g := range groups {
		if g.Name != name {
			continue
		}
		found = true

		// If resetting, allow empty domain
		if domain != "" {
			memberFound := false
			for _, d := range g.Domains {
				if d == domain {
					memberFound = true
					break
				}
			}
			if !memberFound {
				return fmt.Errorf("domain '%s' is not a member of group '%s'", domain, name)
			}
		}

		groups[i].Active = domain
		break
	}

	if !found {
		return fmt.Errorf("group '%s' not found", name)
	}

	return WriteGroups(path, groups)
}

// GroupGetBlockedDomains returns the list of domains that should be blocked
// because of group rules: for each group, every domain EXCEPT the active one.
func GroupGetBlockedDomains(groups []Group) []string {
	var blocked []string
	for _, g := range groups {
		for _, d := range g.Domains {
			if d != g.Active {
				blocked = append(blocked, d)
			}
		}
	}
	return blocked
}
