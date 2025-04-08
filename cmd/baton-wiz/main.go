package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/conductorone/baton-wiz/pkg/connector"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-wiz",
		getConnector,
		field.NewConfiguration(configurationFields, configRelations...),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getConnector(ctx context.Context, v *viper.Viper) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	clientID := v.GetString(clientIDField.FieldName)
	clientSecret := v.GetString(clientSecretField.FieldName)
	endpointURL := v.GetString(endpointURL.FieldName)
	authURL := v.GetString(authURL.FieldName)
	audience := v.GetString(audience.FieldName)
	resourceIDs := v.GetStringSlice(resourceIDs.FieldName)
	resourceTags := v.GetString(tags.FieldName)
	resourceTypes := v.GetStringSlice(resourceTypes.FieldName)
	syncIdentities := v.GetBool(syncIdentities.FieldName)
	syncServiceUsers := v.GetBool(syncServiceUsers.FieldName)
	externalSyncMode := v.GetBool(externalSyncMode.FieldName)
	projectID := v.GetString(projectID.FieldName)

	cb, err := connector.New(ctx, &connector.Config{
		ClientID:            clientID,
		ClientSecret:        clientSecret,
		EndpointURL:         endpointURL,
		AuthURL:             authURL,
		Audience:            audience,
		ResourceIDs:         resourceIDs,
		ResourceTags:        resourceTags,
		ResourceTypes:       resourceTypes,
		SyncIdentities:      syncIdentities,
		SyncServiceAccounts: syncServiceUsers,
		ExternalSyncMode:    externalSyncMode,
		ProjectID:           projectID,
	})
	if err != nil {
		l.Error("wiz-connector: error creating connector", zap.Error(err))
		return nil, err
	}
	connector, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("wiz-connector: error creating connector", zap.Error(err))
		return nil, err
	}
	return connector, nil
}
