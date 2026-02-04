"""Dashboard components for the admin interface.

Courier and delivery management has been moved to admin-ui.
See: admin-ui/ROADMAP.md
"""


def dashboard_callback(request, context):
    """Callback for customizing the admin dashboard.

    This function is called when rendering the admin index page.
    """
    # Courier and delivery components moved to admin-ui
    # This dashboard now focuses on goods and orders management
    return context
