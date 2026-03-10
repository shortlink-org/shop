# 5. Shop authentication via Oathkeeper (JWT, identity propagation)

Date: 2026-02-04

## Status

Accepted

## Context

GraphQL BFF must receive authenticated user context (e.g. JWT claims or Kratos session) to pass to subgraphs. If traffic reached the UI or BFF without Oathkeeper, there would be no validated identity (no JWT, no `X-User-ID`). Authentication must happen before the app.

## Decision

Route shop traffic through Oathkeeper first (same pattern as admin and temporal):

1. **Routing**: Traffic to `https://shop.shortlink.best/` is routed to **Oathkeeper** by an **HTTPRoute** (defined in the shop UI chart, `ui/ops/Helm/values.yaml`): hostname `shop.shortlink.best`, path `/`, `backendRef` → `oathkeeper-proxy` (namespace `auth`, port 4455). The Gateway itself (e.g. external-gateway in istio-ingress) is not defined in the shop repo.
2. **Oathkeeper**: Validates cookie session (Kratos whoami), applies mutators (header + id_token), forwards to **UI** (`shortlink-shop-ui:3000`).
3. **UI**: Receives request with `X-User-ID` and `Authorization: Bearer <jwt>`. For GraphQL, the API route proxies to **BFF** and forwards only `Authorization` (and tracing).
4. **BFF**: Receives request with JWT, forwards only `Authorization` to subgraphs (e.g. oms-graphql). Subgraphs get `x-user-id` from Istio (JWT claim `sub`).

Flow: Client → Oathkeeper → **UI** → BFF → subgraphs. (Routing to Oathkeeper is via HTTPRoute; we do not maintain a Gateway in the shop repo.)

---

## How authorization works (reference)

To avoid forgetting how auth is wired end-to-end:

1. **Oathkeeper** (auth repo, namespace `auth`):
   - Validates the request with **cookie_session** (Kratos whoami).
   - Uses mutators **header** and **id_token**:
     - **header**: sets `X-User-ID` and `X-Email` from the session (for backward compatibility / UI).
     - **id_token**: issues a signed JWT and sets `Authorization: Bearer <jwt>`. The JWT contains claim **`sub`** = user id (from `identity.id`).
   - Upstream (UI for shop) receives both headers and the JWT.

2. **Between services we propagate only the `Authorization` header** (no `X-User-ID`):
   - **UI → BFF**: the API route (`/api/graphql`) forwards only `Authorization` (and tracing headers) to the BFF. It does **not** forward `X-User-ID`.
   - **BFF → subgraphs**: in `bff/router.yaml` we propagate only `Authorization`, `Cookie`, `traceparent`, `trace-id`. We do **not** propagate `X-User-ID`.

3. **Subgraphs** (e.g. oms-graphql, admin-graphql):
   - Receive the request with `Authorization: Bearer <jwt>`.
   - **Istio** (RequestAuthentication) validates the JWT using Oathkeeper's JWKS (`http://oathkeeper-api.auth.svc.cluster.local:4456/.well-known/jwks.json`).
   - **Istio** (outputClaimToHeaders) sets the header **`x-user-id`** from the JWT claim **`sub`** before the request reaches the application.
   - The application code reads `x-user-id` from the request headers; it does not parse the JWT.

4. **Where it is configured**:
   - **Auth repo** (`auth/ops/Helm/oathkeeper`): Oathkeeper config, access rules (e.g. `shop:bff` with mutators `header` + `id_token`), JWKS for signing.
   - **Shop BFF** (`bff/router.yaml`): `headers.all.request` — only `Authorization` (and Cookie, traceparent, trace-id).
   - **Shop UI** (`ui/app/api/graphql/route.ts`): forwards only `authorization` (and tracing) to the BFF.
   - **Shop subgraphs** (e.g. `oms-graphql/ops/Helm/oms-graphql`): Helm template `request-authentication.yaml` with `outputClaimToHeaders: [{ header: "x-user-id", claim: "sub" }]`, and `requestAuthentication.issuer` / `jwksUri` in values pointing at Oathkeeper.

**Summary**: User identity is carried in the JWT (`sub`). Only `Authorization` is propagated between services. Each service that needs `x-user-id` gets it from Istio, which derives it from the JWT.

### Sequence diagram

```mermaid
sequenceDiagram
    participant Client
    participant Route as HTTPRoute
    participant Oathkeeper
    participant UI as UI (Next.js)
    participant BFF as BFF (Cosmo Router)
    participant Istio as Istio (subgraph sidecar)
    participant Subgraph as Subgraph (e.g. oms-graphql)

    Client->>Route: GET/POST shop.shortlink.best + Cookie
    Route->>Oathkeeper: backendRef oathkeeper-proxy:4455
    Oathkeeper->>Oathkeeper: cookie_session → Kratos whoami
    Oathkeeper->>Oathkeeper: header mutator → X-User-ID, X-Email
    Oathkeeper->>Oathkeeper: id_token mutator → JWT (sub = user id)
    Oathkeeper->>UI: request + X-User-ID + Authorization: Bearer &lt;jwt&gt;

    UI->>BFF: POST /graphql + Authorization only (+ trace)
    Note over UI,BFF: X-User-ID not forwarded

    BFF->>Istio: POST /Shop/QueryGetCart + Authorization only
    Note over BFF,Istio: X-User-ID not forwarded
    Istio->>Istio: RequestAuthentication → validate JWT (JWKS)
    Istio->>Istio: outputClaimToHeaders → sub → x-user-id
    Istio->>Subgraph: request + x-user-id header
    Subgraph->>Subgraph: read x-user-id from headers
```

## Oathkeeper Access Rule

The live rules are in the **auth repo** (`auth/ops/Helm/oathkeeper/values.yaml`, `accessRules`). The shop rule (`shop:bff`) matches `https://shop.shortlink.best/<**>`, upstream is the UI (`shortlink-shop-ui`). It uses **cookie_session** and mutators **header** + **id_token**, so the UI receives both `X-User-ID` and `Authorization: Bearer <jwt>` (JWT has claim `sub` = user id).

If you add a new rule (e.g. for another path), use the same pattern: `cookie_session` and mutators `header` + `id_token` so that downstream receives the JWT. Between services we only propagate `Authorization`; backends get `x-user-id` from Istio.

If anonymous or JWT-only access is needed for some paths, add separate rules in auth (e.g. different authenticator or `anonymous`).

### id_token mutator (already in use)

In the auth repo, Oathkeeper's global config enables the **id_token** mutator (`issuer_url`, `jwks_url` pointing at the signing key). The public key is served at `http://oathkeeper-api.auth.svc.cluster.local:4456/.well-known/jwks.json`; Istio RequestAuthentication on subgraphs uses this JWKS. The JWT's **`sub`** claim is set from the session subject (user id). Subgraphs use Helm template `request-authentication.yaml` with `outputClaimToHeaders: [{ header: "x-user-id", claim: "sub" }]`.

## Consequences

- BFF receives the JWT from the UI (which got it from Oathkeeper) and optionally legacy headers like `X-User-ID` on the UI.
- Between services only **Authorization** is propagated; subgraphs get **x-user-id** from Istio (JWT claim `sub` → header).
- Oathkeeper access rules and id_token mutator are maintained in the auth repo/namespace.
- Adding a new backend that needs user identity: ensure it has Istio RequestAuthentication with `outputClaimToHeaders` (sub → x-user-id) and that the BFF propagates `Authorization` to it.
