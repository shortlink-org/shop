import { NextRequest, NextResponse } from 'next/server';

const BFF_GRAPHQL_URL =
  process.env.BFF_GRAPHQL_URL ??
  'http://shortlink-shop-bff.shortlink-shop.svc.cluster.local:9991/graphql';

function normalizePage(value: unknown): number {
  if (typeof value === 'number' && Number.isInteger(value) && value > 0) return value;
  if (typeof value === 'string') {
    const trimmed = value.trim().replace(/^"+|"+$/g, '');
    if (/^[1-9]\d*$/.test(trimmed)) return Number(trimmed);
  }
  return 1;
}

/** If the request is GetGoodsList (goods(page: $page)), coerce variables.page to Int to avoid BFF error. */
function sanitizeGetGoodsListBody(rawBody: string): string {
  try {
    const payload = JSON.parse(rawBody) as { query?: string; variables?: Record<string, unknown> };
    const query = typeof payload.query === 'string' ? payload.query : '';
    // Sanitize for any operation that calls goods(page: ...), not only GetGoodsList.
    if (!/goods\s*\(\s*page\s*:/.test(query)) return rawBody;
    const variables = payload.variables && typeof payload.variables === 'object' ? { ...payload.variables } : {};
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

  let body: string;
  try {
    body = await req.text();
  } catch {
    return NextResponse.json(
      { errors: [{ message: 'Invalid request body' }] },
      { status: 400 }
    );
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

  const res = await fetch(BFF_GRAPHQL_URL, {
    method: 'POST',
    headers,
    body
  });

  const text = await res.text();
  return new NextResponse(text, {
    status: res.status,
    headers: {
      'Content-Type': res.headers.get('Content-Type') ?? 'application/json'
    }
  });
}

export async function OPTIONS(): Promise<NextResponse> {
  return new NextResponse(null, { status: 204 });
}
