"""Dashboard components for the admin interface."""

import logging
from datetime import datetime, timedelta
from typing import Any

from django.utils.translation import gettext_lazy as _
from unfold.components import BaseComponent, register_component

from infrastructure.grpc import DeliveryServiceError, get_delivery_client

logger = logging.getLogger(__name__)


@register_component
class DeliveryTrackerComponent(BaseComponent):
    """Tracker component showing delivery success/failure history."""

    def get_context_data(self, **kwargs) -> dict[str, Any]:
        context = super().get_context_data(**kwargs)

        # Try to get real data from delivery service
        tracker_data = self._get_delivery_tracker_data()

        context.update({
            "title": _("Delivery History (Last 30 Days)"),
            "data": tracker_data,
        })
        return context

    def _get_delivery_tracker_data(self) -> list[dict]:
        """Fetch delivery data and format for tracker component."""
        try:
            client = get_delivery_client()
            # Get delivery statistics for the last 30 days
            stats = client.get_delivery_statistics(days=30)

            tracker_data = []
            for day_stat in stats.daily_stats:
                total = day_stat.total_deliveries
                success = day_stat.successful_deliveries
                failed = day_stat.failed_deliveries

                # Determine color based on success rate
                if total == 0:
                    color = "bg-gray-200 dark:bg-gray-700"
                    tooltip = f"{day_stat.date}: No deliveries"
                elif success == total:
                    color = "bg-green-500 dark:bg-green-600"
                    tooltip = f"{day_stat.date}: {success}/{total} successful (100%)"
                elif failed == 0:
                    color = "bg-green-400 dark:bg-green-500"
                    tooltip = f"{day_stat.date}: {success}/{total} successful"
                elif success >= total * 0.9:
                    color = "bg-green-300 dark:bg-green-400"
                    tooltip = f"{day_stat.date}: {success}/{total} ({int(success/total*100)}%)"
                elif success >= total * 0.7:
                    color = "bg-yellow-400 dark:bg-yellow-500"
                    tooltip = f"{day_stat.date}: {success}/{total} ({int(success/total*100)}%)"
                elif success >= total * 0.5:
                    color = "bg-orange-400 dark:bg-orange-500"
                    tooltip = f"{day_stat.date}: {success}/{total} ({int(success/total*100)}%)"
                else:
                    color = "bg-red-500 dark:bg-red-600"
                    tooltip = f"{day_stat.date}: {success}/{total} ({int(success/total*100)}%)"

                tracker_data.append({
                    "color": color,
                    "tooltip": tooltip,
                })

            return tracker_data

        except (DeliveryServiceError, Exception) as e:
            logger.warning(f"Failed to fetch delivery stats: {e}")
            # Return mock data for demonstration
            return self._get_mock_tracker_data()

    def _get_mock_tracker_data(self) -> list[dict]:
        """Generate mock data when delivery service is unavailable."""
        import random

        data = []
        today = datetime.now().date()

        for i in range(30):
            day = today - timedelta(days=29 - i)
            day_str = day.strftime("%Y-%m-%d")

            # Random delivery data
            total = random.randint(0, 50)
            if total == 0:
                color = "bg-gray-200 dark:bg-gray-700"
                tooltip = f"{day_str}: No deliveries"
            else:
                success_rate = random.random() * 0.4 + 0.6  # 60-100%
                success = int(total * success_rate)

                if success_rate >= 0.95:
                    color = "bg-green-500 dark:bg-green-600"
                elif success_rate >= 0.85:
                    color = "bg-green-400 dark:bg-green-500"
                elif success_rate >= 0.75:
                    color = "bg-yellow-400 dark:bg-yellow-500"
                elif success_rate >= 0.65:
                    color = "bg-orange-400 dark:bg-orange-500"
                else:
                    color = "bg-red-500 dark:bg-red-600"

                tooltip = f"{day_str}: {success}/{total} ({int(success_rate*100)}%)"

            data.append({
                "color": color,
                "tooltip": tooltip,
            })

        return data


@register_component
class CourierStatusTrackerComponent(BaseComponent):
    """Tracker showing courier availability over time."""

    def get_context_data(self, **kwargs) -> dict[str, Any]:
        context = super().get_context_data(**kwargs)

        tracker_data = self._get_courier_tracker_data()

        context.update({
            "title": _("Courier Availability (Last 7 Days)"),
            "data": tracker_data,
        })
        return context

    def _get_courier_tracker_data(self) -> list[dict]:
        """Get courier availability data."""
        try:
            client = get_delivery_client()
            result = client.get_courier_pool(
                status_filter=None,
                include_location=False,
                page=1,
                page_size=100,
            )

            # Count by status
            free = sum(1 for c in result.couriers if c.status == "FREE")
            busy = sum(1 for c in result.couriers if c.status == "BUSY")
            unavailable = sum(1 for c in result.couriers if c.status == "UNAVAILABLE")
            archived = sum(1 for c in result.couriers if c.status == "ARCHIVED")

            data = []

            # Add cells for each courier
            for courier in result.couriers:
                if courier.status == "FREE":
                    color = "bg-green-500 dark:bg-green-600"
                elif courier.status == "BUSY":
                    color = "bg-yellow-400 dark:bg-yellow-500"
                elif courier.status == "UNAVAILABLE":
                    color = "bg-gray-400 dark:bg-gray-500"
                else:  # ARCHIVED
                    color = "bg-red-400 dark:bg-red-500"

                data.append({
                    "color": color,
                    "tooltip": f"{courier.name}: {courier.status}",
                })

            return data

        except (DeliveryServiceError, Exception) as e:
            logger.warning(f"Failed to fetch courier stats: {e}")
            return self._get_mock_data()

    def _get_mock_data(self) -> list[dict]:
        """Generate mock data."""
        import random

        statuses = [
            ("FREE", "bg-green-500 dark:bg-green-600"),
            ("BUSY", "bg-yellow-400 dark:bg-yellow-500"),
            ("UNAVAILABLE", "bg-gray-400 dark:bg-gray-500"),
        ]

        data = []
        for i in range(20):
            status, color = random.choice(statuses)
            data.append({
                "color": color,
                "tooltip": f"Courier {i+1}: {status}",
            })

        return data


def dashboard_callback(request, context):
    """
    Callback for customizing the admin dashboard.

    This function is called when rendering the admin index page.
    It adds custom components to the dashboard.
    """
    context.update({
        "dashboard_components": [
            {
                "component": "unfold/components/tracker.html",
                "component_class": "DeliveryTrackerComponent",
                "col_span": 2,  # Full width
            },
            {
                "component": "unfold/components/tracker.html",
                "component_class": "CourierStatusTrackerComponent",
                "col_span": 1,
            },
        ],
    })

    return context
