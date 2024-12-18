package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const userQuery = `query CloudEntitlementsTable($after: String, $first: Int, $filterBy: EntityEffectiveAccessFilters) {
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
}
fragment EntityEffectiveAccessGraphChartEntity on GraphEntity {
  id
  name
  type
  properties
}`

const resourceQuery = `query CloudEntitlementsTable($after: String, $first: Int, $filterBy: EntityEffectiveAccessFilters) {
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
 accessibleResource {
    ...EntityEffectiveAccessGraphChartEntity
  }
}
fragment EntityEffectiveAccessGraphChartEntity on GraphEntity {
  id
  name
  type
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
  accessTypes
}

fragment EntityEffectiveAccessGraphChartEntity on GraphEntity {
  id
  name
  type
  properties
}`

const DefaultPageSize = 1000

var grantedEntityTypeFilter = []string{"IDENTITY", "USER_ACCOUNT", "SERVICE_ACCOUNT"}

type Client struct {
	baseHttpClient *uhttp.BaseHttpClient
	BearerToken    string
	BaseUrl        *url.URL
	resourceIDs    []string
}

func New(
	ctx context.Context,
	clientId string,
	clientSecret string,
	audience string,
	authUrl string,
	endpointUrlPath string,
	resourceIDs []string,
) (*Client, error) {
	l := ctxzap.Extract(ctx)
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, l))
	if err != nil {
		l.Error("baton-wiz: failed to create http client", zap.Error(err))
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
		return fmt.Errorf("baton-wiz: error parsing auth url: %w", err)
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
		return fmt.Errorf("baton-wiz: error authorizing: %w", err)
	}
	defer resp.Body.Close()

	c.BearerToken = at.AccessToken

	return nil
}

func (c *Client) ListUsersWithAccessToResources(ctx context.Context, pToken *pagination.Token) (*UsersWithAccessQueryResponse, string, error) {
	bag, page, err := parsePageToken(pToken.Token)
	if err != nil {
		return nil, "", fmt.Errorf("wiz-connector: failed to parse page token: %w", err)
	}

	variables := map[string]interface{}{
		"first": DefaultPageSize,
		"after": page,
		"filterBy": map[string]interface{}{
			"grantedEntityType": map[string]interface{}{
				"equals": grantedEntityTypeFilter,
			},
			"resource": map[string]interface{}{
				"id": map[string]interface{}{
					"equals": c.resourceIDs,
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
		return nil, "", fmt.Errorf("wiz-connector: failed to list resources: %w", err)
	}
	defer resp.Body.Close()

	var nextPageToken string
	if res.Data.EntityEffectiveAccessEntries.PageInfo.HasNextPage {
		err = bag.Next(res.Data.EntityEffectiveAccessEntries.PageInfo.EndCursor)
		if err != nil {
			return nil, "", fmt.Errorf("wiz-connector: failed to fetch bag.Next: %w", err)
		}
		nextPageToken, err = bag.Marshal()
		if err != nil {
			return nil, "", err
		}
	}
	return res, nextPageToken, nil
}

func (c *Client) ListResources(ctx context.Context, pToken *pagination.Token) (*ResourceResponse, string, error) {
	bag, page, err := parsePageToken(pToken.Token)
	if err != nil {
		return nil, "", fmt.Errorf("wiz-connector: failed to parse page token: %w", err)
	}

	variables := map[string]interface{}{
		"first": DefaultPageSize,
		"after": page,
		"filterBy": map[string]interface{}{
			"grantedEntityType": map[string]interface{}{
				"equals": grantedEntityTypeFilter,
			},
			"resource": map[string]interface{}{
				"id": map[string]interface{}{
					"equals": c.resourceIDs,
				},
			},
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
	if res.Data.EntityEffectiveAccessEntries.PageInfo.HasNextPage {
		err = bag.Next(res.Data.EntityEffectiveAccessEntries.PageInfo.EndCursor)
		if err != nil {
			return nil, "", fmt.Errorf("wiz-connector: failed to fetch bag.Next: %w", err)
		}
		nextPageToken, err = bag.Marshal()
		if err != nil {
			return nil, "", err
		}
	}

	return res, nextPageToken, nil
}

func (c *Client) ListResourcePermissions(ctx context.Context, resourceId string, pToken *pagination.Token) (*ResourcePermissions, string, error) {
	bag, page, err := parsePageToken(pToken.Token)
	if err != nil {
		return nil, "", fmt.Errorf("wiz-connector: failed to parse page token: %w", err)
	}
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
		err = bag.Next(res.Data.EntityEffectiveAccessEntries.PageInfo.EndCursor)
		if err != nil {
			return nil, "", fmt.Errorf("wiz-connector: failed to fetch bag.Next: %w", err)
		}
		nextPageToken, err = bag.Marshal()
		if err != nil {
			return nil, "", err
		}
	}

	return res, nextPageToken, nil
}

func WithBearerToken(token string) uhttp.RequestOption {
	return uhttp.WithHeader("Authorization", fmt.Sprintf("Bearer %s", token))
}

func parsePageToken(token string) (*pagination.Bag, string, error) {
	b := &pagination.Bag{}
	err := b.Unmarshal(token)
	if err != nil {
		return nil, "", err
	}

	pageToken := ConvertPageToken(token)
	if b.Current() == nil {
		b.Push(pagination.PageState{
			ResourceID: pageToken,
		})
	}

	page := b.PageToken()

	return b, page, nil
}

func ConvertPageToken(token string) string {
	if token == "" {
		return "{{endCursor}}"
	}
	return token
}
