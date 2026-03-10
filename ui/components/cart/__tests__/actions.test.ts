import { beforeEach, describe, expect, it, vi } from 'vitest';

const {
  addToCartMock,
  updateCartMock,
  revalidateTagMock,
  cookiesSetMock,
  headersMock,
  cookiesMock
} = vi.hoisted(() => ({
  addToCartMock: vi.fn(),
  updateCartMock: vi.fn(),
  revalidateTagMock: vi.fn(),
  cookiesSetMock: vi.fn(),
  headersMock: vi.fn(async () => new Headers()),
  cookiesMock: vi.fn(async () => ({
    set: vi.fn()
  }))
}));

vi.mock('lib/shopify', () => ({
  addToCart: addToCartMock,
  updateCart: updateCartMock
}));

vi.mock('next/cache', () => ({
  revalidateTag: revalidateTagMock
}));

vi.mock('next/headers', () => ({
  headers: headersMock,
  cookies: cookiesMock
}));

import { addItem } from '../actions';

describe('cart actions auth propagation', () => {
  beforeEach(() => {
    addToCartMock.mockReset();
    updateCartMock.mockReset();
    revalidateTagMock.mockReset();
    cookiesSetMock.mockReset();
    headersMock.mockReset();
    cookiesMock.mockReset();
    headersMock.mockResolvedValue(new Headers());
    cookiesMock.mockResolvedValue({ set: cookiesSetMock });
  });

  it('passes Authorization and x-user-id to addToCart', async () => {
    headersMock.mockResolvedValue(
      new Headers({
        authorization: 'Bearer token',
        'x-user-id': '550e8400-e29b-41d4-a716-446655440000'
      })
    );

    addToCartMock.mockResolvedValue(undefined);

    const result = await addItem(null, '243fde0d-8120-40d2-b899-c6cd158692bb');

    expect(addToCartMock).toHaveBeenCalledWith(
      [{ goodId: '243fde0d-8120-40d2-b899-c6cd158692bb', quantity: 1 }],
      {
        authorization: 'Bearer token',
        userId: '550e8400-e29b-41d4-a716-446655440000'
      }
    );
    expect(cookiesSetMock).toHaveBeenCalledWith(
      'cartId',
      '550e8400-e29b-41d4-a716-446655440000'
    );
    expect(revalidateTagMock).toHaveBeenCalled();
    expect(result).toEqual({
      ok: true,
      cartId: '550e8400-e29b-41d4-a716-446655440000'
    });
  });
});
