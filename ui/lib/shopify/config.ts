export const domain =
  (typeof process.env.NEXT_PUBLIC_API_URI !== 'undefined'
    ? process.env.NEXT_PUBLIC_API_URI
    : process.env.API_URI) ?? '';

export const getGraphqlEndpoint = (): string =>
  domain ? `${domain}/api/graphql` : '/api/graphql';
