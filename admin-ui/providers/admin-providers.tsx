'use client';

import { ApolloProvider } from '@apollo/client/react';
import { ThemeProvider } from 'next-themes';
import { Toaster } from 'sonner';
import type { ReactNode } from 'react';

import { SessionWrapper } from '@/components/auth/SessionWrapper';
import { apolloClient } from '@/lib/apollo-client';

export function AdminProviders({ children }: { children: ReactNode }) {
  return (
    <ApolloProvider client={apolloClient}>
      <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
        <SessionWrapper requireAuth={false}>
          {children}
          <Toaster closeButton richColors />
        </SessionWrapper>
      </ThemeProvider>
    </ApolloProvider>
  );
}
