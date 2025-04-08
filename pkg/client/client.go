package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const ListUsersResourceTypeResourceID = "resourceID"
const ListUsersResourceTypeResourceTag = "resourceTag"

const userQuery = `query CloudEntitlementsTable($after: String, $first: Int, $filterBy: EntityEffectiveAccessFilters) {
  entityEffectiveAccessEntries(after: $after, first: $first, filterBy: $filterBy) {
    nodes {
       grantedEntity {
        id
        name
        type
        properties
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}`

const resourceQuery = `query GraphSearch($query: GraphEntityQueryInput, $projectId: String!, $first: Int, $after: String) {
  graphSearch(
    query: $query
    projectId: $projectId
    first: $first
    after: $after
  ) {
    nodes {
      entities {
        id
        name
        type
      }
    }
    pageInfo {
      endCursor
      hasNextPage
    }
  }
}`

const resourcePermissionQuery = `query CloudEntitlementsTable($after: String, $first: Int, $filterBy: EntityEffectiveAccessFilters) {
  entityEffectiveAccessEntries(after: $after, first: $first, filterBy: $filterBy) {
    nodes {
      permissions
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}`

const resourceEffectiveAccessQuery = `query CloudEntitlementsTable($after: String, $first: Int, $filterBy: EntityEffectiveAccessFilters) {
  entityEffectiveAccessEntries(after: $after, first: $first, filterBy: $filterBy) {
    nodes {
      grantedEntity {
        id
        name
        type
        properties
        providerUniqueId
      }
      permissions
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}`

const DefaultPageSize = 500
const DefaultEndCursor = "{{endCursor}}"

const GrantedEntityTypeIdentity = "IDENTITY"
const GrantedEntityTypeUserAccount = "USER_ACCOUNT"
const GrantedEntityTypeServiceAccount = "SERVICE_ACCOUNT"
const GrantedEntityTypeGroup = "GROUP"

var grantedEntityTypeUserAccountFilter = []string{GrantedEntityTypeUserAccount}

type GrantedEntityTypeToken struct {
	GrantedEntityType string `json:"granted_entity_type"`
	Token             string `json:"token"`
}

func (gt *GrantedEntityTypeToken) Marshal() (string, error) {
	data, err := json.Marshal(gt)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type Client struct {
	baseHttpClient          *uhttp.BaseHttpClient
	BearerToken             string
	BaseUrl                 *url.URL
	resourceIDs             []string
	resourceTags            []*ResourceTag
	resourceTypes           []string
	grantedEntityTypeFilter []string
	resourceIdSet           mapset.Set[string]
	projectId               string
}

func New(
	ctx context.Context,
	clientId string,
	clientSecret string,
	audience string,
	authUrl string,
	endpointUrlPath string,
	resourceIDs []string,
	resourceTags []*ResourceTag,
	resourceTypes []string,
	syncIdentities bool,
	syncServiceAccounts bool,
	externalSyncMode bool,
	projectId string,
) (*Client, error) {
	l := ctxzap.Extract(ctx)
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, l))
	if err != nil {
		l.Error("wiz-connector: failed to create http client", zap.Error(err))
		return nil, err
	}

	wrapper, err := uhttp.NewBaseHttpClientWithContext(ctx, httpClient)
	if err != nil {
		return nil, err
	}

	endpointUrl, err := url.Parse(endpointUrlPath)
	if err != nil {
		return nil, err
	}

	grantedEntityTypeFilter := grantedEntityTypeUserAccountFilter
	if syncServiceAccounts {
		grantedEntityTypeFilter = append(grantedEntityTypeFilter, GrantedEntityTypeServiceAccount)
	}
	if syncIdentities {
		grantedEntityTypeFilter = append(grantedEntityTypeFilter, GrantedEntityTypeIdentity)
	}
	if externalSyncMode {
		grantedEntityTypeFilter = append(grantedEntityTypeFilter, GrantedEntityTypeGroup)
	}

	if projectId == "" {
		projectId = "*"
	}

	client := Client{
		baseHttpClient:          wrapper,
		BaseUrl:                 endpointUrl,
		resourceIDs:             resourceIDs,
		resourceTags:            resourceTags,
		resourceTypes:           resourceTypes,
		grantedEntityTypeFilter: grantedEntityTypeFilter,
		resourceIdSet:           mapset.NewSet[string](),
		projectId:               projectId,
	}

	err = client.Authorize(ctx, authUrl, clientId, clientSecret, audience)
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func (c *Client) Authorize(
	ctx context.Context,
	authUrlPath string,
	clientId string,
	clientSecret string,
	audience string,
) error {
	form := &url.Values{}
	form.Set("audience", audience)
	form.Set("client_id", clientId)
	form.Set("client_secret", clientSecret)
	form.Set("grant_type", "client_credentials")

	authUrl, err := url.Parse(authUrlPath)
	if err != nil {
		return fmt.Errorf("wiz-connector: error parsing auth url: %w", err)
	}

	request, err := c.baseHttpClient.NewRequest(ctx, http.MethodPost, authUrl, uhttp.WithFormBody(form.Encode()))
	if err != nil {
		return err
	}

	at := &oauth2.Token{}
	resp, err := c.baseHttpClient.Do(
		request,
		uhttp.WithJSONResponse(&at),
	)
	if err != nil {
		return fmt.Errorf("wiz-connector: error authorizing: %w", err)
	}
	defer resp.Body.Close()

	c.BearerToken = at.AccessToken

	return nil
}

