package client

import (
	"context"
	"net/http"
	"net/url"
)

// SDKConnection mirrors the GrowthBook SdkConnection model.
type SDKConnection struct {
	ID                            string   `json:"id"`
	Name                          string   `json:"name"`
	Organization                  string   `json:"organization,omitempty"`
	Languages                     []string `json:"languages"`
	SDKVersion                    string   `json:"sdkVersion,omitempty"`
	Environment                   string   `json:"environment"`
	Projects                      []string `json:"projects,omitempty"`
	EncryptPayload                bool     `json:"encryptPayload"`
	EncryptionKey                 string   `json:"encryptionKey,omitempty"`
	IncludeVisualExperiments      bool     `json:"includeVisualExperiments"`
	IncludeDraftExperiments       bool     `json:"includeDraftExperiments"`
	IncludeDraftExperimentRefs    bool     `json:"includeDraftExperimentRefs"`
	IncludeExperimentNames        bool     `json:"includeExperimentNames"`
	IncludeRedirectExperiments    bool     `json:"includeRedirectExperiments"`
	IncludeRuleIDs                bool     `json:"includeRuleIds"`
	IncludeProjectIDInMetadata    bool     `json:"includeProjectIdInMetadata"`
	IncludeCustomFieldsInMetadata bool     `json:"includeCustomFieldsInMetadata"`
	AllowedCustomFieldsInMetadata []string `json:"allowedCustomFieldsInMetadata,omitempty"`
	IncludeTagsInMetadata         bool     `json:"includeTagsInMetadata"`
	Key                           string   `json:"key,omitempty"`
	ProxyEnabled                  bool     `json:"proxyEnabled"`
	ProxyHost                     string   `json:"proxyHost,omitempty"`
	ProxySigningKey               string   `json:"proxySigningKey,omitempty"`
	SSEEnabled                    bool     `json:"sseEnabled"`
	HashSecureAttributes          bool     `json:"hashSecureAttributes"`
	RemoteEvalEnabled             bool     `json:"remoteEvalEnabled"`
	SavedGroupReferencesEnabled   bool     `json:"savedGroupReferencesEnabled"`
}

// SDKConnectionInput is the create/update request body. Pointer fields are
// omitted from the JSON when nil so server defaults apply.
type SDKConnectionInput struct {
	Name                          string   `json:"name,omitempty"`
	Language                      string   `json:"language,omitempty"`
	SDKVersion                    *string  `json:"sdkVersion,omitempty"`
	Environment                   string   `json:"environment,omitempty"`
	Projects                      []string `json:"projects,omitempty"`
	EncryptPayload                *bool    `json:"encryptPayload,omitempty"`
	IncludeVisualExperiments      *bool    `json:"includeVisualExperiments,omitempty"`
	IncludeDraftExperiments       *bool    `json:"includeDraftExperiments,omitempty"`
	IncludeDraftExperimentRefs    *bool    `json:"includeDraftExperimentRefs,omitempty"`
	IncludeExperimentNames        *bool    `json:"includeExperimentNames,omitempty"`
	IncludeRedirectExperiments    *bool    `json:"includeRedirectExperiments,omitempty"`
	IncludeRuleIDs                *bool    `json:"includeRuleIds,omitempty"`
	IncludeProjectIDInMetadata    *bool    `json:"includeProjectIdInMetadata,omitempty"`
	IncludeCustomFieldsInMetadata *bool    `json:"includeCustomFieldsInMetadata,omitempty"`
	AllowedCustomFieldsInMetadata []string `json:"allowedCustomFieldsInMetadata,omitempty"`
	IncludeTagsInMetadata         *bool    `json:"includeTagsInMetadata,omitempty"`
	ProxyEnabled                  *bool    `json:"proxyEnabled,omitempty"`
	ProxyHost                     *string  `json:"proxyHost,omitempty"`
	HashSecureAttributes          *bool    `json:"hashSecureAttributes,omitempty"`
	RemoteEvalEnabled             *bool    `json:"remoteEvalEnabled,omitempty"`
	SavedGroupReferencesEnabled   *bool    `json:"savedGroupReferencesEnabled,omitempty"`
}

type sdkConnectionEnvelope struct {
	SDKConnection *SDKConnection `json:"sdkConnection"`
}

// CreateSDKConnection creates an SDK connection.
func (c *Client) CreateSDKConnection(ctx context.Context, in SDKConnectionInput) (*SDKConnection, error) {
	var out sdkConnectionEnvelope
	if err := c.doJSON(ctx, http.MethodPost, "/sdk-connections", in, &out); err != nil {
		return nil, err
	}
	return out.SDKConnection, nil
}

// GetSDKConnection fetches an SDK connection by ID.
func (c *Client) GetSDKConnection(ctx context.Context, id string) (*SDKConnection, error) {
	var out sdkConnectionEnvelope
	if err := c.doJSON(ctx, http.MethodGet, "/sdk-connections/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return out.SDKConnection, nil
}

// UpdateSDKConnection updates an SDK connection by ID.
func (c *Client) UpdateSDKConnection(ctx context.Context, id string, in SDKConnectionInput) (*SDKConnection, error) {
	var out sdkConnectionEnvelope
	if err := c.doJSON(ctx, http.MethodPut, "/sdk-connections/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return out.SDKConnection, nil
}

// DeleteSDKConnection deletes an SDK connection by ID.
func (c *Client) DeleteSDKConnection(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/sdk-connections/"+url.PathEscape(id), nil, nil)
}
