package client

import (
	"context"
	"net/http"
	"net/url"
)

// TeamProjectRole is a project-scoped role override attached to a team.
type TeamProjectRole struct {
	Project                  string   `json:"project"`
	Role                     string   `json:"role"`
	LimitAccessByEnvironment bool     `json:"limitAccessByEnvironment"`
	Environments             []string `json:"environments"`
	Teams                    []string `json:"teams,omitempty"`
}

// Team mirrors the GrowthBook Team model.
type Team struct {
	ID                       string            `json:"id,omitempty"`
	Name                     string            `json:"name"`
	CreatedBy                string            `json:"createdBy,omitempty"`
	Description              string            `json:"description"`
	Role                     string            `json:"role"`
	LimitAccessByEnvironment bool              `json:"limitAccessByEnvironment"`
	Environments             []string          `json:"environments"`
	ProjectRoles             []TeamProjectRole `json:"projectRoles,omitempty"`
	Members                  []string          `json:"members,omitempty"`
	ManagedByIdp             bool              `json:"managedByIdp,omitempty"`
	DefaultProject           string            `json:"defaultProject,omitempty"`
	DateCreated              string            `json:"dateCreated,omitempty"`
	DateUpdated              string            `json:"dateUpdated,omitempty"`
}

// TeamInput is the create/update request body for a team. Member management is
// handled separately via AddTeamMembers / RemoveTeamMembers.
type TeamInput struct {
	Name                     string            `json:"name"`
	Description              string            `json:"description"`
	Role                     string            `json:"role"`
	LimitAccessByEnvironment *bool             `json:"limitAccessByEnvironment,omitempty"`
	Environments             []string          `json:"environments,omitempty"`
	ProjectRoles             []TeamProjectRole `json:"projectRoles,omitempty"`
	DefaultProject           *string           `json:"defaultProject,omitempty"`
}

type teamEnvelope struct {
	Team *Team `json:"team"`
}

type teamMembersInput struct {
	Members []string `json:"members"`
}

// CreateTeam creates a team.
func (c *Client) CreateTeam(ctx context.Context, in TeamInput) (*Team, error) {
	var out teamEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/teams", in, &out); err != nil {
		return nil, err
	}
	return out.Team, nil
}

// GetTeam fetches a team by ID.
func (c *Client) GetTeam(ctx context.Context, id string) (*Team, error) {
	var out teamEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/teams/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.Team, nil
}

// UpdateTeam updates a team by ID (PUT).
func (c *Client) UpdateTeam(ctx context.Context, id string, in TeamInput) (*Team, error) {
	var out teamEnvelope
	if err := c.doJSON(ctx, http.MethodPut, "/teams/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.Team, nil
}

// DeleteTeam deletes a team by ID.
func (c *Client) DeleteTeam(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/teams/"+url.PathEscape(id), nil, nil)
}

// AddTeamMembers adds the given member user IDs to a team.
func (c *Client) AddTeamMembers(ctx context.Context, id string, members []string) error {
	if len(members) == 0 {
		return nil
	}
	return c.doJSON(ctx, http.MethodPost, "/teams/"+url.PathEscape(id)+"/members", teamMembersInput{Members: members}, nil)
}

// RemoveTeamMembers removes the given member user IDs from a team.
func (c *Client) RemoveTeamMembers(ctx context.Context, id string, members []string) error {
	if len(members) == 0 {
		return nil
	}
	return c.doJSON(ctx, http.MethodDelete, "/teams/"+url.PathEscape(id)+"/members", teamMembersInput{Members: members}, nil)
}
