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
	client *client.Client
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
	for _, n := range resources.Data.EntityEffectiveAccessEntries.Nodes {
		accessibleResource := n.AccessibleResource
		// TODO(lauren) should we contain type in display name?
		displayName := accessibleResource.Name + " " + strings.ToLower(accessibleResource.Type)
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
		for _, at := range n.AccessTypes {
			ent := o.resourceEntitlement(resource, at)
			rv = append(rv, ent)
		}
	}
	return rv, nextPageToken, nil, nil
}

func (o *resourceBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var rv []*v2.Grant
	resourcePermissions, nextPageToken, err := o.client.ListResourcePermissions(ctx, resource.Id.Resource, pToken)
	if err != nil {
		return nil, "", nil, err
	}

	nodes := resourcePermissions.Data.EntityEffectiveAccessEntries.Nodes
	for _, n := range nodes {
		if n.GrantedEntity == nil {
			continue
		}

		principal := &v2.ResourceId{
			ResourceType: userResourceType.Id,
			Resource:     n.GrantedEntity.Id,
		}

		for _, at := range n.AccessTypes {
			rv = append(rv, sdkGrant.NewGrant(resource, at, principal))
		}
	}
	return rv, nextPageToken, nil, nil
}

func (o *resourceBuilder) resourceEntitlement(resource *v2.Resource, accessType string) *v2.Entitlement {
	return sdkEntitlement.NewPermissionEntitlement(resource, accessType,
		sdkEntitlement.WithGrantableTo(userResourceType),
		sdkEntitlement.WithDisplayName(fmt.Sprintf("%s Resource", resource.DisplayName)),
		sdkEntitlement.WithDescription(fmt.Sprintf("Has %s access on the %s resource", accessType, resource.DisplayName)),
	)
}

func newResourceBuilder(client *client.Client) *resourceBuilder {
	return &resourceBuilder{client: client}
}