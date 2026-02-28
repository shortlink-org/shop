import { HTTP_STATUS_RATE_LIMIT, RATE_LIMIT_MESSAGE } from 'lib/constants';
import { isShopifyError } from 'lib/type-guards';
import { getGraphqlEndpoint } from './config';

type ExtractVariables<T> = T extends { variables: object } ? T['variables'] : never;

function parseRetryAfter(header: string | null): number | undefined {
  if (!header?.trim()) return undefined;
  const seconds = parseInt(header, 10);
  if (!Number.isNaN(seconds) && seconds >= 0) return seconds;
  const date = Date.parse(header);
  if (!Number.isNaN(date)) {
    const delay = Math.ceil((date - Date.now()) / 1000);
    return delay > 0 ? delay : undefined;
  }
  return undefined;
}

export async function shopifyFetch<T>({
  cache = 'force-cache',
  headers,
  query,
  tags,
  variables
}: {
  cache?: RequestCache;
  headers?: HeadersInit;
  query: string;
  tags?: string[];
  variables?: ExtractVariables<T>;
}): Promise<{ status: number; body: T } | never> {
  const endpoint = getGraphqlEndpoint();
  const upstreamUnavailableMessage =
    'Service temporarily unavailable. Please try again in a moment.';

  try {
    const result = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...headers
      },
      body: JSON.stringify({
        ...(query && { query }),
        ...(variables && { variables })
      }),
      cache,
    });

    const text = await result.text();

    if (result.status === HTTP_STATUS_RATE_LIMIT) {
      const retryAfter = parseRetryAfter(result.headers.get('Retry-After'));
      throw {
        status: HTTP_STATUS_RATE_LIMIT,
        message: RATE_LIMIT_MESSAGE,
        ...(retryAfter !== undefined && { retryAfter }),
        query
      };
    }

    let body: T & {
      error?: { message?: string };
      errors?: Array<{ message: string; cause?: unknown; status?: number }>;
    };

    try {
      body = JSON.parse(text) as typeof body;
    } catch {
      const rawMessage = text?.trim() || 'Invalid response from server';
      const message =
        rawMessage.toLowerCase().includes('no healthy upstream')
          ? upstreamUnavailableMessage
          : rawMessage;
      throw {
        cause: 'unknown',
        status: result.status || 500,
        message,
        query
      };
    }

    if (body.errors) {
      throw body.errors[0];
    }

    const bffMessage = body.error?.message ?? '';
    const isUpstreamUnavailable =
      bffMessage.toLowerCase().includes('no healthy upstream') ||
      bffMessage.toLowerCase().includes('failed to fetch from subgraph');
    if (body.error && isUpstreamUnavailable) {
      throw {
        cause: 'unknown',
        status: result.status || 500,
        message: upstreamUnavailableMessage,
        query
      };
    }
    if (body.error) {
      throw {
        cause: 'unknown',
        status: result.status || 500,
        message: bffMessage || 'Request failed',
        query
      };
    }

    return {
      status: result.status,
      body
    };
  } catch (e) {
    if (isShopifyError(e)) {
      throw {
        cause: e.cause?.toString() || 'unknown',
        status: e.status || 500,
        message: e.message,
        query
      };
    }

    if (
      typeof e === 'object' &&
      e !== null &&
      'status' in e &&
      'message' in e &&
      'query' in e
    ) {
      throw e;
    }

    throw {
      error: e,
      query
    };
  }
}
