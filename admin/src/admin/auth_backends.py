"""Custom authentication backend for Oathkeeper header-based auth.

This backend trusts the X-User-ID header injected by Oathkeeper after validating
the Kratos session cookie. It replaces direct Kratos SDK calls with header-based
identity propagation.

See ADR: docs/ADR/decisions/0005-oathkeeper-auth.md
"""

from django.contrib.auth import get_user_model
from django.contrib.auth.backends import RemoteUserBackend

User = get_user_model()


class OryRemoteUserBackend(RemoteUserBackend):
    """Authentication backend that trusts X-User-ID header from Oathkeeper.

    Oathkeeper validates the ory_kratos_session cookie and injects identity.id
    into the X-User-ID header. This backend uses that header to authenticate
    users in Django.

    Attributes:
        create_unknown_user: If True, creates a new Django user when encountering
            an unknown identity.id from Oathkeeper.
    """

    create_unknown_user = True

    def configure_user(self, request, user, created=True):
        """Configure a newly created or existing user.

        All users authenticated via Oathkeeper are trusted admins:
        - Sets is_staff=True to allow Django admin access
        - Sets is_superuser=True to grant full permissions
        - Updates email from X-Email header if available

        Args:
            request: The HTTP request object.
            user: The User instance being configured.
            created: Whether the user was just created (unused, we update all users).

        Returns:
            The configured User instance.
        """
        update_fields = []

        # All users authenticated via Oathkeeper are trusted admins
        # Grant staff and superuser access for Django admin
        if not user.is_staff:
            user.is_staff = True
            update_fields.append("is_staff")
        if not user.is_superuser:
            user.is_superuser = True
            update_fields.append("is_superuser")

        # Update email from Oathkeeper header
        email = request.META.get("HTTP_X_EMAIL")
        if email and user.email != email:
            user.email = email
            update_fields.append("email")

        if update_fields:
            user.save(update_fields=update_fields)

        return user
