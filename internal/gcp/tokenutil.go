package gcp

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/api/iamcredentials/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/sts/v1"
)

const (
	audienceFormat = "//iam.googleapis.com/projects/%s/locations/global/workloadIdentityPools/%s/providers/%s"
	scopeURL       = "https://www.googleapis.com/auth/cloud-platform"
)

func GetFederalToken(idToken, projectNumber, poolId, providerId string) (string, error) {
	ctx := context.Background()
	stsService, err := sts.NewService(ctx, option.WithoutAuthentication())
	if err != nil {
		return "", err
	}

	audience := fmt.Sprintf(audienceFormat, projectNumber, poolId, providerId)

	tokenRequest := &sts.GoogleIdentityStsV1ExchangeTokenRequest{
		GrantType:          "urn:ietf:params:oauth:grant-type:token-exchange",
		SubjectToken:       idToken,
		Audience:           audience,
		Scope:              scopeURL,
		RequestedTokenType: "urn:ietf:params:oauth:token-type:access_token",
		SubjectTokenType:   "urn:ietf:params:oauth:token-type:id_token",
	}

	tokenResponse, err := stsService.V1.Token(tokenRequest).Do()
	if err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

func GetGoogleCloudAccessToken(federatedToken string, serviceAccountEmail string) (string, error) {
	ctx := context.Background()
	token := &oauth2.Token{AccessToken: federatedToken}
	service, err := iamcredentials.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(token)))
	if err != nil {
		return "", err
	}

	name := "projects/-/serviceAccounts/" + serviceAccountEmail
	// rb (request body) specifies parameters for generating an access token.
	rb := &iamcredentials.GenerateAccessTokenRequest{
		Scope: []string{scopeURL},
	}
	// Generate an access token for the service account using the specified parameters
	resp, err := service.Projects.ServiceAccounts.GenerateAccessToken(name, rb).Do()
	if err != nil {
		return "", err
	}

	return resp.AccessToken, nil
}
