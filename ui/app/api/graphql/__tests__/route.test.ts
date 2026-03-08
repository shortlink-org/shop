import { beforeEach, describe, expect, it, vi } from 'vitest';

import { POST } from '../route';

describe('POST /api/graphql sanitize goods page variable', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('coerces UUID-like page value to 1 for GetGoodsList', async () => {
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
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
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
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
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
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

  it('forwards x-user-id to the BFF as X-User-ID', async () => {
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ data: { getCart: { state: null } } }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' }
        })
      );

    const req = {
      headers: new Headers({
        'content-type': 'application/json',
        'x-user-id': '550e8400-e29b-41d4-a716-446655440000'
      }),
      text: vi.fn().mockResolvedValue(
        JSON.stringify({
          query: 'query GetCart { getCart { state { cartId } } }'
        })
      )
    } as never;

    await POST(req);

    const forwardedHeaders = new Headers(fetchMock.mock.calls[0]?.[1]?.headers);
    expect(forwardedHeaders.get('X-User-ID')).toBe('550e8400-e29b-41d4-a716-446655440000');
  });

  it('forwards traceparent to the BFF and echoes trace-id in the response', async () => {
    const fetchMock = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
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
});
