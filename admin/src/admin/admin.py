"""Override default User and Group admin classes with Unfold styling.

This is required because Django's built-in User and Group admin classes
don't inherit from unfold.admin.ModelAdmin, so they appear unstyled.

See: https://unfoldadmin.com/docs/installation/auth/
"""

from typing import ClassVar

from django.contrib import admin, messages
from django.contrib.auth.admin import GroupAdmin as BaseGroupAdmin
from django.contrib.auth.admin import UserAdmin as BaseUserAdmin
from django.contrib.auth.models import Group, User
from django.http import HttpRequest
from django.shortcuts import redirect
from django.urls import reverse
from django.utils.translation import gettext_lazy as _
from unfold.admin import ModelAdmin
from unfold.contrib.filters.admin import (
    AutocompleteSelectMultipleFilter,
    BooleanRadioFilter,
    RangeDateTimeFilter,
)
from unfold.decorators import action, display
from unfold.forms import AdminPasswordChangeForm, UserChangeForm, UserCreationForm
from unfold.paginator import InfinitePaginator

admin.site.unregister(User)
admin.site.unregister(Group)


@admin.register(User)
class UserAdmin(BaseUserAdmin, ModelAdmin):
    """User admin with Unfold styling."""

    # Forms loaded from `unfold.forms`
    form = UserChangeForm
    add_form = UserCreationForm
    change_password_form = AdminPasswordChangeForm
    paginator = InfinitePaginator
    show_full_result_count = False

    list_display = (
        "display_user_header",
        "display_active_status",
        "display_staff_status",
        "display_superuser_status",
        "display_groups",
        "date_joined",
    )

    # Override fieldsets with tab classes for better organization
    fieldsets = (
        (None, {"fields": ("username", "password")}),
        (
            _("Personal Info"),
            {
                "classes": ["tab"],
                "fields": ("first_name", "last_name", "email"),
            },
        ),
        (
            _("Permissions"),
            {
                "classes": ["tab"],
                "fields": (
                    "is_active",
                    "is_staff",
                    "is_superuser",
                    "groups",
                    "user_permissions",
                ),
            },
        ),
        (
            _("Important Dates"),
            {
                "classes": ["tab"],
                "fields": ("last_login", "date_joined"),
            },
        ),
    )

    list_filter = (
        ("is_staff", BooleanRadioFilter),
        ("is_superuser", BooleanRadioFilter),
        ("is_active", BooleanRadioFilter),
        ["groups", AutocompleteSelectMultipleFilter],
        ("date_joined", RangeDateTimeFilter),
        ("last_login", RangeDateTimeFilter),
    )
    list_filter_submit = True
    actions_row: ClassVar[list[str]] = ["activate_user", "deactivate_user"]

    @action(
        description=_("Activate"),
        permissions=["activate_user"],
        url_path="activate",
    )
    def activate_user(self, request: HttpRequest, object_id: int):
        """Activate the user account."""
        user = User.objects.get(pk=object_id)
        if user.is_active:
            messages.info(request, f"User '{user.username}' is already active.")
        else:
            user.is_active = True
            user.save(update_fields=["is_active"])
            messages.success(request, f"User '{user.username}' activated.")
        return redirect(reverse("admin:auth_user_changelist"))

    def has_activate_user_permission(self, request: HttpRequest) -> bool:
        """Check if user can activate users."""
        return request.user.has_perm("auth.change_user")

    @action(
        description=_("Deactivate"),
        permissions=["deactivate_user"],
        url_path="deactivate",
    )
    def deactivate_user(self, request: HttpRequest, object_id: int):
        """Deactivate the user account."""
        user = User.objects.get(pk=object_id)
        if user == request.user:
            messages.error(request, "You cannot deactivate yourself.")
        elif not user.is_active:
            messages.info(request, f"User '{user.username}' is already inactive.")
        else:
            user.is_active = False
            user.save(update_fields=["is_active"])
            messages.success(request, f"User '{user.username}' deactivated.")
        return redirect(reverse("admin:auth_user_changelist"))

    def has_deactivate_user_permission(self, request: HttpRequest) -> bool:
        """Check if user can deactivate users."""
        return request.user.has_perm("auth.change_user")

    @display(header=True)
    def display_user_header(self, obj):
        """Display username with email as two-line header."""
        # Get initials from name or username
        if obj.first_name and obj.last_name:
            initials = f"{obj.first_name[0]}{obj.last_name[0]}".upper()
            name = obj.get_full_name()
        else:
            initials = obj.username[:2].upper()
            name = obj.username
        return [
            name,
            obj.email or _("No email"),
            initials,
        ]

    @display(
        description=_("Active"),
        ordering="is_active",
        label={
            True: "success",
            False: "danger",
        },
    )
    def display_active_status(self, obj):
        """Display active status with colored label."""
        return obj.is_active, _("Active") if obj.is_active else _("Inactive")

    @display(
        description=_("Staff"),
        ordering="is_staff",
        label={
            True: "info",
            False: "warning",
        },
    )
    def display_staff_status(self, obj):
        """Display staff status with colored label."""
        return obj.is_staff, _("Staff") if obj.is_staff else _("No")

    @display(
        description=_("Superuser"),
        ordering="is_superuser",
        label={
            True: "success",
            False: "warning",
        },
    )
    def display_superuser_status(self, obj):
        """Display superuser status with colored label."""
        return obj.is_superuser, _("Admin") if obj.is_superuser else _("No")

    @display(description=_("Groups"), dropdown=True)
    def display_groups(self, obj):
        """Display user groups as dropdown menu."""
        groups = obj.groups.all()
        if not groups:
            return {"title": "-", "items": []}

        return {
            "title": f"{groups.count()} " + (_("group") if groups.count() == 1 else _("groups")),
            "striped": True,
            "items": [
                {
                    "title": group.name,
                    "link": reverse("admin:auth_group_change", args=[group.pk]),
                }
                for group in groups
            ],
        }


@admin.register(Group)
class GroupAdmin(BaseGroupAdmin, ModelAdmin):
    """Group admin with Unfold styling."""

    paginator = InfinitePaginator
    show_full_result_count = False
    search_fields: ClassVar[list[str]] = ["name"]  # Required for autocomplete filter
