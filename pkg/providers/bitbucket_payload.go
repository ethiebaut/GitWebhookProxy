package providers

// PushPayload contains the information for Bitbucket's push hook event
type BitbucketPushPayload struct {
	Actor       struct {
		DisplayName        string   `json:"display_name"`
	} `json:"actor"`
}
