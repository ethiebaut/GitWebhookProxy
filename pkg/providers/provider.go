package providers

import (
	"errors"
	"net/http"
	"strings"
)

const (
	GithubProviderKind            = "github"
	GitlabProviderKind            = "gitlab"
	BitbucketProviderKind         = "bitbucket"
	ContentTypeHeader             = "Content-Type"
	DefaultContentTypeHeaderValue = "application/json"
)

// Event defines a provider hook event type
type Event string

type Provider interface {
	GetHeaderKeys() []string
	Validate(hook Hook) bool
	GetCommitter(hook Hook) string
	GetProviderName() string
}

func assertProviderImplementations() {
	var _ Provider = (*GithubProvider)(nil)
	var _ Provider = (*GitlabProvider)(nil)
	var _ Provider = (*BitbucketProviderKind)(nil)
}

func NewProvider(provider string, secret string) (Provider, error) {
	if len(provider) == 0 {
		return nil, errors.New("Empty provider string specified")
	}

	switch strings.ToLower(provider) {
	case GithubProviderKind:
		return NewGithubProvider(secret)
	case GitlabProviderKind:
		return NewGitlabProvider(secret)
	case BitbucketProviderKind:
		return NewBitbucketProvider(secret)
	default:
		return nil, errors.New("Unknown Git Provider '" + provider + "' specified")
	}
}

type Hook struct {
	Payload       []byte
	Headers       map[string]string
	RequestMethod string
	Request       *http.Request
}