func (c *Client) ListUsersWithAccessToResources(ctx context.Context, pToken *pagination.Token) (*UsersWithAccessQueryResponse, string, error) {
	l := ctxzap.Extract(ctx)
	bag, page, err := c.parseUserPageToken(pToken.Token, c.resourceIDs)
	if err != nil {
		return nil, "", fmt.Errorf("wiz-connector: error parsing user page token: %w", err)
	}

	switch bag.ResourceTypeID() {
	case ListUsersResourceTypeResourceTag:
		// Fetch the resources with the tags and push each resource id to the pagination bag so we can get
		// the users that have access per resource
		resourceToken := &pagination.Token{Token: page}
		resources, resourceNextPage, err := c.ListResources(ctx, resourceToken)
		if err != nil {
			l.Error("wiz-connector: failed to list resources for list users",
				zap.String("page_token", pToken.Token),
				zap.String("page", page),
				zap.Error(err))
			return nil, "", err
		}

		err = bag.Next(resourceNextPage)
		if err != nil {
			return nil, "", err
		}

		for _, n := range resources.Data.GraphSearch.Nodes {
			for _, accessibleResource := range n.Entities {
				if c.resourceIdSet.ContainsOne(accessibleResource.Id) {
					continue
				}
				c.resourceIdSet.Add(accessibleResource.Id)

				for _, gt := range c.grantedEntityTypeFilter {
					userTypeWithToken := &GrantedEntityTypeToken{
						GrantedEntityType: gt,
						Token:             DefaultEndCursor,
					}
					tokenStr, err := userTypeWithToken.Marshal()
					if err != nil {
						return nil, "", err
					}
					bag.Push(pagination.PageState{
						ResourceID:     accessibleResource.Id,
						Token:          tokenStr,
						ResourceTypeID: ListUsersResourceTypeResourceID,
					})
				}
			}
		}

		resourceNextPageMarshal, err := bag.Marshal()
		if err != nil {
			return nil, "", err
		}
		return &UsersWithAccessQueryResponse{}, resourceNextPageMarshal, nil
	case ListUsersResourceTypeResourceID:
		ut, err := parseGrantedEntityTypeToken(page)
		if err != nil {
			return nil, "", fmt.Errorf("wiz-connector: error parsing user type page token: %w, page: %s", err, page)
		}

		variables := map[string]interface{}{
			"first": DefaultPageSize,
			"after": ut.Token,
			"filterBy": map[string]interface{}{
				"grantedEntityType": map[string]interface{}{
					"equals": ut.GrantedEntityType,
				},
				"resource": map[string]interface{}{
					"id": map[string]interface{}{
						"equals": bag.ResourceID(),
					},
				},
			},
		}

		payload := map[string]interface{}{
			"query":     userQuery,
			"variables": variables,
		}

		options := []uhttp.RequestOption{
			uhttp.WithAcceptJSONHeader(),
			uhttp.WithJSONBody(payload),
			WithBearerToken(c.BearerToken),
		}

		req, err := c.baseHttpClient.NewRequest(ctx, http.MethodPost, c.BaseUrl, options...)
		if err != nil {
			return nil, "", err
		}

		res := &UsersWithAccessQueryResponse{}
		resp, err := c.baseHttpClient.Do(
			req,
			uhttp.WithJSONResponse(&res),
		)
		if err != nil {
			l.Error("wiz-connector: failed to list users with access to resources",
				zap.String("page_token", pToken.Token),
				zap.String("page", page),
				zap.String("user_token", ut.Token),
				zap.String("user_type", ut.GrantedEntityType),
				zap.Error(err))
			return nil, "", fmt.Errorf("wiz-connector: failed to list users with access to resources: %w", err)
		}
		defer resp.Body.Close()

		var nextPageToken string
		if res.Data.EntityEffectiveAccessEntries.PageInfo.HasNextPage {
			ut.Token = res.Data.EntityEffectiveAccessEntries.PageInfo.EndCursor
			userTypeTokenStr, err := ut.Marshal()
			if err != nil {
				return nil, "", fmt.Errorf("wiz-connector: error converting user type page token: %w", err)
			}
			err = bag.Next(userTypeTokenStr)
			if err != nil {
				return nil, "", fmt.Errorf("wiz-connector: failed to fetch bag.Next: %w", err)
			}
		} else {
			err = bag.Next("")
			if err != nil {
				return nil, "", fmt.Errorf("wiz-connector: failed to fetch bag.Next: %w", err)
			}
		}

		nextPageToken, err = bag.Marshal()
		if err != nil {
			return nil, "", err
		}

		return res, nextPageToken, nil
	}
	return nil, "", errors.New("wiz-connector: failed to list users: invalid pagination resource type")
}

