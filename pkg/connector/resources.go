package connector

import (
	"context"
	"fmt"
	"strings"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	sdkEntitlement "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	sdkGrant "github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-wiz/pkg/client"
)

type resourceBuilder struct {
	client           *client.Client
	externalSyncMode bool
}

func (o *resourceBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return wizQueryResourceType
}

func (o *resourceBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var rv []*v2.Resource
	resources, nextPageToken, err := o.client.ListResources(ctx, pToken)
	if err != nil {
		return nil, "", nil, err
	}

	for _, n := range resources.Data.GraphSearch.Nodes {
		for _, accessibleResource := range n.Entities {
			displayName := fmt.Sprintf("%s %s", accessibleResource.Name, strings.ToLower(accessibleResource.Type))
			resource, err := rs.NewResource(
				displayName,
				wizQueryResourceType,
				accessibleResource.Id,
			)
			if err != nil {
				return nil, "", nil, err
			}
			rv = append(rv, resource)
		}
	}

	return rv, nextPageToken, nil, nil
}

func (o *resourceBuilder) Entitlements(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement
	resourcePermissions, nextPageToken, err := o.client.ListResourcePermissions(ctx, resource.Id.Resource, pToken)
	if err != nil {
		return nil, "", nil, err
	}
	nodes := resourcePermissions.Data.EntityEffectiveAccessEntries.Nodes
	for _, n := range nodes {
		for _, p := range n.Permissions {
			ent := o.resourceEntitlement(resource, p)
			rv = append(rv, ent)
		}
	}
	return rv, nextPageToken, nil, nil
}

func (o *resourceBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var rv []*v2.Grant
	resourcePermissions, nextPageToken, err := o.client.ListResourcePermissionEffectiveAccess(ctx, resource.Id.Resource, pToken)
	if err != nil {
		return nil, "", nil, err
	}

	nodes := resourcePermissions.Data.EntityEffectiveAccessEntries.Nodes
	for _, n := range nodes {
		grantedEntity := n.GrantedEntity
		if grantedEntity == nil {
			continue
		}

		var resourceType string
		if grantedEntity.Type == client.GrantedEntityTypeGroup {
			resourceType = groupResourceType.Id
		} else {
			resourceType = userResourceType.Id
		}

		// TODO(lauren) Check if resource ID should be grantedEntity.Properties.ExternalId when in externalSyncMode
		principal := &v2.ResourceId{
			ResourceType: resourceType,
			Resource:     grantedEntity.Id,
		}

		grantOpts := make([]sdkGrant.GrantOption, 0)
		if o.externalSyncMode {
			// TODO(lauren) do we want to exclude entities with no external id when in this mode?
			// If so, consider adding a filter to graphql query
			if grantedEntity.Properties.ExternalId == "" {
				continue
			}
			grantOpts = append(grantOpts, sdkGrant.WithAnnotation(&v2.ExternalResourceMatchID{
				Id: grantedEntity.Properties.ExternalId,
			}))
		} else {
			if grantedEntity.Properties.PrimaryEmail != "" {
				principal.Resource = grantedEntity.Properties.PrimaryEmail
			} else if grantedEntity.Properties.Email != "" {
				principal.Resource = grantedEntity.Properties.Email
			}
		}

		for _, p := range n.Permissions {
			rv = append(rv, sdkGrant.NewGrant(resource, p, principal, grantOpts...))
		}
	}

	return rv, nextPageToken, nil, nil
}

func (o *resourceBuilder) resourceEntitlement(resource *v2.Resource, accessType string) *v2.Entitlement {
	grantableTo := []*v2.ResourceType{userResourceType}
	if o.externalSyncMode {
		grantableTo = append(grantableTo, groupResourceType)
	}
	return sdkEntitlement.NewPermissionEntitlement(resource, accessType,
		sdkEntitlement.WithGrantableTo(grantableTo...),
		sdkEntitlement.WithDisplayName(fmt.Sprintf("%s Resource", resource.DisplayName)),
		sdkEntitlement.WithDescription(fmt.Sprintf("Has %s access on the %s resource", accessType, resource.DisplayName)),
	)
}

func newResourceBuilder(client *client.Client, externalSyncMode bool) *resourceBuilder {
	return &resourceBuilder{client: client, externalSyncMode: externalSyncMode}
}
