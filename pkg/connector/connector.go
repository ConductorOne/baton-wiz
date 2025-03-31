package connector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-wiz/pkg/client"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

var resourceTagErr = errors.New(`error parsing resource tags, format should be [{"key":"key1","val":"val1"}, {"key":"key2","val":"val2"}]`)

type Config struct {
	ClientID            string
	ClientSecret        string
	EndpointURL         string
	AuthURL             string
	Audience            string
	ResourceIDs         []string
	ResourceTags        string
	ResourceTypes       []string
	SyncIdentities      bool
	SyncServiceAccounts bool
}

type Connector struct {
	Client *client.Client
	Config *Config
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		// TODO(lauren) check mode before doing these changes to not change current wiz connector behavior
		// newUserBuilder(d.Client),
		newResourceBuilder(d.Client),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *Connector) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *Connector) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Wiz Graph Query Connector",
		Description: "Connector syncing specified wiz resources and users that have access to those resources",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	err := d.Client.Authorize(ctx, d.Config.AuthURL, d.Config.ClientID, d.Config.ClientSecret, d.Config.Audience)
	if err != nil {
		return nil, fmt.Errorf("wiz-connector: error authorizing: %w", err)
	}
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, config *Config) (*Connector, error) {
	l := ctxzap.Extract(ctx)
	var resourceTags []*client.ResourceTag

	if config.ResourceTags != "" {
		err := json.Unmarshal([]byte(config.ResourceTags), &resourceTags)
		if err != nil {
			return nil, resourceTagErr
		}
	}

	for _, rt := range resourceTags {
		if rt.Key == "" || rt.Value == "" {
			return nil, resourceTagErr
		}
	}

	cli, err := client.New(ctx,
		config.ClientID,
		config.ClientSecret,
		config.Audience,
		config.AuthURL,
		config.EndpointURL,
		config.ResourceIDs,
		resourceTags,
		config.ResourceTypes,
		config.SyncIdentities,
		config.SyncServiceAccounts,
	)
	if err != nil {
		l.Error("wiz-connector: failed to read token response", zap.Error(err))
		return nil, err
	}

	return &Connector{Client: cli, Config: config}, nil
}
