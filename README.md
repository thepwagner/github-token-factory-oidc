# GitHub Token Factory OIDC

This is a server that issues short-lived GitHub installation tokens in response to clients authenticated by OpenID Connect identity tokens:
* GitHub Actions can authenticate to access multiple repositories, or organization-level resources.
* Google Accounts can run GitHub scripts.
* Tekton jobs can authenticate to GitHub from a kubernetes cluster.

Should you store a Personal Access Token as a CI Secret? No, GTFO.

Best used with [thepwagner/token-action](https://github.com/thepwagner/token-action)!

## Setup

Users must [create a GitHub app](https://docs.github.com/en/developers/apps/building-github-apps/creating-a-github-app).
Select any permissions you'd like to be able to grant, be sure to include the minimum permissions:
  * `contents:read`

A single issued token may not span repositories belonging to multiple users/organizations.
The server may issue tokens to multiple users/organizations using a shared GitHub App, or multiple GitHub Apps. GitHub requires that apps installed to multiple users/organizations are public.

## Security Model

The server holds secrets for all configured GitHub applications. It is what issues GitHub tokens to clients, so owning the server means owning the organizations/users its apps are installed to. Don't let that happen.

GitHub tokens issued by the server are valid for 1 hour. This is a hardcoded GitHub API limitation. Your token may be valid for less than that if the server returns a cached token.

Since policy decisions may be expensive, the server defines a global list of valid OIDC issuers. Tokens presented by other issuers are rejected.

### Permissions Enforcement

All permissions requests are denied by default.

Permissions must be granted to the target repository. Repositories can add a [Rego policy file](https://www.openpolicyagent.org/docs/latest/policy-language/) that grants permissions at `.github/tokens.rego`.

```rego
default allow = false

# Private Actions can do everything except write
allow {
	input.claims.iss == "https://token.actions.githubusercontent.com"
	input.claims.repository_owner == "thepwagner"
	input.claims.repository_visibility != "public"
	input.permissions.contents != "write"
}
```

Repository owners may also host a policy for multiple repositories, by adding a `.github/tokens.rego` file to a repository called `.github`. (e.g. `https://github.com/thepwagner/.github`).

When combined with branch protection and `CODEOWNERS`, hosting within repositories allows teams to decentralize and GitOps their collaborations.

Real talk: the primary reason for hosting policies within repositories is to allow the server to be hosted serverless.
