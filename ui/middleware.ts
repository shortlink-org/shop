import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const TRACE_ID_HEADER = 'trace-id';
const TRACEPARENT_HEADER = 'traceparent';

/** W3C/OpenTelemetry trace-id is 32 hex chars. Normalize UUID or other formats to that. */
function normalizeTraceId(value: string): string {
  const hex = value.replace(/-/g, '').toLowerCase();
  return /^[0-9a-f]{32}$/.test(hex) ? hex : crypto.randomUUID().replace(/-/g, '');
}

/** Generate 16 hex chars for W3C span-id. */
function randomSpanId(): string {
  return crypto.getRandomValues(new Uint8Array(8))
    .reduce((s, b) => s + b.toString(16).padStart(2, '0'), '');
}

/**
 * Get trace ID and optional traceparent from request.
 * Returns [traceId, traceparent to set on request or null if already present].
 */
function getTraceContext(request: NextRequest): { traceId: string; traceparent: string } {
  const existing = request.headers.get(TRACEPARENT_HEADER);
  if (existing) {
    const parts = existing.split('-');
    const traceId = parts[1];
    if (traceId && traceId.length === 32) {
      return { traceId, traceparent: existing };
    }
  }

  const fromHeader =
    request.headers.get(TRACE_ID_HEADER) ??
    request.headers.get('x-trace-id') ??
    request.headers.get('x-request-id');
  const traceId = fromHeader ? normalizeTraceId(fromHeader) : crypto.randomUUID().replace(/-/g, '');
  const traceparent = `00-${traceId}-${randomSpanId()}-01`;
  return { traceId, traceparent };
}

/**
 * Middleware: propagate or create trace context so the same trace-id is used by OpenTelemetry
 * (and exported to Tempo) and returned in the response header.
 */
export function middleware(request: NextRequest) {
  const { traceId, traceparent } = getTraceContext(request);
  const requestHeaders = new Headers(request.headers);
  requestHeaders.set(TRACEPARENT_HEADER, traceparent);

  const response = NextResponse.next({ request: { headers: requestHeaders } });
  response.headers.set(TRACE_ID_HEADER, traceId);
  return response;
}
