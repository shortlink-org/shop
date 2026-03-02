// Server: use API_URI (internal cluster URL) so fetch never goes through Oathkeeper.
// Client: use NEXT_PUBLIC_API_URI or '' for same-origin /api/graphql.
const domain =
  typeof window === 'undefined'
    ? (process.env.API_URI ?? '')
    : (process.env.NEXT_PUBLIC_API_URI ?? '');

export { domain };

export const getGraphqlEndpoint = (): string =>
  domain ? `${domain}/api/graphql` : '/api/graphql';