func (c *Client) ListResources(ctx context.Context, pToken *pagination.Token) (*ResourceResponse, string, error) {
	l := ctxzap.Extract(ctx)
	page := getEndCursor(pToken.Token)

	whereClause := make(map[string]interface{}, 0)
	if len(c.resourceIDs) != 0 {
		whereClause["_vertexID"] = map[string]interface{}{
			"EQUALS": c.resourceIDs,
		}
	}

	if len(c.resourceTags) != 0 {
		tagKeyValSlice := make([]map[string]interface{}, 0)
		for _, tag := range c.resourceTags {
			tagKeyValSlice = append(tagKeyValSlice, map[string]interface{}{
				"key": tag.Key, "value": tag.Value,
			})
		}
		whereClause["tags"] = map[string]interface{}{
			"TAG_CONTAINS_ANY": tagKeyValSlice,
		}
	}

	resourceTypes := c.resourceTypes
	if len(resourceTypes) == 0 {
		resourceTypes = []string{"ANY"} // TODO(lauren) might be able to filter with CLOUD_RESOURCE
	}

	variables := map[string]interface{}{
		"first":     DefaultPageSize,
		"after":     page,
		"projectId": c.projectId,
		"query": map[string]interface{}{
			"type":  resourceTypes,
			"where": whereClause,
		},
	}
	payload := map[string]interface{}{
		"query":     resourceQuery,
		"variables": variables,
	}

	options := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithJSONBody(payload),
		WithBearerToken(c.BearerToken),
	}

	req, err := c.baseHttpClient.NewRequest(ctx, http.MethodPost, c.BaseUrl, options...)
	if err != nil {
		return nil, "", err
	}

	res := &ResourceResponse{}
	resp, err := c.baseHttpClient.Do(req, uhttp.WithJSONResponse(&res))
	if err != nil {
		l.Error("wiz-connector: failed to list resources",
			zap.String("token", pToken.Token),
			zap.String("page", page),
			zap.Error(err))
		return nil, "", fmt.Errorf("wiz-connector: failed to list resources: %w", err)
	}
	defer resp.Body.Close()

	var nextPageToken string
	if res.Data.GraphSearch.PageInfo.HasNextPage {
		nextPageToken = res.Data.GraphSearch.PageInfo.EndCursor
	}

	return res, nextPageToken, nil
}

