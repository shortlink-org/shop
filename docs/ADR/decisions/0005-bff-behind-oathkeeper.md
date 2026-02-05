# 5. BFF Behind Oathkeeper for JWT/Session

Date: 2026-02-04

## Status

Accepted

## Context

GraphQL BFF must receive authenticated user context (e.g. JWT claims or Kratos session) to pass to subgraphs. If the Gateway sends traffic directly to BFF, BFF never gets validated tokens or headers like `X-User-ID`. Authentication must happen before BFF.

## Decision

Route `/api` through Oathkeeper first (same pattern as admin and temporal):

1. **Gateway**: Path `/api` with URL rewrite `/api` → `/`, backendRef → `oathkeeper-proxy` (namespace `auth`, port 4455).
2. **Oathkeeper**: Validates cookie session or JWT, injects headers (e.g. `X-User-ID`), forwards to BFF.
3. **BFF**: Receives request with headers and forwards them to subgraphs (e.g. admin-graphql).

Flow: Client → Istio Gateway → Oathkeeper → BFF → subgraphs.

## Oathkeeper Access Rule

Add to Oathkeeper ConfigMap in namespace `auth` (e.g. `/etc/rules/access-rules.json`), same style as `django-admin` (see admin/docs/ADR/decisions/0005-oathkeeper-auth.md):

```json
{
  "id": "shop-bff",
  "match": {
    "url": "http://shortlink-shop-bff.shortlink-shop.svc.cluster.local/<**>",
    "methods": ["GET", "POST", "OPTIONS"]
  },
  "authenticators": [{ "handler": "cookie_session" }],
  "authorizer": { "handler": "allow" },
  "mutators": [{ "handler": "header" }]
}
```

Gateway rewrites `/api/*` to `/*`, so Oathkeeper receives paths like `/graphql` and forwards to `http://shortlink-shop-bff.shortlink-shop:9991/graphql`. Use the same `cookie_session` authenticator and `header` mutator as admin so that `X-User-ID` is set for downstream services.

If anonymous or JWT-only access is needed for some paths, add separate rules in auth (e.g. different authenticator or `anonymous`).

## Consequences

- BFF gets validated identity headers from Oathkeeper.
- Subgraphs can rely on forwarded `Authorization` / `X-User-ID` from BFF.
- Oathkeeper access rule for BFF must be maintained in the auth repo/namespace.
