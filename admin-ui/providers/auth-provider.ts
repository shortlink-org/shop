/**
 * Refine Auth Provider for Ory Kratos integration
 */

import { AuthProvider } from '@refinedev/core';
import { Session } from '@ory/client';

import { RATE_LIMIT_MESSAGE } from '@/lib/constants';
import {
  fetchSessionOptional,
  logout as oryLogout,
  getLoginUrl,
  getUserName,
  getUserEmail,
  isAdmin,
} from '@/lib/ory/api';

export const authProvider: AuthProvider = {
  /**
   * Login - redirect to Ory login page
   */
  login: async () => {
    const loginUrl = getLoginUrl(window.location.href);
    window.location.href = loginUrl;
    
    return {
      success: true,
      redirectTo: loginUrl,
    };
  },

  /**
   * Logout - create logout flow and redirect
   */
  logout: async () => {
    try {
      await oryLogout();
      return {
        success: true,
        redirectTo: '/',
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error : new Error('Logout failed'),
      };
    }
  },

  /**
   * Check if user is authenticated
   */
  check: async () => {
    try {
      const session = await fetchSessionOptional();
      
      if (session) {
        return {
          authenticated: true,
        };
      }

      return {
        authenticated: false,
        redirectTo: getLoginUrl(),
        error: new Error('Not authenticated'),
      };
    } catch (error) {
      return {
        authenticated: false,
        redirectTo: getLoginUrl(),
        error: error instanceof Error ? error : new Error('Auth check failed'),
      };
    }
  },

  /**
   * Handle authentication errors
   */
  onError: async (error) => {
    const status = (error as any)?.response?.status;
    const statusCode = (error as any)?.statusCode;

    if (status === 401 || status === 403) {
      return {
        logout: true,
        redirectTo: getLoginUrl(),
        error,
      };
    }

    if (statusCode === 429 || status === 429) {
      return {
        error: new Error(RATE_LIMIT_MESSAGE),
      };
    }

    return {
      error,
    };
  },

  /**
   * Get current user identity
   */
  getIdentity: async () => {
    try {
      const session = await fetchSessionOptional();
      
      if (!session) {
        return null;
      }

      return {
        id: session.identity?.id,
        name: getUserName(session),
        email: getUserEmail(session),
        avatar: undefined, // Ory doesn't provide avatars by default
      };
    } catch {
      return null;
    }
  },

  /**
   * Get user permissions/roles
   */
  getPermissions: async () => {
    try {
      const session = await fetchSessionOptional();
      
      if (!session) {
        return null;
      }

      const permissions: string[] = [];
      
      if (isAdmin(session)) {
        permissions.push('admin');
      }
      
      // Add default permissions for authenticated users
      permissions.push('courier:read', 'courier:write');
      
      return permissions;
    } catch {
      return null;
    }
  },
};

export default authProvider;
