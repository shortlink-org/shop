import { describe, expect, it, beforeEach } from 'vitest';
import { getLoginUrl, getUserName, isAdmin } from '@/lib/ory/api';
import type { Session } from '@ory/client';
import kratosSessionJson from './fixtures/kratos-session.json';

/** Real Kratos session response (name as { first, last }, email in traits) */
const kratosSessionFixture = kratosSessionJson as Session;

function mockSession(traits: Record<string, unknown>): Session {
  return {
    identity: {
      id: 'test-id',
      traits
    }
  } as Session;
}

describe('getUserName', () => {
  it('returns "Unknown" for null session', () => {
    expect(getUserName(null)).toBe('Unknown');
  });

  it('returns "User" when traits has no name or email', () => {
    expect(getUserName(mockSession({}))).toBe('User');
  });

  it('returns traits.name when it is a string', () => {
    expect(getUserName(mockSession({ name: 'Alice' }))).toBe('Alice');
  });

  it('returns concatenated string when traits.name is { first, last }', () => {
    expect(getUserName(mockSession({ name: { first: 'John', last: 'Doe' } }))).toBe('John Doe');
  });

  it('returns "first last" for real Kratos session fixture (name object)', () => {
    expect(getUserName(kratosSessionFixture)).toBe('Viktor Login');
  });

  it('falls back to email from real Kratos fixture when name would be empty', () => {
    const sessionWithoutName = {
      ...kratosSessionFixture,
      identity: {
        ...kratosSessionFixture.identity!,
        traits: { ...kratosSessionFixture.identity!.traits, name: { first: '', last: '' } }
      }
    } as Session;
    expect(getUserName(sessionWithoutName)).toBe('batazor111@gmail.com');
  });

  it('handles partial { first, last } (only first)', () => {
    expect(getUserName(mockSession({ name: { first: 'Jane', last: '' } }))).toBe('Jane');
  });

  it('handles partial { first, last } (only last)', () => {
    expect(getUserName(mockSession({ name: { first: '', last: 'Smith' } }))).toBe('Smith');
  });

  it('falls back to email when name is empty or missing', () => {
    expect(getUserName(mockSession({ email: 'user@example.com' }))).toBe('user@example.com');
  });

  it('prefers name over email when both present', () => {
    expect(
      getUserName(mockSession({ name: 'Bob', email: 'bob@example.com' }))
    ).toBe('Bob');
  });
});

describe('getLoginUrl', () => {
  const defaultBase = 'http://localhost:4433';

  beforeEach(() => {
    delete process.env.NEXT_PUBLIC_ORY_SDK_URL;
  });

  it('returns URL with login path and default base when no env', () => {
    const url = getLoginUrl();
    expect(url).toContain('/self-service/login/browser');
    expect(url.startsWith('http://localhost:4433') || url.startsWith('https://')).toBe(true);
  });

  it('includes return_to query when provided', () => {
    const returnTo = 'https://shop.example.com/callback';
    const url = getLoginUrl(returnTo);
    expect(url).toContain('/self-service/login/browser');
    expect(new URL(url).searchParams.get('return_to')).toBe(returnTo);
  });
});

describe('isAdmin', () => {
  it('returns false for null session', () => {
    expect(isAdmin(null)).toBe(false);
  });

  it('returns false when traits has no role', () => {
    expect(isAdmin(mockSession({}))).toBe(false);
  });

  it('returns true when traits.role === "admin"', () => {
    expect(isAdmin(mockSession({ role: 'admin' }))).toBe(true);
  });

  it('returns true when traits.is_admin === true', () => {
    expect(isAdmin(mockSession({ is_admin: true }))).toBe(true);
  });

  it('returns false for non-admin role', () => {
    expect(isAdmin(mockSession({ role: 'user' }))).toBe(false);
  });

  it('returns false for real Kratos fixture (no admin role in traits)', () => {
    expect(isAdmin(kratosSessionFixture)).toBe(false);
  });

  it('returns true when fixture traits include role admin', () => {
    const adminSession = {
      ...kratosSessionFixture,
      identity: {
        ...kratosSessionFixture.identity!,
        traits: { ...kratosSessionFixture.identity!.traits, role: 'admin' }
      }
    } as Session;
    expect(isAdmin(adminSession)).toBe(true);
  });
});