func (c *Client) ListResourcePermissions(ctx context.Context, resourceId string, pToken *pagination.Token) (*ResourcePermissions, string, error) {
	l := ctxzap.Extract(ctx)
	bag, page, err := c.getGrantedEntityTypeToken(pToken.Token)
	if err != nil {
		return nil, "", fmt.Errorf("wiz-connector: error getting granted entity type page token: %w", err)
	}
	gt, err := parseGrantedEntityTypeToken(page)
	if err != nil {
		return nil, "", fmt.Errorf("wiz-connector: error parsing granted entity type page token: %w", err)
	}

	variables := map[string]interface{}{
		"first": DefaultPageSize,
		"after": gt.Token,
		"filterBy": map[string]interface{}{
			"grantedEntityType": map[string]interface{}{
				"equals": gt.GrantedEntityType,
			},
			"resource": map[string]interface{}{
				"id": map[string]interface{}{
					"equals": []string{resourceId},
				},
			},
		},
	}
	payload := map[string]interface{}{
		"query":     resourcePermissionQuery,
		"variables": variables,
	}

	options := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithJSONBody(payload),
		WithBearerToken(c.BearerToken),
	}

	req, err := c.baseHttpClient.NewRequest(ctx, http.MethodPost, c.BaseUrl, options...)
	if err != nil {
		return nil, "", err
	}

	res := &ResourcePermissions{}
	resp, err := c.baseHttpClient.Do(
		req,
		uhttp.WithJSONResponse(&res),
	)
	if err != nil {
		l.Error("wiz-connector: failed to list resources permissions",
			zap.String("page_token", pToken.Token),
			zap.String("page", page),
			zap.String("granted_entity_token", gt.Token),
			zap.String("granted_entity_type", gt.GrantedEntityType),
			zap.Error(err))
		return nil, "", fmt.Errorf("wiz-connector: failed to list resources permissions: %w", err)
	}
	defer resp.Body.Close()

	if res.Data.EntityEffectiveAccessEntries.PageInfo.HasNextPage {
		gt.Token = res.Data.EntityEffectiveAccessEntries.PageInfo.EndCursor
		grantedEntityTypeTokenStr, err := gt.Marshal()
		if err != nil {
			return nil, "", fmt.Errorf("wiz-connector: error converting granted entity type page token: %w", err)
		}
		err = bag.Next(grantedEntityTypeTokenStr)
		if err != nil {
			return nil, "", fmt.Errorf("wiz-connector: failed to fetch bag.Next: %w", err)
		}
	} else {
		err = bag.Next("")
		if err != nil {
			return nil, "", fmt.Errorf("wiz-connector: failed to fetch bag.Next: %w", err)
		}
	}
	nextPageToken, err := bag.Marshal()
	if err != nil {
		return nil, "", err
	}

	return res, nextPageToken, nil
}

