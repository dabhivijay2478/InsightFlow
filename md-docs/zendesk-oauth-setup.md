# Zendesk Support OAuth setup

MantrixFlow reads Zendesk Support ticketing data through the Zendesk Support
API. Authentication is OAuth-only. The connector does not accept an email,
password, API token, or manually pasted OAuth token.

Zendesk's Conversations API is not a replacement for the Support API used by
this connector. Conversations API covers live chat conversations and agent or
customer messaging. MantrixFlow currently extracts Support resources such as
tickets, ticket events, users, organizations, groups, forms, SLAs, and
satisfaction ratings.

## Zendesk client setup

1. During development, create a **confidential** local OAuth client in the
   Zendesk Admin Center under **Apps and integrations → APIs → OAuth clients**.
   Public clients require PKCE and are not compatible with MantrixFlow's
   server-side confidential authorization-code flow.
2. Register this exact redirect URL:

   ```text
   https://<public-api-host>/api/v1/zendesk/oauth/callback
   ```

3. Configure the OAuth client to allow the `read` scope. Zendesk documents
   `read` as access to GET endpoints, which matches this source-only connector.
4. Configure the Go API:

   ```dotenv
   ZENDESK_OAUTH_CLIENT_ID=zdg-mantrixflow
   ZENDESK_OAUTH_CLIENT_SECRET=<secret>
   ZENDESK_OAUTH_REDIRECT_BASE_URL=https://<public-api-host>
   ```

   `ZENDESK_OAUTH_REDIRECT_BASE_URL` may be omitted when `API_PUBLIC_URL`
   already contains the correct public API origin.

   Replace every example placeholder with a real value. The client identifier
   must exactly match Zendesk's **Identifier** field, and the resulting callback
   URL must exactly match one of the client's **Redirect URLs**.

5. For a production integration used by multiple Zendesk customers, request a
   global OAuth client from Zendesk. Zendesk requires global OAuth for this
   distribution model. The local client's identifier should use the `zdg-`
   prefix before it is submitted for conversion.

Official references:

- [Migrating from API tokens to OAuth access tokens](https://developer.zendesk.com/documentation/authentication/oauth-migration/)
- [Set up a global OAuth client](https://developer.zendesk.com/documentation/marketplace/building-a-marketplace-app/set-up-a-global-oauth-client/)
- [Zendesk OAuth scopes](https://developer.zendesk.com/api-reference/ticketing/oauth/oauth_tokens/#scopes)
- [Zendesk Support API introduction](https://developer.zendesk.com/api-reference/ticketing/introduction/)
- [Conversations API introduction](https://developer.zendesk.com/api-reference/live-chat/chat-conversations-api/conversations-api/)

## Runtime flow

1. The browser submits only the connection name and Zendesk subdomain to the
   organization-scoped Go endpoint.
2. Go creates a short-lived, single-use, hashed OAuth state and redirects the
   browser to the tenant's Zendesk authorization page.
3. Zendesk returns an authorization code to the public callback.
4. Go exchanges it server-side using the confidential client secret, verifies
   the granted Bearer token against the Support API, and encrypts both access
   and refresh tokens before storing them.
5. Before discovery, preview, or a pipeline run, Go refreshes an access token
   that is close to expiry and persists both rotated tokens. Zendesk invalidates
   the old token pair after a successful refresh.
6. The ELT server receives only the current Bearer access token and calls
   read-only Support endpoints.

Access tokens request a 30-minute lifetime. Refresh tokens request the maximum
documented 90-day lifetime. If authorization or refresh expires, reconnect the
Zendesk connection from the connection screen.

## Existing API-token connections

Legacy Zendesk API-token configurations are intentionally not supported by the
new runtime. They are not silently upgraded, because no server can exchange an
API token for an OAuth grant. Open each existing Zendesk connection and use
**Reconnect Zendesk**. After all workflows are reconnected, revoke the old
API tokens in Zendesk Admin Center.

## Verification

- The Zendesk card is available only when the Go server has both OAuth client
  environment variables.
- The form contains connection name and subdomain only.
- The Zendesk approval screen requests read access.
- The callback creates or updates the connection and returns to its detail
  page.
- Discovery and preview succeed without an email or API token.
- Stored connection JSON exposes neither access nor refresh token in API or
  agent-facing responses.

## Authorization-page troubleshooting

If Zendesk shows **Your request experienced a server error**, inspect the
authorization request and OAuth client before checking Zendesk status:

- `client_id` must exactly equal the OAuth client's **Identifier**. A local
  identifier such as `mantrixflow` is different from `zdg-mantrixflow`.
- `redirect_uri` must exactly equal a registered **Redirect URL**, including
  scheme, hostname, port, path, and trailing-slash behavior.
- The OAuth client must be **Confidential** for this integration.
- `response_type=code` and `scope=read` must be present.

Restart the Go API after changing any `ZENDESK_OAUTH_*` environment variable.
MantrixFlow rejects unresolved callback placeholders and non-HTTPS remote
callback origins before redirecting to Zendesk.
