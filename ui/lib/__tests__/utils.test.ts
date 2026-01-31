import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';
import { createUrl, ensureStartsWith, validateEnvironmentVariables } from '../utils';

describe('createUrl', () => {
  it('should create URL with query parameters', () => {
    const params = new URLSearchParams({ page: '1', sort: 'asc' });
    const result = createUrl('/products', params);
    expect(result).toBe('/products?page=1&sort=asc');
  });

  it('should create URL without query string when params are empty', () => {
    const params = new URLSearchParams();
    const result = createUrl('/products', params);
    expect(result).toBe('/products');
  });

  it('should handle single parameter', () => {
    const params = new URLSearchParams({ search: 'test' });
    const result = createUrl('/search', params);
    expect(result).toBe('/search?search=test');
  });

  it('should handle special characters in params', () => {
    const params = new URLSearchParams({ query: 'hello world' });
    const result = createUrl('/search', params);
    expect(result).toBe('/search?query=hello+world');
  });

  it('should handle root pathname', () => {
    const params = new URLSearchParams({ id: '123' });
    const result = createUrl('/', params);
    expect(result).toBe('/?id=123');
  });
});

describe('ensureStartsWith', () => {
  it('should add prefix if string does not start with it', () => {
    const result = ensureStartsWith('example.com', 'https://');
    expect(result).toBe('https://example.com');
  });

  it('should not add prefix if string already starts with it', () => {
    const result = ensureStartsWith('https://example.com', 'https://');
    expect(result).toBe('https://example.com');
  });

  it('should handle empty string', () => {
    const result = ensureStartsWith('', '/');
    expect(result).toBe('/');
  });

  it('should handle prefix that is longer than string', () => {
    const result = ensureStartsWith('a', 'prefix');
    expect(result).toBe('prefixa');
  });

  it('should work with slash prefix', () => {
    const result = ensureStartsWith('path/to/resource', '/');
    expect(result).toBe('/path/to/resource');
  });

  it('should not duplicate slash prefix', () => {
    const result = ensureStartsWith('/path/to/resource', '/');
    expect(result).toBe('/path/to/resource');
  });
});

describe('validateEnvironmentVariables', () => {
  const originalEnv = process.env;

  beforeEach(() => {
    vi.resetModules();
    process.env = { ...originalEnv };
  });

  afterEach(() => {
    process.env = originalEnv;
  });

  it('should throw error when API_URI is missing', () => {
    delete process.env.API_URI;
    delete process.env.SHOPIFY_STOREFRONT_ACCESS_TOKEN;

    expect(() => validateEnvironmentVariables()).toThrow(
      /The following environment variables are missing/
    );
  });

  it('should throw error when API_URI contains brackets', () => {
    process.env.API_URI = 'https://[store].myshopify.com';
    process.env.SHOPIFY_STOREFRONT_ACCESS_TOKEN = 'test-token';

    expect(() => validateEnvironmentVariables()).toThrow(
      /Your `API_URI` environment variable includes brackets/
    );
  });

  it('should not throw when all required variables are present', () => {
    process.env.API_URI = 'https://store.myshopify.com';
    process.env.SHOPIFY_STOREFRONT_ACCESS_TOKEN = 'test-token';

    expect(() => validateEnvironmentVariables()).not.toThrow();
  });

  it('should throw error listing all missing variables', () => {
    delete process.env.API_URI;
    delete process.env.SHOPIFY_STOREFRONT_ACCESS_TOKEN;

    expect(() => validateEnvironmentVariables()).toThrow(/API_URI/);
    expect(() => validateEnvironmentVariables()).toThrow(/SHOPIFY_STOREFRONT_ACCESS_TOKEN/);
  });
});
