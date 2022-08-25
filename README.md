# github-token-action-server

Every repository has a file: `.github/acl.yaml`

This defines identities that may request permissions against the repository:

```yaml
- workflow:
    oidc_field: filter_value
    oidc_field2: /filter_regex/
  permissions:
    key: value
- identity:
    issuer: issue_mc_issuer
    oidc_field: filter_value
    oidc_field2: /filter_regex/
  permissions:
    key: value
```

When a request is made:
1. The identity of the token is identified.
2. Every repository's `.github/acl.yaml` is read. Organization permissions and organization-wide permissions are read from the `.github` repository.
3. 