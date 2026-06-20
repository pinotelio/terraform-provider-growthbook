package client

import "context"

// Member mirrors the read-only GrowthBook Member model. Membership is created
// out of band (invite/SSO), so the provider exposes members only for reading.
type Member struct {
	ID                       string   `json:"id"`
	Name                     string   `json:"name,omitempty"`
	Email                    string   `json:"email"`
	GlobalRole               string   `json:"globalRole"`
	Environments             []string `json:"environments,omitempty"`
	LimitAccessByEnvironment bool     `json:"limitAccessByEnvironment,omitempty"`
	ManagedByIDP             bool     `json:"managedbyIdp,omitempty"`
	Teams                    []string `json:"teams,omitempty"`
	LastLoginDate            string   `json:"lastLoginDate,omitempty"`
	DateCreated              string   `json:"dateCreated,omitempty"`
	DateUpdated              string   `json:"dateUpdated,omitempty"`
}

// ListMembers returns every organization member, following pagination.
func (c *Client) ListMembers(ctx context.Context) ([]Member, error) {
	return fetchAll(ctx, c, "/members", func(b []byte) ([]Member, pagination, error) {
		var page struct {
			Members []Member `json:"members"`
			pagination
		}
		if err := unmarshal(b, &page); err != nil {
			return nil, pagination{}, err
		}
		return page.Members, page.pagination, nil
	})
}
