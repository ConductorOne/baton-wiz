package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	clientIDField     = field.StringField("wiz-client-id", field.WithRequired(true), field.WithDescription("The client ID used to authenticate with Wiz"))
	clientSecretField = field.StringField("wiz-client-secret", field.WithRequired(true), field.WithDescription("The client secret used to authenticate with Wiz"))
	endpointURL       = field.StringField("endpoint-url", field.WithRequired(true), field.WithDescription("The endpoint url used to authenticate with Wiz"))
	authURL           = field.StringField("auth-url", field.WithRequired(true), field.WithDescription("The auth url used to authenticate with Wiz"))
	audience          = field.StringField("audience", field.WithDefaultValue("wiz-api"), field.WithDescription("The audience used to authenticate with Wiz"))
	resourceIDs       = field.StringSliceField("resource-ids", field.WithDescription("The resource ids to sync"))
	tags              = field.StringField("tags", field.WithDescription("The tags on resources to sync"))
	resourceTypes     = field.StringSliceField("wiz-resource-types", field.WithDescription("The wiz resource-types to sync"))
	syncIdentities    = field.BoolField("sync-identities", field.WithDescription("Enable if wiz identities should be synced"))
	syncServiceUsers  = field.BoolField("sync-service-accounts", field.WithDescription("Enable if wiz service accounts should be synced"))
	externalSyncMode  = field.BoolField("external-sync-mode", field.WithDescription("Enable external sync mode"))
	projectID         = field.StringField("project-id",
		field.WithDescription("Scope the resource graph query to a specific project. Required if service account does not have access to all projects."))

	configurationFields = []field.SchemaField{
		clientIDField, clientSecretField, endpointURL, authURL, audience, resourceIDs, tags, resourceTypes, syncIdentities, syncServiceUsers, externalSyncMode, projectID,
	}
)

var configRelations = []field.SchemaFieldRelationship{
	field.FieldsAtLeastOneUsed(resourceIDs, tags),
	field.FieldsMutuallyExclusive(resourceIDs, tags),
}
