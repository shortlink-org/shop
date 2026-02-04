"""Ory auth context processor with error handling."""

import hashlib
import logging

import requests
from django.conf import settings
from django.core.cache import cache
from django.urls import reverse

logger = logging.getLogger(__name__)


def _logout_cache_key(request):
    session_cookie = request.COOKIES.get("ory_kratos_session")
    if not session_cookie:
        return None
    digest = hashlib.sha256(session_cookie.encode("utf-8")).hexdigest()
    return f"ory_logout_url:{digest}"


def processor(request):
    """Context processor for Ory auth URLs."""
    # Default to Django logout as fallback
    django_logout_url = reverse("admin:logout") if request else "/admin/logout/"

    context = {
        "login_url": f"{settings.ORY_UI_URL}/login",
        "signup_url": f"{settings.ORY_UI_URL}/registration",
        "logout_url": django_logout_url,
        "recovery_url": f"{settings.ORY_UI_URL}/recovery",
        "verify_url": f"{settings.ORY_UI_URL}/verification",
        "profile_url": f"{settings.ORY_UI_URL}/settings",
    }

    if request.user.is_authenticated:
        cache_key = _logout_cache_key(request)
        if cache_key:
            cached_logout = cache.get(cache_key)
            if cached_logout:
                context["logout_url"] = cached_logout
                return context
        try:
            url = f"{settings.ORY_SDK_URL}/self-service/logout/browser"
            r = requests.get(
                url,
                cookies=request.COOKIES,
                timeout=5,
            )
            logger.info("Kratos logout request: %s -> %s", url, r.status_code)
            if r.status_code == requests.codes.ok:
                ory_logout = r.json().get("logout_url", "")
                if ory_logout:
                    context["logout_url"] = ory_logout
                    if cache_key:
                        cache.set(cache_key, ory_logout, timeout=60)
                    logger.info("Got Kratos logout URL: %s", ory_logout)
        except (requests.RequestException, ValueError) as e:
            logger.warning("Failed to get logout URL from Ory: %s", e)

    return context
