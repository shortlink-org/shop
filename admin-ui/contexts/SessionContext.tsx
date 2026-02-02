'use client';

/**
 * SessionContext - Centralized session management
 * 
 * Features:
 * - Single source of truth for session
 * - Uses Ory Kratos for authentication
 * - Automatic session refresh
 */

import { createContext, useContext, ReactNode } from 'react';
import { Session } from '@ory/client';

interface SessionContextValue {
  session: Session | null;
  hasSession: boolean;
  isLoading: boolean;
  error: Error | null;
  logout: () => Promise<void>;
}

const SessionContext = createContext<SessionContextValue | undefined>(undefined);

export function useSession() {
  const context = useContext(SessionContext);
  if (context === undefined) {
    throw new Error('useSession must be used within SessionProvider');
  }
  return context;
}

interface SessionProviderProps {
  children: ReactNode;
  session: Session | null;
  isLoading?: boolean;
  error?: Error | null;
  onLogout?: () => Promise<void>;
}

/**
 * SessionProvider - Provides session to all child components
 */
export function SessionProvider({ 
  children, 
  session, 
  isLoading = false,
  error = null,
  onLogout,
}: SessionProviderProps) {
  const logout = async () => {
    if (onLogout) {
      await onLogout();
    }
  };

  const value: SessionContextValue = {
    session,
    hasSession: !!session,
    isLoading,
    error,
    logout,
  };

  return (
    <SessionContext.Provider value={value}>
      {children}
    </SessionContext.Provider>
  );
}

export default SessionContext;
