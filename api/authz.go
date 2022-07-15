package api

// WorkflowID identifies a workflow execution.
// JSON keys should align with the claims of the Actions OIDC token
// https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect#understanding-the-oidc-token
type WorkflowID struct {
	Environment       string `json:"environment"`
	Ref               string `json:"ref"`
	SHA               string `json:"sha"`
	Repository        string `json:"repository"`
	RepositoryID      string `json:"repository_id"`
	RepositoryOwner   string `json:"repository_owner"`
	RepositoryOwnerID string `json:"repository_owner_id"`
	ActorID           string `json:"actor_id"`
	RunID             string `json:"run_id"`
	RunNumber         string `json:"run_number"`
	RunAttempt        string `json:"run_attempt"`
	Actor             string `json:"actor"`
	WorkflowName      string `json:"workflow"`
	HeadRef           string `json:"head_ref"`
	BaseRef           string `json:"base_ref"`
	EventName         string `json:"event_name"`
	RefType           string `json:"ref_type"`
	JobWorkflowRef    string `json:"job_workflow_ref"`
}

// TokenRequest is a request from a workflow for permissions
type TokenRequest struct {
	Repositories []string          `json:"repositories"`
	Permissions  map[string]string `json:"permissions"`
}

// TokenCheck checks if a workflow is authorized to request a token
type TokenCheck func(*WorkflowID, *TokenRequest) (bool, error)

func TokenCheckYOLO(*WorkflowID, *TokenRequest) (bool, error) {
	return true, nil
}
