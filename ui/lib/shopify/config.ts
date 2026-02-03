export const domain = process.env.API_URI ?? '';

export const getGraphqlEndpoint = (): string => `${domain}/graphql`;
