import { SpanStatusCode, context as otelContext, propagation, trace } from '@opentelemetry/api';
import { NextRequest, NextResponse } from 'next/server';

const BFF_GRAPHQL_URL =
  process.env.BFF_GRAPHQL_URL ??
  'http://shortlink-shop-bff.shortlink-shop.svc.cluster.local:9991/graphql';
const TRACE_ID_HEADER = 'trace-id';
const TRACEPARENT_HEADER = 'traceparent';

function normalizePage(value: unknown): number {
  if (typeof value === 'number' && Number.isInteger(value) && value > 0) return value;
  if (typeof value === 'string') {
    const trimmed = value.trim().replace(/^"+|"+$/g, '');
    if (/^[1-9]\d*$/.test(trimmed)) return Number(trimmed);
  }
  return 1;
}

function parseTraceIdFromTraceparent(traceparent: string | undefined): string | undefined {
  if (!traceparent) {
    return undefined;
  }

  const parts = traceparent.split('-');
  const traceId = parts[1];
  return traceId && traceId.length === 32 ? traceId : undefined;
}

function getTraceContext(request: NextRequest): { traceId?: string; traceparent?: string } {
  const requestTraceparent = request.headers.get(TRACEPARENT_HEADER) ?? undefined;
  const headerTraceId =
    request.headers.get(TRACE_ID_HEADER) ??
    request.headers.get('x-trace-id') ??
    request.headers.get('x-request-id');

  if (requestTraceparent || headerTraceId) {
    return {
      traceparent: requestTraceparent,
      traceId: headerTraceId ?? parseTraceIdFromTraceparent(requestTraceparent)
    };
  }

  const carrier: Record<string, string> = {};
  propagation.inject(otelContext.active(), carrier);

  const activeTraceparent = carrier[TRACEPARENT_HEADER];
  const activeSpanTraceId = trace.getActiveSpan()?.spanContext?.().traceId;

  if (activeTraceparent || activeSpanTraceId) {
    return {
      traceparent: activeTraceparent,
      traceId: activeSpanTraceId ?? parseTraceIdFromTraceparent(activeTraceparent)
    };
  }

  return {};
}

function annotateActiveSpan(
  operationName: string | null,
  traceId: string | undefined,
  status: number,
  errorMessage?: string,
  graphqlErrors?: string[]
): void {
  const activeSpan = trace.getActiveSpan();
  if (!activeSpan) return;

  activeSpan.setAttribute('graphql.operation.name', operationName ?? '(anonymous)');
  activeSpan.setAttribute('graphql.response.status', status);
  if (traceId) {
    activeSpan.setAttribute('graphql.trace_id', traceId);
  }

  if (graphqlErrors?.length) {
    activeSpan.setAttribute('graphql.errors.count', graphqlErrors.length);
  }

  if (!errorMessage) {
    return;
  }

  activeSpan.setStatus({
    code: SpanStatusCode.ERROR,
    message: errorMessage
  });
  activeSpan.recordException(new Error(errorMessage));
}

