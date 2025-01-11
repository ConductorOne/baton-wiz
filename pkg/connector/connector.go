package connector

import (
	"context"
	"fmt"
	"io"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-wiz/pkg/client"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type Config struct {
	ClientID     string
	ClientSecret string
	EndpointURL  string
	AuthURL      string
	Audience     string
	ResourceIDs  []string
	ResourceTags []string
}

type Connector struct {
	Client *client.Client
	Config *Config
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(d.Client),
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
	cli, err := client.New(ctx, config.ClientID, config.ClientSecret, config.Audience, config.AuthURL, config.EndpointURL, config.ResourceIDs, config.ResourceTags)
	if err != nil {
		l.Error("wiz-connector: failed to read token response", zap.Error(err))
		return nil, err
	}

	return &Connector{Client: cli, Config: config}, nil
}
