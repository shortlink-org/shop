import { NextRequest, NextResponse } from 'next/server';

const BFF_GRAPHQL_URL =
  process.env.BFF_GRAPHQL_URL ??
  'http://shortlink-shop-bff.shortlink-shop.svc.cluster.local:9991/graphql';

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
