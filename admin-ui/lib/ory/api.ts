/**
 * Ory API functions for session management
 */

import { Session } from '@ory/client';
import ory from './sdk';

/**
 * Fetch current user session from Ory Kratos
 * @throws Error if not authenticated
 */
export async function fetchSession(): Promise<Session> {
  const { data } = await ory.toSession();
  return data;
}

/**
 * Fetch session without throwing (returns null if not authenticated)
 */
export async function fetchSessionOptional(): Promise<Session | null> {
  try {
    const { data } = await ory.toSession();
    return data;
  } catch {
    return null;
  }
}

/**
 * Create logout URL and redirect
 */
export async function logout(): Promise<void> {
  try {
    const { data } = await ory.createBrowserLogoutFlow();
    window.location.href = data.logout_url;
  } catch (error) {
    console.error('Logout error:', error);
    // Fallback: redirect to login page
    window.location.href = '/';
  }
}

/**
 * Get login URL
 */
export function getLoginUrl(returnTo?: string): string {
  const baseUrl = process.env.NEXT_PUBLIC_ORY_SDK_URL || 'http://localhost:4433';
  const url = new URL(`${baseUrl}/self-service/login/browser`);
  if (returnTo) {
    url.searchParams.set('return_to', returnTo);
  }
  return url.toString();
}

/**
 * Check if user has admin role
 */
export function isAdmin(session: Session | null): boolean {
  if (!session?.identity?.traits) {
    return false;
  }
  
  // Check for admin role in identity traits
  const traits = session.identity.traits as Record<string, unknown>;
  return traits.role === 'admin' || traits.is_admin === true;
}

/** Name trait can be string or Ory-style { first, last } */
function nameTraitsToString(name: unknown): string {
  if (typeof name === 'string') return name;
  if (name && typeof name === 'object' && 'first' in name && 'last' in name) {
    const n = name as { first?: string; last?: string };
    return [n.first, n.last].filter(Boolean).join(' ').trim() || '';
  }
  return '';
}

/**
 * Get user display name from session
 */
export function getUserName(session: Session | null): string {
  if (!session?.identity?.traits) {
    return 'Unknown';
  }

  const traits = session.identity.traits as Record<string, unknown>;
  const nameStr = nameTraitsToString(traits.name);
  return nameStr || (traits.email as string) || 'User';
}

/**
 * Get user email from session
 */
export function getUserEmail(session: Session | null): string | undefined {
  if (!session?.identity?.traits) {
    return undefined;
  }
  
  const traits = session.identity.traits as Record<string, unknown>;
  return traits.email as string | undefined;
}
