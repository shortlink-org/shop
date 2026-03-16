import { beforeEach, describe, expect, it, vi } from 'vitest';

const activeSpan = {
  setAttribute: vi.fn(),
  setStatus: vi.fn(),
  recordException: vi.fn(),
  spanContext: vi.fn(() => ({
    traceId: '4bf92f3577b34da6a3ce929d0e0e4736'
  }))
};

vi.mock('@opentelemetry/api', () => ({
  SpanStatusCode: {
    ERROR: 2
  },
  context: {
    active: () => ({})
  },
  propagation: {
    inject: (_ctx: unknown, carrier: Record<string, string>) => {
      carrier.traceparent = '00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01';
    }
  },
  trace: {
    getActiveSpan: () => activeSpan
  }
}));

import { POST } from '../route';

describe('POST /api/graphql sanitize goods page variable', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    activeSpan.setAttribute.mockClear();
    activeSpan.setStatus.mockClear();
    activeSpan.recordException.mockClear();
    activeSpan.spanContext.mockClear();
  });

  it('coerces UUID-like page value to 1 for GetGoodsList', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { goods: { results: [] } } }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' }
      })
    );

    const req = {
      headers: new Headers({ 'content-type': 'application/json' }),
      text: vi.fn().mockResolvedValue(
        JSON.stringify({
          query: 'query GetGoodsList($page: Int = 1) { goods(page: $page) { count } }',
          variables: { page: '""243fde0d-8120-40d2-b899-c6cd158692bb""' }
        })
      )
    } as never;

    await POST(req);

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const body = JSON.parse(String(fetchMock.mock.calls[0]?.[1]?.body));
    expect(body.variables.page).toBe(1);
  });

  it('coerces numeric string page to integer for GetGoodsList', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { goods: { results: [] } } }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' }
      })
    );

    const req = {
      headers: new Headers({ 'content-type': 'application/json' }),
      text: vi.fn().mockResolvedValue(
        JSON.stringify({
          query: 'query GetGoodsList($page: Int = 1) { goods(page: $page) { count } }',
          variables: { page: '2' }
        })
      )
    } as never;

    await POST(req);

    const body = JSON.parse(String(fetchMock.mock.calls[0]?.[1]?.body));
    expect(body.variables.page).toBe(2);
  });

  it('sanitizes non-GetGoodsList queries when they include goods(page: $page)', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { goods: { results: [] } } }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' }
      })
    );

    const rawBody = JSON.stringify({
      query: 'query GetGoods($page: Int) { goods(page: $page) { count } }',
      variables: { page: '""243fde0d-8120-40d2-b899-c6cd158692bb""' }
    });

    const req = {
      headers: new Headers({ 'content-type': 'application/json' }),
      text: vi.fn().mockResolvedValue(rawBody)
    } as never;

    await POST(req);

    const body = JSON.parse(String(fetchMock.mock.calls[0]?.[1]?.body));
    expect(body.variables.page).toBe(1);
  });

  it('forwards Authorization to the BFF (x-user-id is set by Istio from JWT on subgraphs)', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { getCart: { state: null } } }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' }
      })
    );

    const req = {
      headers: new Headers({
        'content-type': 'application/json',
        authorization:
          'Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1NTBlODQwMC1lMjllLTQxZDQtYTcxNi00NDY2NTU0NDAwMDAifQ.x'
      }),
      text: vi.fn().mockResolvedValue(
        JSON.stringify({
          query: 'query GetCart { getCart { state { cartId } } }'
        })
      )
    } as never;

    await POST(req);

    const forwardedHeaders = new Headers(fetchMock.mock.calls[0]?.[1]?.headers);
    expect(forwardedHeaders.get('Authorization')).toMatch(/^Bearer /);
    expect(forwardedHeaders.get('X-User-ID')).toBeNull();
  });

  it('forwards Authorization to the BFF', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { addItem: { _: true } } }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' }
      })
    );

    const req = {
      headers: new Headers({
        'content-type': 'application/json',
        authorization: 'Bearer token',
        'x-user-id': '550e8400-e29b-41d4-a716-446655440000'
      }),
      text: vi.fn().mockResolvedValue(
        JSON.stringify({
          query:
            'mutation AddToCart($addRequest: ItemRequest!) { addItem(addRequest: $addRequest) { _ } }',
          variables: { addRequest: { items: [{ goodId: '1', quantity: 1 }] } }
        })
      )
    } as never;

    await POST(req);

    const forwardedHeaders = new Headers(fetchMock.mock.calls[0]?.[1]?.headers);
    expect(forwardedHeaders.get('Authorization')).toBe('Bearer token');
    expect(forwardedHeaders.get('X-User-ID')).toBeNull();
  });

  it('forwards traceparent to the BFF and echoes trace-id in the response', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { goods: { results: [] } } }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' }
      })
    );

    const req = {
      headers: new Headers({
        'content-type': 'application/json',
        traceparent: '00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01'
      }),
      text: vi.fn().mockResolvedValue(
        JSON.stringify({
          query: 'query GetGoodsList { goods { count } }'
        })
      )
    } as never;

    const response = await POST(req);

    const forwardedHeaders = new Headers(fetchMock.mock.calls[0]?.[1]?.headers);
    expect(forwardedHeaders.get('traceparent')).toBe(
      '00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01'
    );
    expect(forwardedHeaders.get('trace-id')).toBe('4bf92f3577b34da6a3ce929d0e0e4736');
    expect(response.headers.get('trace-id')).toBe('4bf92f3577b34da6a3ce929d0e0e4736');
  });

  it('injects traceparent from the active span when the request has none', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { goods: { results: [] } } }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' }
      })
    );

    const req = {
      headers: new Headers({
        'content-type': 'application/json'
      }),
      text: vi.fn().mockResolvedValue(
        JSON.stringify({
          query: 'query GetGoodsList { goods { count } }'
        })
      )
    } as never;

    const response = await POST(req);

    const forwardedHeaders = new Headers(fetchMock.mock.calls[0]?.[1]?.headers);
    expect(forwardedHeaders.get('traceparent')).toBe(
      '00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01'
    );
    expect(forwardedHeaders.get('trace-id')).toBe('4bf92f3577b34da6a3ce929d0e0e4736');
    expect(response.headers.get('trace-id')).toBe('4bf92f3577b34da6a3ce929d0e0e4736');
  });

  it('marks the active span as error for GraphQL errors', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ errors: [{ message: "Failed to fetch from Subgraph 'carts'." }] }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' }
      })
    );

    const req = {
      headers: new Headers({ 'content-type': 'application/json' }),
      text: vi.fn().mockResolvedValue(
        JSON.stringify({
          query: 'query GetCart { getCart { state { cartId } } }'
        })
      )
    } as never;

    await POST(req);

    expect(activeSpan.setStatus).toHaveBeenCalledWith({
      code: 2,
      message: "Failed to fetch from Subgraph 'carts'."
    });
    expect(activeSpan.recordException).toHaveBeenCalled();
  });
});
