'use client';

import { useState } from 'react';
import LoadingDots from 'components/loading-dots';

export interface DeliveryAddress {
  street: string;
  city: string;
  postalCode: string;
  country: string;
  latitude?: number;
  longitude?: number;
}

export interface DeliveryPeriod {
  startTime: string;
  endTime: string;
}

export interface CheckoutFormData {
  deliveryAddress: DeliveryAddress;
  deliveryPeriod: DeliveryPeriod;
  priority: 'NORMAL' | 'URGENT';
}

interface CheckoutFormProps {
  onSubmit: (data: CheckoutFormData) => Promise<void>;
  isLoading?: boolean;
}

const TIME_SLOTS = [
  { label: '09:00 - 12:00', start: '09:00', end: '12:00' },
  { label: '12:00 - 15:00', start: '12:00', end: '15:00' },
  { label: '15:00 - 18:00', start: '15:00', end: '18:00' },
  { label: '18:00 - 21:00', start: '18:00', end: '21:00' }
];

export default function CheckoutForm({ onSubmit, isLoading = false }: CheckoutFormProps) {
  const [formData, setFormData] = useState<CheckoutFormData>({
    deliveryAddress: {
      street: '',
      city: '',
      postalCode: '',
      country: 'Germany'
    },
    deliveryPeriod: {
      startTime: '',
      endTime: ''
    },
    priority: 'NORMAL'
  });

  const [selectedDate, setSelectedDate] = useState('');
  const [selectedTimeSlot, setSelectedTimeSlot] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.deliveryAddress.street.trim()) {
      newErrors.street = 'Street address is required';
    }
    if (!formData.deliveryAddress.city.trim()) {
      newErrors.city = 'City is required';
    }
    if (!formData.deliveryAddress.country.trim()) {
      newErrors.country = 'Country is required';
    }
    if (!selectedDate) {
      newErrors.date = 'Delivery date is required';
    }
    if (!selectedTimeSlot) {
      newErrors.timeSlot = 'Delivery time slot is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    const timeSlot = TIME_SLOTS.find((slot) => slot.label === selectedTimeSlot);
    if (!timeSlot) return;

    const startTime = new Date(`${selectedDate}T${timeSlot.start}:00`).toISOString();
    const endTime = new Date(`${selectedDate}T${timeSlot.end}:00`).toISOString();

    const submitData: CheckoutFormData = {
      ...formData,
      deliveryPeriod: {
        startTime,
        endTime
      }
    };

    await onSubmit(submitData);
  };

  const updateAddress = (field: keyof DeliveryAddress, value: string) => {
    setFormData((prev) => ({
      ...prev,
      deliveryAddress: {
        ...prev.deliveryAddress,
        [field]: value
      }
    }));
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: '' }));
    }
  };

  // Get minimum date (tomorrow)
  const getMinDate = () => {
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    return tomorrow.toISOString().split('T')[0];
  };

  // Get maximum date (2 weeks from now)
  const getMaxDate = () => {
    const maxDate = new Date();
    maxDate.setDate(maxDate.getDate() + 14);
    return maxDate.toISOString().split('T')[0];
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Delivery Address Section */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold text-black dark:text-white">Delivery Address</h3>

        <div>
          <label
            htmlFor="street"
            className="block text-sm font-medium text-neutral-700 dark:text-neutral-300"
          >
            Street Address *
          </label>
          <input
            type="text"
            id="street"
            value={formData.deliveryAddress.street}
            onChange={(e) => updateAddress('street', e.target.value)}
            className={`mt-1 block w-full rounded-md border px-3 py-2 shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-neutral-800 dark:text-white ${
              errors.street
                ? 'border-red-500'
                : 'border-neutral-300 dark:border-neutral-600'
            }`}
            placeholder="123 Main Street, Apt 4"
          />
          {errors.street && <p className="mt-1 text-sm text-red-500">{errors.street}</p>}
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label
              htmlFor="city"
              className="block text-sm font-medium text-neutral-700 dark:text-neutral-300"
            >
              City *
            </label>
            <input
              type="text"
              id="city"
              value={formData.deliveryAddress.city}
              onChange={(e) => updateAddress('city', e.target.value)}
              className={`mt-1 block w-full rounded-md border px-3 py-2 shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-neutral-800 dark:text-white ${
                errors.city
                  ? 'border-red-500'
                  : 'border-neutral-300 dark:border-neutral-600'
              }`}
              placeholder="Berlin"
            />
            {errors.city && <p className="mt-1 text-sm text-red-500">{errors.city}</p>}
          </div>

          <div>
            <label
              htmlFor="postalCode"
              className="block text-sm font-medium text-neutral-700 dark:text-neutral-300"
            >
              Postal Code
            </label>
            <input
              type="text"
              id="postalCode"
              value={formData.deliveryAddress.postalCode}
              onChange={(e) => updateAddress('postalCode', e.target.value)}
              className="mt-1 block w-full rounded-md border border-neutral-300 px-3 py-2 shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-neutral-600 dark:bg-neutral-800 dark:text-white"
              placeholder="10115"
            />
          </div>
        </div>

        <div>
          <label
            htmlFor="country"
            className="block text-sm font-medium text-neutral-700 dark:text-neutral-300"
          >
            Country *
          </label>
          <select
            id="country"
            value={formData.deliveryAddress.country}
            onChange={(e) => updateAddress('country', e.target.value)}
            className={`mt-1 block w-full rounded-md border px-3 py-2 shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-neutral-800 dark:text-white ${
              errors.country
                ? 'border-red-500'
                : 'border-neutral-300 dark:border-neutral-600'
            }`}
          >
            <option value="Germany">Germany</option>
            <option value="Austria">Austria</option>
            <option value="Switzerland">Switzerland</option>
            <option value="Netherlands">Netherlands</option>
            <option value="Belgium">Belgium</option>
            <option value="France">France</option>
          </select>
          {errors.country && <p className="mt-1 text-sm text-red-500">{errors.country}</p>}
        </div>
      </div>

      {/* Delivery Period Section */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold text-black dark:text-white">Delivery Time</h3>

        <div>
          <label
            htmlFor="deliveryDate"
            className="block text-sm font-medium text-neutral-700 dark:text-neutral-300"
          >
            Delivery Date *
          </label>
          <input
            type="date"
            id="deliveryDate"
            value={selectedDate}
            onChange={(e) => {
              setSelectedDate(e.target.value);
              if (errors.date) setErrors((prev) => ({ ...prev, date: '' }));
            }}
            min={getMinDate()}
            max={getMaxDate()}
            className={`mt-1 block w-full rounded-md border px-3 py-2 shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-neutral-800 dark:text-white ${
              errors.date
                ? 'border-red-500'
                : 'border-neutral-300 dark:border-neutral-600'
            }`}
          />
          {errors.date && <p className="mt-1 text-sm text-red-500">{errors.date}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-neutral-700 dark:text-neutral-300">
            Time Slot *
          </label>
          <div className="mt-2 grid grid-cols-2 gap-2">
            {TIME_SLOTS.map((slot) => (
              <button
                key={slot.label}
                type="button"
                onClick={() => {
                  setSelectedTimeSlot(slot.label);
                  if (errors.timeSlot) setErrors((prev) => ({ ...prev, timeSlot: '' }));
                }}
                className={`rounded-md border px-4 py-2 text-sm font-medium transition-colors ${
                  selectedTimeSlot === slot.label
                    ? 'border-blue-500 bg-blue-500 text-white'
                    : 'border-neutral-300 bg-white text-neutral-700 hover:bg-neutral-50 dark:border-neutral-600 dark:bg-neutral-800 dark:text-neutral-300 dark:hover:bg-neutral-700'
                }`}
              >
                {slot.label}
              </button>
            ))}
          </div>
          {errors.timeSlot && <p className="mt-1 text-sm text-red-500">{errors.timeSlot}</p>}
        </div>
      </div>

      {/* Priority Section */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold text-black dark:text-white">Delivery Priority</h3>
        <div className="flex gap-4">
          <label className="flex cursor-pointer items-center">
            <input
              type="radio"
              name="priority"
              value="NORMAL"
              checked={formData.priority === 'NORMAL'}
              onChange={() => setFormData((prev) => ({ ...prev, priority: 'NORMAL' }))}
              className="h-4 w-4 text-blue-600 focus:ring-blue-500"
            />
            <span className="ml-2 text-sm text-neutral-700 dark:text-neutral-300">
              Normal Delivery
            </span>
          </label>
          <label className="flex cursor-pointer items-center">
            <input
              type="radio"
              name="priority"
              value="URGENT"
              checked={formData.priority === 'URGENT'}
              onChange={() => setFormData((prev) => ({ ...prev, priority: 'URGENT' }))}
              className="h-4 w-4 text-blue-600 focus:ring-blue-500"
            />
            <span className="ml-2 text-sm text-neutral-700 dark:text-neutral-300">
              Urgent Delivery (+$10)
            </span>
          </label>
        </div>
      </div>

      {/* Submit Button */}
      <button
        type="submit"
        disabled={isLoading}
        className="w-full rounded-full bg-blue-600 px-6 py-3 text-center text-sm font-medium text-white opacity-90 hover:opacity-100 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {isLoading ? <LoadingDots className="bg-white" /> : 'Place Order'}
      </button>
    </form>
  );
}
