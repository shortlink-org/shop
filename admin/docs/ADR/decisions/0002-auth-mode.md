# 2. Authentication: Oathkeeper + RemoteUser

Date: 2024-03-24  
Updated: 2026-02-04

## Status

Accepted

## Context

We need a robust and secure authentication system for the Django admin. Identity is provided by Ory Kratos.

**Previous approach (superseded):** We used `django_ory_auth`, a Django middleware that talked directly to Kratos. That had drawbacks:

- Each request triggered an extra network call to Kratos
- Django had to know the Kratos SDK URL
- No centralized authentication policy at the edge

## Decision

Use **Oathkeeper** as an authentication proxy in front of Django:

1. **Oathkeeper** validates the `ory_kratos_session` cookie via Kratos `/sessions/whoami` (cookie_session authenticator).
2. Oathkeeper injects the **X-User-ID** header with `identity.id`.
3. **Django** trusts this header via `RemoteUserMiddleware` and does not call Kratos for auth.

Traffic flow: Client → Istio Gateway → Oathkeeper → Admin (Django). Requests to Admin that come via **admin-graphql** (BFF subgraph) are allowed by Istio AuthorizationPolicy; admin-graphql forwards **X-User-ID**, **Authorization**, and **Cookie** to Django so identity is preserved.

### Oathkeeper access rule

Add to Oathkeeper ConfigMap in namespace `auth` (e.g. `/etc/rules/access-rules.json`):

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

Global `cookie_session` authenticator and `header` mutator (example):

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

### Security

- **Istio AuthorizationPolicy** allows access to Django only from:
  - Requests with **X-User-ID** header (set by Oathkeeper), or
  - The **admin-graphql** service account (server-to-server calls from BFF subgraph).
- This prevents clients from bypassing Oathkeeper and spoofing `X-User-ID`.

### Admin roles (Postgres)

To grant admin rights to a user, set in Postgres:

- `is_staff` = `true`
- `is_superuser` = `true`

## Consequences

- Authentication is centralized at Oathkeeper; Django does not depend on Kratos SDK for auth.
- Session validation is done (and can be cached) at Oathkeeper.
- Django configuration is simpler (RemoteUserMiddleware + trust of X-User-ID).
- The AuthorizationPolicy must be kept to prevent header spoofing.
- Admin-graphql and BFF must forward X-User-ID (and optionally Authorization, Cookie) to Django when calling the Admin API from the GraphQL layer.
