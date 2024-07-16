# GitHub Token Factory OIDC

## Deprecated

I got tired of maintaining this, you should use https://github.com/octo-sts/app .

I _still_ have to use PATs, because GitHub Apps can't request organization-level package permissions.

## Old README

This is a server that issues short-lived GitHub installation tokens in response to clients authenticated by OpenID Connect identity tokens:

- GitHub Actions can use this to authenticate to access multiple private repositories, or to use organization-level resources like adding issues to project boards.
- Google Accounts can use this to script uploading GitHub releases.
- Tekton jobs can use this to authenticate to GitHub from a kubernetes cluster.

Should you store a Personal Access Token as a CI Secret? No, try GTFO.

Best used with [thepwagner/token-action](https://github.com/thepwagner/token-action)!

### Limitations

* Issued tokens are valid for 1 hour, and can not be shorter or longer. This is a GitHub limitation.
* A single issued token may not have permissions to multiple GitHub users/organizations. This is a GitHub limitation.
* The server may issue tokens to multiple users/organizations using a public GitHub App with multiple installations, or multiple private GitHub Apps.

### Setup

Users must [create a GitHub app](https://docs.github.com/en/developers/apps/building-github-apps/creating-a-github-app).
Select any permissions you'd like to be able to grant, be sure to include the minimum permissions:

- `contents:read`

If you intend to support multiple users/organizations from a single app, GitHub requires that apps installed to multiple users/organizations are public.
Unless you are really good at writing policies, you should probably not do this - set up a private app for each user/organization.

### Security Model

The server holds secrets for all configured GitHub applications. It is what issues GitHub tokens to clients, so owning the server means owning the organizations/users its apps are installed to. Don't let that happen.

Since deciding if a token should be issued can be expensive, the server defines a global list of valid OIDC issuers. Tokens presented by other issuers are rejected.

#### Policy Files

All request for tokens are denied by default. Permissions must be granted by the target repository. Repositories can add a [Rego policy file](https://www.openpolicyagent.org/docs/latest/policy-language/) that grants permissions at `.github/tokens.rego`:

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

Repository owners may also host a policy for all of their repositories, by adding a `.github/tokens.rego` file to a repository called `.github`. For a token to be issued, it must be allowed by one of:

- The repository owner's policy, hosted at `.github/tokens.rego` in `${user}/.github` (e.g. `thepwagner/.github`)
- EVERY requested repository's policy, hosted at `.github/tokens.rego` in `${user}/${repo}` in each repository. (e.g. `thepwagner/foo`, `thepwagner/bar`, ...)

Individual repository policies are intended to avoid organizations bottlenecking in the `.github` policy monorepo: collaborations between projects can be setup peer-to-peer.
Certain permissions, those that affect the user/organization and not just repositories, can only be granted by the owner-level policy from the `.github` repository.

Storing policies in the repository means `contents:write` can be escalated to other permissions, by pushing new policies.
Policies are always fetched from a repository's default branch, so you can enable branch protection to discourage this. Depending on your organization, consider a `CODEOWNERS` in the `.github` repository.

Real talk: the primary reason for hosting policies within repositories is to allow the server to be hosted serverless. It is a trade-off, not a design principle.


## Prior Work

* In 2018, I was on the team that built the first version of GitHub Actions. To enable the most use cases and drive adoption, we gave every workflow most permissions to the repository it was hosted in. Sorry about that.
* In 2020, at a GitHub hack day, I made a thing called [Secret Garden](https://github.com/thepwagner/secret-garden) to enable GitHub Actions to run with reduced permissions, to access multiple repositories, and to integrate with resources like projects. The technique was "token juggling" - periodically regenerating scoped tokens so they are available when needed.
* In April 2021, GitHub [added support for reducing permissions](https://github.blog/changelog/2021-04-20-github-actions-control-permissions-for-github_token/). Still no multi-repo or projects support.
* In July 2022, my team at Shopify wanted to add an issue to a project board. I wondered if the recently added [OIDC token support](https://github.blog/changelog/2021-10-27-github-actions-secure-cloud-deployments-with-openid-connect/) could be leveraged to do "token JIT-ing" and determined [it can](https://twitter.com/meofthecloud/status/1547959659222315010).
* In August 2022, someone on Twitter [said they wanted to clone GitHub repos based on OIDC tokens](https://twitter.com/mattomata/status/1561345258545446913), so I thought this idea was worth finishing enough to share. Slap in a little Rego and ship it.

This is where the project's biases towards serverless hosting and storing policies as repository content come from.
