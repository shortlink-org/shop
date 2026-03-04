import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const TRACE_ID_HEADER = 'trace-id';

/**
 * Get trace ID from request: use incoming header or W3C traceparent, otherwise generate.
 */
function getTraceId(request: NextRequest): string {
  const traceId = request.headers.get(TRACE_ID_HEADER)
    ?? request.headers.get('x-trace-id')
    ?? request.headers.get('x-request-id');
  if (traceId) return traceId;

  const traceparent = request.headers.get('traceparent');
  if (traceparent) {
    const parts = traceparent.split('-');
    if (parts.length >= 2 && parts[1].length === 32) return parts[1];
  }

  return crypto.randomUUID().replace(/-/g, '');
}

export function middleware(request: NextRequest) {
  const traceId = getTraceId(request);
  const response = NextResponse.next();
  response.headers.set(TRACE_ID_HEADER, traceId);
  return response;
}
