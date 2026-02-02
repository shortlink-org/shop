/**
 * Data layer for API requests
 */

import { Session } from '@ory/client';
import ory from '@/lib/ory/sdk';

/**
 * Fetch user session
 * Cached for 5 minutes
 */
export async function fetchSession(): Promise<Session> {
  const { data } = await ory.toSession();
  return data;
}
