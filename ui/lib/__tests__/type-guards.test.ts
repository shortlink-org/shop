import { describe, expect, it } from 'vitest';
import { isObject, isShopifyError, ShopifyErrorLike } from '../type-guards';

describe('isObject', () => {
  it('should return true for plain objects', () => {
    expect(isObject({})).toBe(true);
    expect(isObject({ key: 'value' })).toBe(true);
    expect(isObject({ nested: { object: true } })).toBe(true);
  });

  it('should return false for arrays', () => {
    expect(isObject([])).toBe(false);
    expect(isObject([1, 2, 3])).toBe(false);
    expect(isObject(['a', 'b'])).toBe(false);
  });

  it('should return false for null', () => {
    expect(isObject(null)).toBe(false);
  });

  it('should return false for undefined', () => {
    expect(isObject(undefined)).toBe(false);
  });

  it('should return false for primitives', () => {
    expect(isObject('string')).toBe(false);
    expect(isObject(123)).toBe(false);
    expect(isObject(true)).toBe(false);
    expect(isObject(Symbol('test'))).toBe(false);
  });

  it('should return false for functions', () => {
    expect(isObject(() => {})).toBe(false);
    expect(isObject(function () {})).toBe(false);
  });

  it('should return true for object created with Object.create', () => {
    expect(isObject(Object.create(null))).toBe(true);
    expect(isObject(Object.create({}))).toBe(true);
  });
});

describe('isShopifyError', () => {
  it('should return true for Error instances', () => {
    const error = new Error('Test error');
    expect(isShopifyError(error)).toBe(true);
  });

  it('should return true for TypeError instances', () => {
    const error = new TypeError('Type error');
    expect(isShopifyError(error)).toBe(true);
  });

  it('should return false for plain objects with error-like structure but no Error prototype', () => {
    // isShopifyError checks if object has Error in prototype chain, not just shape
    const shopifyLikeObject = {
      status: 400,
      message: new Error('Bad request')
    };
    // This is NOT an error because the object itself is not an Error instance
    expect(isShopifyError(shopifyLikeObject)).toBe(false);
  });

  it('should return false for null', () => {
    expect(isShopifyError(null)).toBe(false);
  });

  it('should return false for undefined', () => {
    expect(isShopifyError(undefined)).toBe(false);
  });

  it('should return false for plain objects without error properties', () => {
    expect(isShopifyError({ foo: 'bar' })).toBe(false);
  });

  it('should return false for arrays', () => {
    expect(isShopifyError([])).toBe(false);
    expect(isShopifyError([new Error()])).toBe(false);
  });

  it('should return false for primitives', () => {
    expect(isShopifyError('error')).toBe(false);
    expect(isShopifyError(500)).toBe(false);
    expect(isShopifyError(true)).toBe(false);
  });

  it('should handle objects with Error in prototype chain', () => {
    class CustomError extends Error {
      status: number;
      constructor(message: string, status: number) {
        super(message);
        this.status = status;
      }
    }
    const customError = new CustomError('Custom error', 500);
    expect(isShopifyError(customError)).toBe(true);
  });
});
