package providers

import (
	"encoding/json"
	"log"
	"strings"
)

// Header constants
const (
	XBitbucketEventKey = "X-Event-Key"
	BitbucketName   = "bitbucket"
)

type BitbucketProvider struct {
	secret string
}

func NewBitbucketProvider(secret string) (*BitbucketProvider, error) {
	return &BitbucketProvider{
		secret: secret,
	}, nil
}

func (p *BitbucketProvider) GetProviderName() string {
	return BitbucketName
}

// Not adding XBitbucketToken will make token validation optional
func (p *BitbucketProvider) GetHeaderKeys() []string {
	if len(strings.TrimSpace(p.secret)) > 0 {
		return []string{
			XBitbucketEventKey,
			ContentTypeHeader,
		}
	}

	return []string{
		XBitbucketEventKey,
		ContentTypeHeader,
	}
}

// Bitbucket token validation:
func (p *BitbucketProvider) Validate(hook Hook) bool {
	tokens, ok := hook.Request.URL.Query()["secret"]
	// Validation fails if secret is configured but did not receive from Bitbucket
	if !ok || len(tokens[0]) < 1 {
		return false
	}
	token := tokens[0]

	return strings.TrimSpace(token) == strings.TrimSpace(p.secret)
}

func (p *BitbucketProvider) GetCommitter(hook Hook) string {
	var payloadData BitbucketPushPayload
	if err := json.Unmarshal(hook.Payload, &payloadData); err != nil {
		log.Printf("Bitbucket hook payload unmarshalling failed")
		return ""
	}

	return payloadData.Actor.DisplayName
}
