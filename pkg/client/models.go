package client

import (
	"encoding/json"
	"fmt"
	"time"
)

type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type Emails []string

// Emails can be a single string or string array
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
		// ProviderUniqueId interface{} `json:"providerUniqueId"`
	} `json:"properties"`
}

type UsersWithAccessQueryResponse struct {
	Data struct {
		EntityEffectiveAccessEntries struct {
			Nodes []struct {
				GrantedEntity GrantedEntity `json:"grantedEntity"`
			} `json:"nodes"`
			PageInfo PageInfo `json:"pageInfo"`
		} `json:"entityEffectiveAccessEntries"`
	} `json:"data"`
}

type ResourceResponse struct {
	Data struct {
		EntityEffectiveAccessEntries struct {
			Nodes []struct {
				AccessibleResource struct {
					Id         string `json:"id"`
					Name       string `json:"name"`
					Type       string `json:"type"`
					Properties struct {
						VertexID                            string      `json:"_vertexID"`
						AccessibleFromInternet              bool        `json:"accessibleFrom.internet,omitempty"`
						CloudPlatform                       string      `json:"cloudPlatform"`
						CloudProviderURL                    string      `json:"cloudProviderURL"`
						CreationDate                        time.Time   `json:"creationDate,omitempty"`
						Encrypted                           bool        `json:"encrypted,omitempty"`
						EncryptionInTransit                 bool        `json:"encryptionInTransit,omitempty"`
						ExternalId                          string      `json:"externalId"`
						FullResourceName                    interface{} `json:"fullResourceName"`
						HasSensitiveData                    bool        `json:"hasSensitiveData,omitempty"`
						IsPublic                            bool        `json:"isPublic,omitempty"`
						LoggingEnabled                      bool        `json:"loggingEnabled,omitempty"`
						MaxExposureLevel                    int         `json:"maxExposureLevel,omitempty"`
						Name                                string      `json:"name"`
						NativeType                          string      `json:"nativeType"`
						NumAddressesOpenForHTTP             int         `json:"numAddressesOpenForHTTP,omitempty"`
						NumAddressesOpenForHTTPS            int         `json:"numAddressesOpenForHTTPS,omitempty"`
						NumAddressesOpenForNonStandardPorts int         `json:"numAddressesOpenForNonStandardPorts,omitempty"`
						NumAddressesOpenForRDP              int         `json:"numAddressesOpenForRDP,omitempty"`
						NumAddressesOpenForSSH              int         `json:"numAddressesOpenForSSH,omitempty"`
						NumAddressesOpenForWINRM            int         `json:"numAddressesOpenForWINRM,omitempty"`
						OpenToAllInternet                   bool        `json:"openToAllInternet,omitempty"`
						ProviderUniqueId                    string      `json:"providerUniqueId"`
						PublicExposure                      string      `json:"publicExposure,omitempty"`
						Region                              string      `json:"region"`
						RegionLocation                      string      `json:"regionLocation,omitempty"`
						RegionType                          string      `json:"regionType,omitempty"`
						ResourceGroupExternalId             interface{} `json:"resourceGroupExternalId"`
						RetentionPeriod                     int         `json:"retentionPeriod,omitempty"`
						Status                              string      `json:"status"`
						SubscriptionExternalId              string      `json:"subscriptionExternalId"`
						UpdatedAt                           time.Time   `json:"updatedAt"`
						VersioningEnabled                   bool        `json:"versioningEnabled,omitempty"`
						WebHostingEnabled                   bool        `json:"webHostingEnabled,omitempty"`
						Zone                                interface{} `json:"zone"`
						IsManaged                           bool        `json:"isManaged,omitempty"`
						IsPaaS                              bool        `json:"isPaaS,omitempty"`
						EncryptsSensitiveObject             bool        `json:"encryptsSensitiveObject,omitempty"`
						ManagementType                      string      `json:"managementType,omitempty"`
						PendingDeletion                     bool        `json:"pendingDeletion,omitempty"`
						Rotation                            bool        `json:"rotation,omitempty"`
					} `json:"properties"`
				} `json:"accessibleResource"`
			} `json:"nodes"`
			PageInfo PageInfo `json:"pageInfo"`
		} `json:"entityEffectiveAccessEntries"`
	} `json:"data"`
}

