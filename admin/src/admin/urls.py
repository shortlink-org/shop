"""URL configuration for admin project.

The `urlpatterns` list routes URLs to views. For more information please see:
    https://docs.djangoproject.com/en/5.0/topics/http/urls/

Examples:
Function views
    1. Add an import:  from my_app import views
    2. Add a URL to urlpatterns:  path('', views.home, name='home')
Class-based views
    1. Add an import:  from other_app.views import Home
    2. Add a URL to urlpatterns:  path('', Home.as_view(), name='home')
Including another URLconf
    1. Import the include() function: from django.urls import include, path
    2. Add a URL to urlpatterns:  path('blog/', include('blog.urls'))
"""

from asgiref.sync import async_to_sync
from django.contrib import admin
from django.urls import include, path
from django.views.generic import TemplateView
from drf_spectacular.views import SpectacularAPIView, SpectacularRedocView, SpectacularSwaggerView
from health_check.views import HealthCheckView

from . import admin as user_admin  # noqa: F401 - Register User/Group with Unfold styling
from . import views

# django-health-check 4.x view is async; under WSGI (gunicorn) the async view's coroutines
# are never awaited, causing 500 and "coroutine 'HealthCheck.get_result' was never awaited".
# Wrap with async_to_sync so the health check runs in an event loop and returns a real response.
_health_check_view = HealthCheckView.as_view(
    checks=[
        "health_check.Database",
        "health_check.Cache",
        "health_check.contrib.migrations.Migrations",
    ],
)

urlpatterns = [
    path("__debug__/", include("debug_toolbar.urls")),
    path("admin/logout/", views.logout_view, name="admin_logout"),
    path("admin/orders/", include("domain.orders.urls", namespace="orders")),
    path("admin/", admin.site.urls),
    path("hello/", views.hello, name="hello"),
    path("healthz/", async_to_sync(_health_check_view)),
    path("", TemplateView.as_view(template_name="base.html"), name="home"),
    # REST API:
    path("goods/", include("domain.goods.urls")),
    # API Schema:
    path("api/schema/", SpectacularAPIView.as_view(), name="schema"),
    # Optional UI:
    path("api/schema/swagger-ui/", SpectacularSwaggerView.as_view(url_name="schema"), name="swagger-ui"),
    path("api/schema/redoc/", SpectacularRedocView.as_view(url_name="schema"), name="redoc"),
]
