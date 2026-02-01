"""URL patterns for courier management."""

from django.urls import path

from . import views

app_name = "couriers"

urlpatterns = [
    # List and detail views
    path("", views.courier_list, name="list"),
    path("register/", views.courier_register, name="register"),
    path("map/", views.courier_map, name="map"),
    path("<str:courier_id>/", views.courier_detail, name="detail"),

    # Lifecycle actions
    path("<str:courier_id>/activate/", views.courier_activate, name="activate"),
    path("<str:courier_id>/deactivate/", views.courier_deactivate, name="deactivate"),
    path("<str:courier_id>/archive/", views.courier_archive, name="archive"),

    # Profile actions
    path("<str:courier_id>/update-contact/", views.courier_update_contact, name="update_contact"),
    path("<str:courier_id>/update-schedule/", views.courier_update_schedule, name="update_schedule"),
    path("<str:courier_id>/change-transport/", views.courier_change_transport, name="change_transport"),
]