/** If the request is GetGoodsList (goods(page: $page)), coerce variables.page to Int to avoid BFF error. */
function sanitizeGetGoodsListBody(rawBody: string): string {
  try {
    const payload = JSON.parse(rawBody) as { query?: string; variables?: Record<string, unknown> };
    const query = typeof payload.query === 'string' ? payload.query : '';
    // Sanitize for any operation that calls goods(page: ...), not only GetGoodsList.
    if (!/goods\s*\(\s*page\s*:/.test(query)) return rawBody;
    const variables =
      payload.variables && typeof payload.variables === 'object' ? { ...payload.variables } : {};
    const page = variables.page;
    const safePage = normalizePage(page);
    if (page !== safePage) variables.page = safePage;
    return JSON.stringify({ ...payload, variables });
  } catch {
    return rawBody;
  }
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  const authorization = req.headers.get('authorization') ?? undefined;
  const contentType = req.headers.get('content-type') ?? 'application/json';
  const { traceId, traceparent } = getTraceContext(req);

  let body: string;
  try {
    body = await req.text();
  } catch {
    return NextResponse.json({ errors: [{ message: 'Invalid request body' }] }, { status: 400 });
  }

  let operationName: string | null = null;
  try {
    const payload = JSON.parse(body) as {
      query?: string;
      operationName?: string;
      variables?: Record<string, unknown>;
    };
    operationName =
      typeof payload.operationName === 'string' && payload.operationName
        ? payload.operationName
        : (payload.query?.match(/(?:query|mutation)\s+(\w+)/)?.[1] ?? null);
  } catch {
    // ignore
  }

  const sanitizedBody = sanitizeGetGoodsListBody(body);
  if (sanitizedBody !== body) {
    try {
      const before = JSON.parse(body) as { variables?: Record<string, unknown> };
      const after = JSON.parse(sanitizedBody) as { variables?: Record<string, unknown> };
      if (before?.variables?.page !== after?.variables?.page) {
        console.warn('[api/graphql] Sanitized goods page variable', {
          originalPage: before?.variables?.page,
          safePage: after?.variables?.page,
          userAgent: req.headers.get('user-agent') ?? '',
          forwardedFor: req.headers.get('x-forwarded-for') ?? ''
        });
      }
    } catch {
      // no-op
    }
  }
  body = sanitizedBody;

  const headers: HeadersInit = {
    'Content-Type': contentType
  };
  if (authorization) {
    headers['Authorization'] = authorization;
  }
  if (traceparent) {
    headers[TRACEPARENT_HEADER] = traceparent;
  }
  if (traceId) {
    headers[TRACE_ID_HEADER] = traceId;
  }

  let res: Response;
  try {
    res = await fetch(BFF_GRAPHQL_URL, {
      method: 'POST',
      headers,
      body
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Failed to reach BFF GraphQL endpoint';
    annotateActiveSpan(operationName, traceId, 0, message);
    throw error;
  }

  const text = await res.text();
  let graphqlErrors: string[] = [];
  let spanErrorMessage: string | undefined;

  // Log incoming request and outgoing response (with errors when present)
  const logPayload: Record<string, unknown> = {
    operation: operationName ?? '(anonymous)',
    traceId: traceId ?? null,
    status: res.status
  };
  if (res.status >= 400) {
    spanErrorMessage = text.slice(0, 500) || `BFF returned HTTP ${res.status}`;
    logPayload.responseError = spanErrorMessage;
  } else {
    try {
      const parsed = JSON.parse(text) as {
        errors?: Array<{ message?: string }>;
        error?: {
          message?: string;
          extensions?: { errors?: Array<{ message?: string }> };
        };
      };
      if (Array.isArray(parsed?.errors) && parsed.errors.length > 0) {
        graphqlErrors = parsed.errors.map((e) => e?.message ?? String(e));
      } else if (parsed?.error?.message) {
        graphqlErrors = [
          parsed.error.message,
          ...(parsed.error.extensions?.errors?.map((e) => e?.message ?? String(e)) ?? [])
        ];
      }

      if (graphqlErrors.length > 0) {
        spanErrorMessage = graphqlErrors[0];
        logPayload.graphqlErrors = graphqlErrors;
      }
    } catch {
      // not JSON or no errors
    }
  }
  annotateActiveSpan(operationName, traceId, res.status, spanErrorMessage, graphqlErrors);
  console.info('[api/graphql]', logPayload);

  return new NextResponse(text, {
    status: res.status,
    headers: {
      'Content-Type': res.headers.get('Content-Type') ?? 'application/json',
      ...(res.headers.get(TRACE_ID_HEADER)
        ? { [TRACE_ID_HEADER]: res.headers.get(TRACE_ID_HEADER) as string }
        : traceId
          ? { [TRACE_ID_HEADER]: traceId }
          : {})
    }
  });
}

export async function OPTIONS(): Promise<NextResponse> {
  return new NextResponse(null, { status: 204 });
}
