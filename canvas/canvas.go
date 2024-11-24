package canvas

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// CanvasClient is a client for interacting with the Canvas API.
type CanvasClient struct {
	baseUrl     string
	accessToken string
	pageSize    int
	httpClient  *http.Client
	WebUrl      string
}

// authTransport is a custom RoundTripper that adds the Authorization header to all requests.
type authTransport struct {
	Transport   http.RoundTripper
	AccessToken string
}

// RoundTrip adds the Authorization header to every request.
func (a *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clonedReq := req.Clone(req.Context())
	clonedReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.AccessToken))
	return a.Transport.RoundTrip(clonedReq)
}

func NewCanvasClient(baseUrl, accessToken string, pageSize int) (*CanvasClient, error) {
	if baseUrl == "" {
		return nil, fmt.Errorf("invalid base url")
	}

	if accessToken == "" {
		return nil, fmt.Errorf("invalid access token")
	}

	if pageSize <= 0 {
		return nil, fmt.Errorf("invalid page size")
	}

	httpClient := &http.Client{
		Timeout: time.Second * 10,
		Transport: &authTransport{
			Transport:   http.DefaultTransport,
			AccessToken: accessToken,
		},
	}

	canvasClient := &CanvasClient{
		baseUrl:     baseUrl,
		accessToken: accessToken,
		pageSize:    pageSize,
		httpClient:  httpClient,
		WebUrl:      getWebUrl(baseUrl),
	}

	return canvasClient, nil
}

func getWebUrl(baseUrl string) string {
	index := strings.Index(baseUrl, ".com")

	if index != -1 {
		return baseUrl[:index+4]
	}

	return ""
}

// getNextUrl extracts the next url from the Link header string.
// It returns empty string if there is no next url.
//
// Canvas API provides pagination information in the Link header as comma separated string:
// Link:
// <https://<canvas>/api/v1/courses/:id/discussion_topics.json?opaqueA>; rel="current",
// <https://<canvas>/api/v1/courses/:id/discussion_topics.json?opaqueB>; rel="next",
// <https://<canvas>/api/v1/courses/:id/discussion_topics.json?opaqueC>; rel="first",
// <https://<canvas>/api/v1/courses/:id/discussion_topics.json?opaqueD>; rel="last"
func getNextUrl(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}

	links := strings.Split(linkHeader, ",")
	nextRegEx := regexp.MustCompile(`^<(.*)>; rel="next"$`)

	for _, link := range links {
		if nextRegEx.MatchString(strings.TrimSpace(link)) {

			start := strings.Index(link, "<")
			end := strings.Index(link, ">")

			if start != -1 && end != -1 && end > start {
				return link[start+1 : end]
			}
		}
	}

	return ""
}
