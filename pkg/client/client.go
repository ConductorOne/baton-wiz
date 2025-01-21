package client

import (
	"context"
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
      ...EntityEffectiveAccessDetails
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}

fragment EntityEffectiveAccessDetails on EntityEffectiveAccess {
  grantedEntity {
    ...EntityEffectiveAccessGraphChartEntity
  }
  permissions
}

fragment EntityEffectiveAccessGraphChartEntity on GraphEntity {
  id
  name
  type
  properties
}`

const DefaultPageSize = 500
const DefaultEndCursor = "{{endCursor}}"

var grantedEntityTypeFilter = []string{"IDENTITY", "USER_ACCOUNT", "SERVICE_ACCOUNT"}

type Client struct {
	baseHttpClient *uhttp.BaseHttpClient
	BearerToken    string
	BaseUrl        *url.URL
	resourceIDs    []string
	resourceTags   []*ResourceTag
	resourceTypes  []string
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

	client := Client{
		baseHttpClient: wrapper,
		BaseUrl:        endpointUrl,
		resourceIDs:    resourceIDs,
		resourceTags:   resourceTags,
		resourceTypes:  resourceTypes,
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
	bag, page, err := parseUserPageToken(pToken.Token, c.resourceIDs)
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
			return nil, "", err
		}

		err = bag.Next(resourceNextPage)
		if err != nil {
			return nil, "", err
		}

		resourceIdSet := mapset.NewSet[string]()

		for _, n := range resources.Data.GraphSearch.Nodes {
			for _, accessibleResource := range n.Entities {
				if resourceIdSet.ContainsOne(accessibleResource.Id) {
					continue
				}
				resourceIdSet.Add(accessibleResource.Id)
				bag.Push(pagination.PageState{
					ResourceID:     accessibleResource.Id,
					Token:          DefaultEndCursor,
					ResourceTypeID: ListUsersResourceTypeResourceID,
				})
			}
		}
		resourceNextPageMarshal, err := bag.Marshal()
		if err != nil {
			return nil, "", err
		}
		return &UsersWithAccessQueryResponse{}, resourceNextPageMarshal, nil
	case ListUsersResourceTypeResourceID:
		variables := map[string]interface{}{
			"first": DefaultPageSize,
			"after": page,
			"filterBy": map[string]interface{}{
				"grantedEntityType": map[string]interface{}{
					"equals": grantedEntityTypeFilter,
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
			return nil, "", fmt.Errorf("wiz-connector: failed to list users with access to resources: %w", err)
		}
		defer resp.Body.Close()

		var nextPageToken string
		if res.Data.EntityEffectiveAccessEntries.PageInfo.HasNextPage {
			err = bag.Next(res.Data.EntityEffectiveAccessEntries.PageInfo.EndCursor)
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
		"projectId": "*",
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
	page := getEndCursor(pToken.Token)

	variables := map[string]interface{}{
		"first": DefaultPageSize,
		"after": page,
		"filterBy": map[string]interface{}{
			"grantedEntityType": map[string]interface{}{
				"equals": grantedEntityTypeFilter,
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
		return nil, "", fmt.Errorf("wiz-connector: failed to list resources permissions: %w", err)
	}
	defer resp.Body.Close()

	var nextPageToken string
	if res.Data.EntityEffectiveAccessEntries.PageInfo.HasNextPage {
		nextPageToken = res.Data.EntityEffectiveAccessEntries.PageInfo.EndCursor
	}

	return res, nextPageToken, nil
}

func WithBearerToken(token string) uhttp.RequestOption {
	return uhttp.WithHeader("Authorization", fmt.Sprintf("Bearer %s", token))
}

func parseUserPageToken(token string, resourceIDs []string) (*pagination.Bag, string, error) {
	b := &pagination.Bag{}
	err := b.Unmarshal(token)
	if err != nil {
		return nil, "", err
	}

	if b.Current() == nil {
		if len(resourceIDs) != 0 {
			for _, resourceID := range resourceIDs {
				b.Push(pagination.PageState{
					ResourceID:     resourceID,
					Token:          DefaultEndCursor,
					ResourceTypeID: ListUsersResourceTypeResourceID,
				})
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

func getEndCursor(token string) string {
	if token == "" {
		return DefaultEndCursor
	}
	return token
}
