"""URL patterns for order management."""

from django.urls import path

from . import views

app_name = "orders"

urlpatterns = [
    # List and detail views
    path("", views.order_list, name="list"),
    path("<str:order_id>/", views.order_detail, name="detail"),
    # Actions
    path("<str:order_id>/cancel/", views.order_cancel, name="cancel"),
]
