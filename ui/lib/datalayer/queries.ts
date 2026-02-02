import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import type { Session } from '@ory/client';

import { fetchSession } from './api';

export const queryKeys = {
  session: ['session'] as const,
};

export function useSessionQuery(): UseQueryResult<Session> {
  return useQuery<Session>({
    queryKey: queryKeys.session,
    queryFn: () => fetchSession(),
    staleTime: 5 * 60 * 1000,
    retry: 1,
  }) as UseQueryResult<Session>;
}

export function useOptionalSessionQuery(): UseQueryResult<Session | null> {
  return useQuery<Session | null>({
    queryKey: queryKeys.session,
    queryFn: async () => {
      try {
        return await fetchSession();
      } catch {
        return null;
      }
    },
    staleTime: 5 * 60 * 1000,
    retry: false,
  }) as UseQueryResult<Session | null>;
}
