"""Custom RemoteUserMiddleware for Oathkeeper X-User-ID header.

Django's RemoteUserMiddleware defaults to REMOTE_USER header.
This middleware reads from X-User-ID header set by Oathkeeper.

See ADR: docs/ADR/decisions/0005-oathkeeper-auth.md
"""

from django.contrib.auth.middleware import RemoteUserMiddleware


class OathkeeperRemoteUserMiddleware(RemoteUserMiddleware):
    """RemoteUserMiddleware that reads X-User-ID header from Oathkeeper.
    
    Oathkeeper validates the Kratos session and injects X-User-ID header.
    Django converts HTTP headers to META keys: X-User-ID -> HTTP_X_USER_ID
    """
    
    header = "HTTP_X_USER_ID"