type ResourcePermissions struct {
	Data struct {
		EntityEffectiveAccessEntries struct {
			Nodes []struct {
				GrantedEntity      *GrantedEntity `json:"grantedEntity"`
				AccessibleResource struct {
					Id         string `json:"id"`
					Name       string `json:"name"`
					Type       string `json:"type"`
					Properties struct {
						VertexID                            string      `json:"_vertexID"`
						AccessibleFromInternet              bool        `json:"accessibleFrom.internet"`
						CloudPlatform                       string      `json:"cloudPlatform"`
						CloudProviderURL                    string      `json:"cloudProviderURL"`
						CreationDate                        time.Time   `json:"creationDate"`
						Encrypted                           bool        `json:"encrypted"`
						EncryptionInTransit                 bool        `json:"encryptionInTransit"`
						ExternalId                          string      `json:"externalId"`
						FullResourceName                    interface{} `json:"fullResourceName"`
						HasSensitiveData                    bool        `json:"hasSensitiveData"`
						IsPublic                            bool        `json:"isPublic"`
						LoggingEnabled                      bool        `json:"loggingEnabled"`
						MaxExposureLevel                    int         `json:"maxExposureLevel"`
						Name                                string      `json:"name"`
						NativeType                          string      `json:"nativeType"`
						NumAddressesOpenForHTTP             int         `json:"numAddressesOpenForHTTP"`
						NumAddressesOpenForHTTPS            int64       `json:"numAddressesOpenForHTTPS"`
						NumAddressesOpenForNonStandardPorts int         `json:"numAddressesOpenForNonStandardPorts"`
						NumAddressesOpenForRDP              int         `json:"numAddressesOpenForRDP"`
						NumAddressesOpenForSSH              int         `json:"numAddressesOpenForSSH"`
						NumAddressesOpenForWINRM            int         `json:"numAddressesOpenForWINRM"`
						OpenToAllInternet                   bool        `json:"openToAllInternet"`
						ProviderUniqueId                    string      `json:"providerUniqueId"`
						PublicExposure                      string      `json:"publicExposure"`
						Region                              string      `json:"region"`
						RegionLocation                      string      `json:"regionLocation"`
						RegionType                          string      `json:"regionType"`
						ResourceGroupExternalId             interface{} `json:"resourceGroupExternalId"`
						RetentionPeriod                     int         `json:"retentionPeriod"`
						Status                              string      `json:"status"`
						SubscriptionExternalId              string      `json:"subscriptionExternalId"`
						UpdatedAt                           time.Time   `json:"updatedAt"`
						VersioningEnabled                   bool        `json:"versioningEnabled"`
						WebHostingEnabled                   bool        `json:"webHostingEnabled"`
						Zone                                interface{} `json:"zone"`
						Tags                                struct {
							Name string `json:"Name"`
						} `json:"tags,omitempty"`
						HasIAMAccessFromOutsideOrganization []string `json:"hasIAMAccessFromOutsideOrganization,omitempty"`
						NumAddressesOpenToInternet          int64    `json:"numAddressesOpenToInternet,omitempty"`
						NumPortsOpenToInternet              int      `json:"numPortsOpenToInternet,omitempty"`
						ValidatedOpenPorts                  int      `json:"validatedOpenPorts,omitempty"`
						WebHostingHosts                     string   `json:"webHostingHosts,omitempty"`
					} `json:"properties"`
					HasOriginalObject bool        `json:"hasOriginalObject"`
					UserMetadata      interface{} `json:"userMetadata"`
					IssueAnalytics    struct {
						HighSeverityCount     int `json:"highSeverityCount"`
						CriticalSeverityCount int `json:"criticalSeverityCount"`
					} `json:"issueAnalytics"`
				} `json:"accessibleResource"`
				ResourceCloudAccount struct {
					Id            string `json:"id"`
					Name          string `json:"name"`
					ExternalId    string `json:"externalId"`
					CloudProvider string `json:"cloudProvider"`
				} `json:"resourceCloudAccount"`
				GrantedEntityCloudAccount struct {
					Id            string `json:"id"`
					Name          string `json:"name"`
					ExternalId    string `json:"externalId"`
					CloudProvider string `json:"cloudProvider"`
				} `json:"grantedEntityCloudAccount"`
				Paths []struct {
					Path []struct {
						Entity struct {
							Id         string `json:"id"`
							Name       string `json:"name"`
							Type       string `json:"type"`
							Properties struct {
								VertexID                 string      `json:"_vertexID"`
								AwsUniqueIdentifier      string      `json:"awsUniqueIdentifier"`
								ClientId                 interface{} `json:"clientId"`
								CloudProviderURL         string      `json:"cloudProviderURL"`
								CreatedAt                time.Time   `json:"createdAt"`
								Description              interface{} `json:"description"`
								DisplayName              interface{} `json:"displayName"`
								Email                    interface{} `json:"email"`
								Enabled                  bool        `json:"enabled"`
								ExternalId               string      `json:"externalId"`
								FullResourceName         interface{} `json:"fullResourceName"`
								HasAccessToSensitiveData bool        `json:"hasAccessToSensitiveData"`
								HasAdminPrivileges       bool        `json:"hasAdminPrivileges"`
								HasHighPrivileges        bool        `json:"hasHighPrivileges"`
								InactiveInLast90Days     bool        `json:"inactiveInLast90Days"`
								Managed                  bool        `json:"managed"`
								Name                     string      `json:"name"`
								Namespace                interface{} `json:"namespace"`
								NativeType               string      `json:"nativeType"`
								ProviderUniqueId         string      `json:"providerUniqueId"`
								Region                   interface{} `json:"region"`
								SubscriptionExternalId   string      `json:"subscriptionExternalId"`
								UpdatedAt                time.Time   `json:"updatedAt"`
								UserDirectory            string      `json:"userDirectory"`
							} `json:"properties"`
							HasOriginalObject bool        `json:"hasOriginalObject"`
							UserMetadata      interface{} `json:"userMetadata"`
							IssueAnalytics    struct {
								HighSeverityCount     int `json:"highSeverityCount"`
								CriticalSeverityCount int `json:"criticalSeverityCount"`
							} `json:"issueAnalytics"`
						} `json:"entity"`
					} `json:"path"`
					AccessTypes       []string `json:"accessTypes"`
					Permissions       []string `json:"permissions"`
					PrincipalPolicies []struct {
						Policy struct {
							Name string `json:"name"`
							Type string `json:"type"`
						} `json:"policy"`
						GrantedEntity interface{} `json:"grantedEntity"`
					} `json:"principalPolicies"`
					ResourcePolicies []struct {
						Policy struct {
							Id         string `json:"id"`
							Name       string `json:"name"`
							Type       string `json:"type"`
							Properties struct {
								VertexID               string      `json:"_vertexID"`
								CloudPlatform          string      `json:"cloudPlatform"`
								ExternalId             string      `json:"externalId"`
								Name                   string      `json:"name"`
								Namespace              interface{} `json:"namespace"`
								NativeType             string      `json:"nativeType"`
								ProviderUniqueId       interface{} `json:"providerUniqueId"`
								Region                 interface{} `json:"region"`
								Statements             string      `json:"statements"`
								SubscriptionExternalId string      `json:"subscriptionExternalId"`
								UpdatedAt              time.Time   `json:"updatedAt"`
							} `json:"properties"`
							HasOriginalObject bool        `json:"hasOriginalObject"`
							UserMetadata      interface{} `json:"userMetadata"`
							IssueAnalytics    struct {
								HighSeverityCount     int `json:"highSeverityCount"`
								CriticalSeverityCount int `json:"criticalSeverityCount"`
							} `json:"issueAnalytics"`
						} `json:"policy"`
						GrantedEntity interface{} `json:"grantedEntity"`
					} `json:"resourcePolicies"`
				} `json:"paths"`
				AccessTypes []string `json:"accessTypes"`
				Permissions []string `json:"permissions"`
			} `json:"nodes"`
			PageInfo PageInfo `json:"pageInfo"`
		} `json:"entityEffectiveAccessEntries"`
	} `json:"data"`
}
