import { ApolloClient, HttpLink, InMemoryCache } from '@apollo/client';
import { SetContextLink } from '@apollo/client/link/context';
import { RetryLink } from '@apollo/client/link/retry';

import { HTTP_STATUS_RATE_LIMIT, RATE_LIMIT_MESSAGE } from './constants';

const RATE_LIMIT_RETRY_DELAY_MS = 5000;

const graphqlUri =
  process.env.NEXT_PUBLIC_GRAPHQL_URL || 'http://localhost:9991/graphql';

function createCustomFetch() {
  return async (input: RequestInfo | URL, init?: RequestInit) => {
    const response = await fetch(input, init);
    if (response.status === HTTP_STATUS_RATE_LIMIT) {
      const err = new Error(RATE_LIMIT_MESSAGE) as Error & {
        statusCode: number;
      };
      err.statusCode = HTTP_STATUS_RATE_LIMIT;
      throw err;
    }
    return response;
  };
}

const httpLink = new HttpLink({
  uri: graphqlUri,
  fetch: createCustomFetch(),
});

const retryLink = new RetryLink({
  delay: (attempt, _operation, error) =>
    (error as { statusCode?: number })?.statusCode === HTTP_STATUS_RATE_LIMIT
      ? RATE_LIMIT_RETRY_DELAY_MS
      : 300,
  attempts: {
    max: 2,
    retryIf: (error) =>
      (error as { statusCode?: number })?.statusCode === HTTP_STATUS_RATE_LIMIT,
  },
});

// Add auth token to requests
const authLink = new SetContextLink((prevContext, _operation) => {
  const token =
    typeof window !== 'undefined' ? localStorage.getItem('auth_token') : null;
  const headers = (prevContext.headers ?? {}) as Record<string, string>;
  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : '',
    },
  };
});

export const apolloClient = new ApolloClient({
  link: authLink.concat(retryLink.concat(httpLink)),
  cache: new InMemoryCache(),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: 'cache-and-network',
    },
  },
});
