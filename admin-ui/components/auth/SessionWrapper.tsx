'use client';

/**
 * SessionWrapper - Session provider for admin pages
 * 
 * Fetches session on mount and provides it to children
 * Redirects to login if not authenticated
 */

import { Button, FeedbackPanel } from '@shortlink-org/ui-kit';
import React, { ReactNode, useEffect, useReducer } from 'react';
import { Session } from '@ory/client';

import { SessionProvider } from '@/contexts/SessionContext';
import { fetchSessionOptional, logout, getLoginUrl } from '@/lib/ory/api';

interface SessionWrapperProps {
  children: ReactNode;
  requireAuth?: boolean;
}

type SessionState = {
  session: Session | null;
  isLoading: boolean;
  error: Error | null;
};

type SessionAction =
  | { type: 'LOADED'; payload: Session | null }
  | { type: 'ERROR'; payload: Error };

function sessionReducer(state: SessionState, action: SessionAction): SessionState {
  switch (action.type) {
    case 'LOADED':
      return { session: action.payload, isLoading: false, error: null };
    case 'ERROR':
      return { ...state, error: action.payload, isLoading: false };
    default:
      return state;
  }
}

export function SessionWrapper({ children, requireAuth = true }: SessionWrapperProps) {
  const [state, dispatch] = useReducer(sessionReducer, {
    session: null,
    isLoading: true,
    error: null
  });

  useEffect(() => {
    const loadSession = async () => {
      try {
        const sessionData = await fetchSessionOptional();

        if (requireAuth && !sessionData) {
          window.location.href = getLoginUrl(window.location.href);
          return;
        }

        queueMicrotask(() => dispatch({ type: 'LOADED', payload: sessionData }));
      } catch (err) {
        queueMicrotask(() =>
          dispatch({
            type: 'ERROR',
            payload: err instanceof Error ? err : new Error('Failed to load session')
          })
        );
      }
    };

    loadSession();
  }, [requireAuth]);

  const { session, isLoading, error } = state;

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
      session={state.session}
      isLoading={false}
      error={state.error}
      onLogout={logout}
    >
      {children}
    </SessionProvider>
  );
}

