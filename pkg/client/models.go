package client

import (
	"encoding/json"
	"fmt"
)

type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type Emails []string

// Emails can be a single string or string array.
func (e *Emails) UnmarshalJSON(data []byte) error {
	var singleEmail string
	if err := json.Unmarshal(data, &singleEmail); err == nil {
		*e = Emails{singleEmail}
		return nil
	}

	var multipleEmails []string
	if err := json.Unmarshal(data, &multipleEmails); err == nil {
		*e = multipleEmails
		return nil
	}

	return fmt.Errorf("emails should be a string or a slice of strings")
}

type GrantedEntity struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Properties struct {
		VertexID     string `json:"_vertexID"`
		Email        string `json:"email"`
		Emails       Emails `json:"emails,omitempty"`
		Name         string `json:"name"`
		NativeType   string `json:"nativeType"`
		PrimaryEmail string `json:"primaryEmail"`
		Enabled      *bool  `json:"accountEnabled"`
		ExternalId   string `json:"externalId"`
	} `json:"properties"`
}

type UsersWithAccessQueryResponse struct {
	Data struct {
		EntityEffectiveAccessEntries struct {
			Nodes []struct {
				GrantedEntity *GrantedEntity `json:"grantedEntity"`
			} `json:"nodes"`
			PageInfo PageInfo `json:"pageInfo"`
		} `json:"entityEffectiveAccessEntries"`
	} `json:"data"`
}

type ResourcePermissions struct {
	Data struct {
		EntityEffectiveAccessEntries struct {
			Nodes []struct {
				GrantedEntity *GrantedEntity `json:"grantedEntity"`
				AccessTypes   []string       `json:"accessTypes"`
			} `json:"nodes"`
			PageInfo PageInfo `json:"pageInfo"`
		} `json:"entityEffectiveAccessEntries"`
	} `json:"data"`
}

type ResourceResponse struct {
	Data struct {
		GraphSearch struct {
			Nodes []struct {
				Entities []struct {
					Id   string `json:"id"`
					Name string `json:"name"`
					Type string `json:"type"`
				} `json:"entities"`
			} `json:"nodes"`
			PageInfo PageInfo `json:"pageInfo"`
		} `json:"graphSearch"`
	} `json:"data"`
}
