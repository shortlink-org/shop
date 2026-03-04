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
    const traceIdFromParent = parts[1];
    if (traceIdFromParent && traceIdFromParent.length === 32) return traceIdFromParent;
  }

  return crypto.randomUUID().replace(/-/g, '');
}

export function middleware(request: NextRequest) {
  const traceId = getTraceId(request);
  const response = NextResponse.next();
  response.headers.set(TRACE_ID_HEADER, traceId);
  return response;
}
