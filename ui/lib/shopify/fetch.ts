import { HTTP_STATUS_RATE_LIMIT, RATE_LIMIT_MESSAGE } from 'lib/constants';
import { isShopifyError } from 'lib/type-guards';
import { getGraphqlEndpoint } from './config';

type ExtractVariables<T> = T extends { variables: object } ? T['variables'] : never;
type GraphqlErrorDetails = {
  message?: string;
  cause?: unknown;
  status?: number;
  path?: unknown;
  extensions?: Record<string, unknown>;
};

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

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function getNestedProperty(value: unknown, key: string): unknown {
  if (!isRecord(value)) return undefined;
  if (key in value) return value[key];
  if ('error' in value) return getNestedProperty(value.error, key);
  return undefined;
}

function normalizeGraphqlError(
  error: GraphqlErrorDetails,
  query: string,
  status: number,
  traceId?: string
): Record<string, unknown> {
  return {
    cause: error.cause?.toString() || 'unknown',
    status: error.status || status || 500,
    message: error.message || 'Request failed',
    query,
    ...(error.path !== undefined ? { path: error.path } : {}),
    ...(error.extensions ? { extensions: error.extensions } : {}),
    ...(traceId ? { traceId } : {}),
    error
  };
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
  let traceId: string | undefined;

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
    traceId = result.headers.get('trace-id') ?? undefined;

    const text = await result.text();

    if (result.status === HTTP_STATUS_RATE_LIMIT) {
      const retryAfter = parseRetryAfter(result.headers.get('Retry-After'));
      throw {
        status: HTTP_STATUS_RATE_LIMIT,
        message: RATE_LIMIT_MESSAGE,
        ...(retryAfter !== undefined && { retryAfter }),
        query,
        ...(traceId ? { traceId } : {})
      };
    }

    let body: T & {
      error?: GraphqlErrorDetails;
      errors?: GraphqlErrorDetails[];
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
        query,
        ...(traceId ? { traceId } : {})
      };
    }

    const firstError = body.errors?.[0];
    if (firstError) {
      throw normalizeGraphqlError(firstError, query, result.status, traceId);
    }

    const bffMessage = body.error?.message ?? '';
    const isUpstreamUnavailable =
      bffMessage.toLowerCase().includes('no healthy upstream') ||
      bffMessage.toLowerCase().includes('failed to fetch from subgraph');
    if (body.error && isUpstreamUnavailable) {
      throw {
        ...normalizeGraphqlError(body.error, query, result.status, traceId),
        status: result.status || 500,
        message: upstreamUnavailableMessage,
      };
    }
    if (body.error) {
      throw normalizeGraphqlError(
        { ...body.error, message: bffMessage || 'Request failed' },
        query,
        result.status,
        traceId
      );
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
        query,
        ...(traceId ? { traceId } : {})
      };
    }

    if (
      typeof e === 'object' &&
      e !== null &&
      'status' in e &&
      'message' in e &&
      'query' in e
    ) {
      throw {
        ...e,
        ...(traceId && !('traceId' in e) ? { traceId } : {})
      };
    }

    throw {
      error: e,
      query,
      ...(traceId ? { traceId } : {}),
      ...(getNestedProperty(e, 'path') !== undefined ? { path: getNestedProperty(e, 'path') } : {}),
      ...(getNestedProperty(e, 'extensions') !== undefined
        ? { extensions: getNestedProperty(e, 'extensions') }
        : {})
    };
  }
}