func (c *Client) ListResourcePermissionEffectiveAccess(ctx context.Context, resourceId string, pToken *pagination.Token) (*ResourcePermissions, string, error) {
	l := ctxzap.Extract(ctx)
	bag, page, err := c.getGrantedEntityTypeToken(pToken.Token)
	if err != nil {
		return nil, "", fmt.Errorf("wiz-connector: error getting granted entity type page token: %w", err)
	}
	gt, err := parseGrantedEntityTypeToken(page)
	if err != nil {
		return nil, "", fmt.Errorf("wiz-connector: error parsing granted entity type page token: %w", err)
	}

	variables := map[string]interface{}{
		"first": DefaultPageSize,
		"after": gt.Token,
		"filterBy": map[string]interface{}{
			"grantedEntityType": map[string]interface{}{
				"equals": gt.GrantedEntityType,
			},
			"resource": map[string]interface{}{
				"id": map[string]interface{}{
					"equals": []string{resourceId},
				},
			},
		},
	}
	payload := map[string]interface{}{
		"query":     resourceEffectiveAccessQuery,
		"variables": variables,
	}

	options := []uhttp.RequestOption{
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithJSONBody(payload),
		WithBearerToken(c.BearerToken),
	}

	req, err := c.baseHttpClient.NewRequest(ctx, http.MethodPost, c.BaseUrl, options...)
	if err != nil {
		return nil, "", err
	}

	res := &ResourcePermissions{}
	resp, err := c.baseHttpClient.Do(
		req,
		uhttp.WithJSONResponse(&res),
	)
	if err != nil {
		l.Error("wiz-connector: failed to list resource permissions effective access",
			zap.String("page_token", pToken.Token),
			zap.String("page", page),
			zap.String("granted_entity_token", gt.Token),
			zap.String("granted_entity_type", gt.GrantedEntityType),
			zap.Error(err))
		return nil, "", fmt.Errorf("wiz-connector: failed to list resource permissions effective access: %w", err)
	}
	defer resp.Body.Close()

	if res.Data.EntityEffectiveAccessEntries.PageInfo.HasNextPage {
		gt.Token = res.Data.EntityEffectiveAccessEntries.PageInfo.EndCursor
		grantedEntityTypeTokenStr, err := gt.Marshal()
		if err != nil {
			return nil, "", fmt.Errorf("wiz-connector: error converting granted entity type page token: %w", err)
		}
		err = bag.Next(grantedEntityTypeTokenStr)
		if err != nil {
			return nil, "", fmt.Errorf("wiz-connector: failed to fetch bag.Next: %w", err)
		}
	} else {
		err = bag.Next("")
		if err != nil {
			return nil, "", fmt.Errorf("wiz-connector: failed to fetch bag.Next: %w", err)
		}
	}
	nextPageToken, err := bag.Marshal()
	if err != nil {
		return nil, "", err
	}

	return res, nextPageToken, nil
}

func WithBearerToken(token string) uhttp.RequestOption {
	return uhttp.WithHeader("Authorization", fmt.Sprintf("Bearer %s", token))
}

func (c *Client) parseUserPageToken(token string, resourceIDs []string) (*pagination.Bag, string, error) {
	b := &pagination.Bag{}
	err := b.Unmarshal(token)
	if err != nil {
		return nil, "", fmt.Errorf("wiz-connector: failed to unmarshal bag: %w", err)
	}

	if b.Current() == nil {
		if len(resourceIDs) != 0 {
			for _, resourceID := range resourceIDs {
				for _, ut := range c.grantedEntityTypeFilter {
					userTypeWithToken := &GrantedEntityTypeToken{
						GrantedEntityType: ut,
						Token:             DefaultEndCursor,
					}
					tokenStr, err := userTypeWithToken.Marshal()
					if err != nil {
						return nil, "", err
					}
					b.Push(pagination.PageState{
						ResourceID:     resourceID,
						Token:          tokenStr,
						ResourceTypeID: ListUsersResourceTypeResourceID,
					})
				}
			}
		} else {
			b.Push(pagination.PageState{
				Token:          DefaultEndCursor,
				ResourceTypeID: ListUsersResourceTypeResourceTag,
			})
		}
	}

	page := b.PageToken()

	return b, page, nil
}

func parseGrantedEntityTypeToken(token string) (*GrantedEntityTypeToken, error) {
	gtt := &GrantedEntityTypeToken{}
	err := json.Unmarshal([]byte(token), &gtt)
	if err != nil {
		return nil, err
	}
	return gtt, nil
}

func (c *Client) getGrantedEntityTypeToken(token string) (*pagination.Bag, string, error) {
	b := &pagination.Bag{}
	err := b.Unmarshal(token)
	if err != nil {
		return nil, "", err
	}

	if b.Current() == nil {
		for _, gt := range c.grantedEntityTypeFilter {
			grantedEntityTypeWithToken := &GrantedEntityTypeToken{
				GrantedEntityType: gt,
				Token:             DefaultEndCursor,
			}
			tokenStr, err := grantedEntityTypeWithToken.Marshal()
			if err != nil {
				return nil, "", err
			}
			b.Push(pagination.PageState{
				Token: tokenStr,
			})
		}
	}

	page := b.PageToken()

	return b, page, nil
}

func getEndCursor(token string) string {
	if token == "" {
		return DefaultEndCursor
	}
	return token
}
