# 5. Oathkeeper Cookie Session Authentication

Date: 2026-02-01

## Status

Accepted

## Context

Django admin needs to authenticate users via Ory Kratos sessions. Previously, we used `django_ory_auth`
which made direct API calls to Kratos. This approach has drawbacks:

- Each request requires an additional network call to Kratos
- Django needs to know Kratos SDK URL
- No centralized authentication policy

## Decision

Use Oathkeeper as authentication proxy with cookie_session authenticator:

1. Oathkeeper validates `ory_kratos_session` cookie via Kratos `/sessions/whoami`
2. Oathkeeper injects `X-User-ID` header with identity.id
3. Django trusts this header via `RemoteUserMiddleware`

## Oathkeeper Access Rule

Add to `/etc/rules/access-rules.json` in Oathkeeper ConfigMap (namespace `auth`):

```json
{
  "id": "django-admin",
  "match": {
    "url": "http://admin.shortlink-shop.svc.cluster.local/<**>",
    "methods": ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"]
  },
  "authenticators": [{ "handler": "cookie_session" }],
  "authorizer": { "handler": "allow" },
  "mutators": [{ "handler": "header" }]
}
```

The global `cookie_session` authenticator and `header` mutator are already configured in Oathkeeper:

```yaml
authenticators:
  cookie_session:
    config:
      check_session_url: http://kratos-public.auth.svc.cluster.local/sessions/whoami
      subject_from: identity.id
      extra_from: '@this'
    enabled: true

mutators:
  header:
    config:
      headers:
        X-User-ID: '{{ print .Subject }}'
    enabled: true
```

## Security

- Istio AuthorizationPolicy restricts access to Django only from namespace `auth` (Oathkeeper)
- This prevents clients from bypassing Oathkeeper and spoofing `X-User-ID` header

## Consequences

- Centralized authentication at Oathkeeper level
- Django configuration is simpler (no Kratos SDK dependency for auth)
- Session validation is cached at Oathkeeper level
- Must maintain AuthorizationPolicy to prevent header spoofing
