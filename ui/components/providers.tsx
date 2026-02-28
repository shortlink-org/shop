'use client';

import { HTTP_STATUS_RATE_LIMIT } from 'lib/constants';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactNode, useState } from 'react';
import { SessionWrapper } from './user/session-wrapper';

const DEFAULT_RETRY_DELAY_MS = 1000;
const RATE_LIMIT_RETRY_DELAY_SEC = 5;

export function Providers({ children }: { children: ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            retry: (failureCount) => failureCount < 1,
            retryDelay: (failureCount, error) => {
              const e = error as { status?: number; retryAfter?: number };
              if (e?.status === HTTP_STATUS_RATE_LIMIT) {
                const sec = e.retryAfter ?? RATE_LIMIT_RETRY_DELAY_SEC;
                return sec * 1000;
              }
              return DEFAULT_RETRY_DELAY_MS;
            },
            refetchOnWindowFocus: false,
          },
        },
      })
  );

  return (
    <QueryClientProvider client={queryClient}>
      <SessionWrapper>{children}</SessionWrapper>
    </QueryClientProvider>
  );
}

export default Providers;
