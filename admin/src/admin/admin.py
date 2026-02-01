"""
Override default User and Group admin classes with Unfold styling.

This is required because Django's built-in User and Group admin classes
don't inherit from unfold.admin.ModelAdmin, so they appear unstyled.

See: https://unfoldadmin.com/docs/installation/auth/
"""

from django.contrib import admin
from django.contrib.auth.admin import GroupAdmin as BaseGroupAdmin
from django.contrib.auth.admin import UserAdmin as BaseUserAdmin
from django.contrib.auth.models import Group, User

from unfold.admin import ModelAdmin
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


@admin.register(Group)
class GroupAdmin(BaseGroupAdmin, ModelAdmin):
    """Group admin with Unfold styling."""

    paginator = InfinitePaginator
    show_full_result_count = False
