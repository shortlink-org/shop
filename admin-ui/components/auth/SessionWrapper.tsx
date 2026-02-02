'use client';

/**
 * SessionWrapper - Session provider for admin pages
 * 
 * Fetches session on mount and provides it to children
 * Redirects to login if not authenticated
 */

import React, { ReactNode, useState, useEffect } from 'react';
import { Spin, Result, Button } from 'antd';
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
      <div className="flex items-center justify-center min-h-screen">
        <Spin size="large" tip="Загрузка..." />
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <Result
        status="error"
        title="Ошибка авторизации"
        subTitle={error.message}
        extra={[
          <Button type="primary" key="login" href={getLoginUrl()}>
            Войти
          </Button>,
          <Button key="retry" onClick={() => window.location.reload()}>
            Повторить
          </Button>,
        ]}
      />
    );
  }

  // Not authenticated (and requireAuth is true)
  if (requireAuth && !session) {
    return (
      <Result
        status="403"
        title="Требуется авторизация"
        subTitle="Пожалуйста, войдите в систему для доступа к админ-панели"
        extra={
          <Button type="primary" href={getLoginUrl()}>
            Войти
          </Button>
        }
      />
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
