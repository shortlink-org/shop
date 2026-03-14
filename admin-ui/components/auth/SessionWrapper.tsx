'use client';

/**
 * SessionWrapper - Session provider for admin pages
 * 
 * Fetches session on mount and provides it to children
 * Redirects to login if not authenticated
 */

import { Button, FeedbackPanel } from '@shortlink-org/ui-kit';
import React, { ReactNode, useState, useEffect } from 'react';
import { Session } from '@ory/client';

import { SessionProvider } from '@/contexts/SessionContext';
import { fetchSessionOptional, logout, getLoginUrl } from '@/lib/ory/api';

interface SessionWrapperProps {
  children: ReactNode;
  requireAuth?: boolean;
}

export function SessionWrapper({ children, requireAuth = true }: SessionWrapperProps) {
  const [session, setSession] = useState<Session | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const loadSession = async () => {
      try {
        const sessionData = await fetchSessionOptional();
        setSession(sessionData);
        
        // Redirect to login if auth required and no session
        if (requireAuth && !sessionData) {
          window.location.href = getLoginUrl(window.location.href);
          return;
        }
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to load session'));
      } finally {
        setIsLoading(false);
      }
    };

    loadSession();
  }, [requireAuth]);

  // Loading state
  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center px-4">
        <div className="admin-card w-full max-w-xl p-8">
          <FeedbackPanel
            variant="loading"
            eyebrow="Authorization"
            title="Loading session"
            message="Checking your Ory session before rendering the admin workspace."
          />
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center px-4 py-10">
        <div className="admin-card w-full max-w-2xl p-8">
          <FeedbackPanel
            variant="error"
            eyebrow="Authorization"
            title="Authorization error"
            message={error.message}
            action={
              <div className="flex flex-wrap gap-3">
                <Button as="a" asProps={{ href: getLoginUrl() }}>
                  Sign in
                </Button>
                <Button variant="secondary" onClick={() => window.location.reload()}>
                  Retry
                </Button>
              </div>
            }
          />
        </div>
      </div>
    );
  }

  // Not authenticated (and requireAuth is true)
  if (requireAuth && !session) {
    return (
      <div className="flex min-h-screen items-center justify-center px-4 py-10">
        <div className="admin-card w-full max-w-2xl p-8">
          <FeedbackPanel
            variant="empty"
            eyebrow="Authorization"
            title="Authorization required"
            message="Please sign in to access the admin panel."
            action={
              <Button as="a" asProps={{ href: getLoginUrl() }}>
                Sign in
              </Button>
            }
          />
        </div>
      </div>
    );
  }

  return (
    <SessionProvider 
      session={session} 
      isLoading={false}
      error={error}
      onLogout={logout}
    >
      {children}
    </SessionProvider>
  );
}

export default SessionWrapper;
