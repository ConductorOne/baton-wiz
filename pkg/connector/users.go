package connector

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-wiz/pkg/client"
)

type userBuilder struct {
	client *client.Client
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var rv []*v2.Resource
	usersWithAccess, nextPageToken, err := o.client.ListUsersWithAccessToResources(ctx, pToken)
	if err != nil {
		return nil, "", nil, err
	}

	for _, n := range usersWithAccess.Data.EntityEffectiveAccessEntries.Nodes {
		user := n.GrantedEntity
		primaryEmail := user.Properties.PrimaryEmail
		if primaryEmail == "" {
			primaryEmail = user.Properties.Email
		}

		firstName, lastName := rs.SplitFullName(user.Name)
		profile := map[string]interface{}{
			"login":      primaryEmail,
			"user_id":    user.Id,
			"first_name": firstName,
			"last_name":  lastName,
		}

		userTraitOptions := []rs.UserTraitOption{
			rs.WithEmail(primaryEmail, true),
			rs.WithUserLogin(primaryEmail),
			rs.WithUserProfile(profile),
		}

		if user.Properties.Enabled != nil {
			if *user.Properties.Enabled {
				userTraitOptions = append(userTraitOptions, rs.WithStatus(v2.UserTrait_Status_STATUS_ENABLED))
			} else {
				userTraitOptions = append(userTraitOptions, rs.WithStatus(v2.UserTrait_Status_STATUS_DISABLED))
			}
		}

		if user.Type == client.GrantedEntityTypeServiceAccount {
			userTraitOptions = append(userTraitOptions, rs.WithAccountType(v2.UserTrait_ACCOUNT_TYPE_SERVICE))
		}

		for _, email := range user.Properties.Emails {
			if email != primaryEmail {
				userTraitOptions = append(userTraitOptions, rs.WithEmail(email, false))
			}
		}

		userId := primaryEmail
		if userId == "" {
			userId = user.Id
		}

		resource, err := rs.NewUserResource(
			user.Name,
			userResourceType,
			userId,
			userTraitOptions,
		)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, resource)
	}

	return rv, nextPageToken, nil, nil
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(client *client.Client) *userBuilder {
	return &userBuilder{client: client}
}
