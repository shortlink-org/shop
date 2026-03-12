'use client';

import { Button, FeedbackPanel } from '@shortlink-org/ui-kit';
import { useMutation } from '@apollo/client/react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useMemo, useState } from 'react';
import { toast } from 'sonner';

import { REGISTER_COURIER } from '@/graphql/mutations/couriers';
import type { TransportType } from '@/types/courier';

type RegisterCourierMutationResult = {
  registerCourier?: {
    courierId: string;
    status?: string | null;
    createdAt?: string | null;
  } | null;
};

interface FormValues {
  name: string;
  phone: string;
  email: string;
  transportType: TransportType;
  maxDistanceKm: number;
  workZone: string;
  workStart: string;
  workEnd: string;
  workDays: number[];
}

type FormErrors = Partial<Record<keyof FormValues, string>>;

const WEEKDAYS = [
  { value: 0, label: 'Sunday' },
  { value: 1, label: 'Monday' },
  { value: 2, label: 'Tuesday' },
  { value: 3, label: 'Wednesday' },
  { value: 4, label: 'Thursday' },
  { value: 5, label: 'Friday' },
  { value: 6, label: 'Saturday' },
];

const INITIAL_VALUES: FormValues = {
  name: '',
  phone: '',
  email: '',
  transportType: 'BICYCLE',
  maxDistanceKm: 10,
  workZone: '',
  workStart: '09:00',
  workEnd: '18:00',
  workDays: [1, 2, 3, 4, 5],
};

function validateForm(values: FormValues): FormErrors {
  const errors: FormErrors = {};

  if (!values.name.trim()) {
    errors.name = 'Enter the courier name';
  }

  if (!values.phone.trim()) {
    errors.phone = 'Enter a phone number';
  } else if (!/^\+?[0-9]{10,15}$/.test(values.phone.trim())) {
    errors.phone = 'Invalid phone number format';
  }

  if (!values.email.trim()) {
    errors.email = 'Enter an email address';
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(values.email.trim())) {
    errors.email = 'Invalid email format';
  }

  if (!values.transportType) {
    errors.transportType = 'Select a transport type';
  }

  if (!Number.isFinite(values.maxDistanceKm) || values.maxDistanceKm < 1 || values.maxDistanceKm > 100) {
    errors.maxDistanceKm = 'Distance must be between 1 and 100 km';
  }

  if (!values.workZone.trim()) {
    errors.workZone = 'Enter the work zone';
  }

  if (!values.workStart) {
    errors.workStart = 'Select the start time';
  }

  if (!values.workEnd) {
    errors.workEnd = 'Select the end time';
  }

  if (values.workStart && values.workEnd && values.workStart >= values.workEnd) {
    errors.workEnd = 'End time must be later than start time';
  }

  if (!values.workDays.length) {
    errors.workDays = 'Select at least one work day';
  }

  return errors;
}

export default function CreateCourierPage() {
  const router = useRouter();
  const [values, setValues] = useState<FormValues>(INITIAL_VALUES);
  const [errors, setErrors] = useState<FormErrors>({});
  const [registerCourier, { loading }] = useMutation<RegisterCourierMutationResult>(REGISTER_COURIER);

  const selectedWorkDays = useMemo(() => new Set(values.workDays), [values.workDays]);

  const updateField = <K extends keyof FormValues>(field: K, value: FormValues[K]) => {
    setValues((current) => ({ ...current, [field]: value }));
    setErrors((current) => ({ ...current, [field]: undefined }));
  };

  const toggleWorkDay = (day: number) => {
    const nextDays = selectedWorkDays.has(day)
      ? values.workDays.filter((value) => value !== day)
      : [...values.workDays, day].sort((left, right) => left - right);
    updateField('workDays', nextDays);
  };

  const onFinish = async () => {
    const validationErrors = validateForm(values);
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      toast.error('Please fix the highlighted fields');
      return;
    }

    try {
      const { data } = await registerCourier({
        variables: {
          input: {
            name: values.name.trim(),
            phone: values.phone.trim(),
            email: values.email.trim(),
            transportType: values.transportType,
            maxDistanceKm: values.maxDistanceKm,
            workZone: values.workZone.trim(),
            workHours: {
              startTime: values.workStart,
              endTime: values.workEnd,
              workDays: values.workDays,
            },
          },
        },
      });

      const courierId = data?.registerCourier?.courierId;
      if (!courierId) {
        throw new Error('Courier registration did not return an id');
      }

      toast.success('Courier registered');
      router.push(`/couriers/${courierId}`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to register courier');
    }
  };

  return (
    <div className="space-y-6">
      <section className="admin-card overflow-hidden p-6 sm:p-8">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-3">
            <Button as={Link} asProps={{ href: '/couriers' }} variant="secondary" size="sm">
              Back to couriers
            </Button>
            <div>
              <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-[var(--color-muted-foreground)]">
                Courier onboarding
              </p>
              <h1 className="mt-2 text-3xl font-semibold tracking-tight">Register courier</h1>
              <p className="mt-2 max-w-2xl text-sm text-[var(--color-muted-foreground)]">
                Create a new courier profile with contact details, transport configuration, and work schedule.
              </p>
            </div>
          </div>

          <div className="admin-card p-4">
            <p className="text-sm font-semibold">Operational note</p>
            <p className="mt-2 text-sm text-[var(--color-muted-foreground)]">
              New couriers start as unavailable until operations activates them.
            </p>
          </div>
        </div>
      </section>

      <section className="admin-card p-6">
        <form
          onSubmit={(event) => {
            event.preventDefault();
            void onFinish();
          }}
        >
          <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
            <div className="space-y-5">
              <div>
                <h2 className="text-lg font-semibold tracking-tight">Personal information</h2>
                <p className="mt-1 text-sm text-[var(--color-muted-foreground)]">
                  Contact fields used by dispatchers and internal operations.
                </p>
              </div>

              <label className="admin-field">
                <span className="admin-label">Full name</span>
                <input
                  className="admin-input"
                  placeholder="John Doe"
                  value={values.name}
                  onChange={(event) => updateField('name', event.target.value)}
                />
                {errors.name && <span className="admin-error">{errors.name}</span>}
              </label>

              <label className="admin-field">
                <span className="admin-label">Phone</span>
                <input
                  className="admin-input"
                  placeholder="+79001234567"
                  value={values.phone}
                  onChange={(event) => updateField('phone', event.target.value)}
                />
                {errors.phone && <span className="admin-error">{errors.phone}</span>}
              </label>

              <label className="admin-field">
                <span className="admin-label">Email</span>
                <input
                  className="admin-input"
                  type="email"
                  placeholder="courier@example.com"
                  value={values.email}
                  onChange={(event) => updateField('email', event.target.value)}
                />
                {errors.email && <span className="admin-error">{errors.email}</span>}
              </label>
            </div>

            <div className="space-y-5">
              <div>
                <h2 className="text-lg font-semibold tracking-tight">Work configuration</h2>
                <p className="mt-1 text-sm text-[var(--color-muted-foreground)]">
                  Delivery capacity and zone settings used for assignment decisions.
                </p>
              </div>

              <label className="admin-field">
                <span className="admin-label">Transport type</span>
                <select
                  className="admin-select"
                  value={values.transportType}
                  onChange={(event) => updateField('transportType', event.target.value as TransportType)}
                >
                  <option value="WALKING">Walking</option>
                  <option value="BICYCLE">Bicycle</option>
                  <option value="MOTORCYCLE">Motorcycle</option>
                  <option value="CAR">Car</option>
                </select>
                {errors.transportType && <span className="admin-error">{errors.transportType}</span>}
              </label>

              <label className="admin-field">
                <span className="admin-label">Maximum distance (km)</span>
                <input
                  className="admin-input"
                  type="number"
                  min={1}
                  max={100}
                  value={values.maxDistanceKm}
                  onChange={(event) => updateField('maxDistanceKm', Number(event.target.value))}
                />
                {errors.maxDistanceKm && <span className="admin-error">{errors.maxDistanceKm}</span>}
              </label>

              <label className="admin-field">
                <span className="admin-label">Work zone</span>
                <input
                  className="admin-input"
                  placeholder="Center, North, South..."
                  value={values.workZone}
                  onChange={(event) => updateField('workZone', event.target.value)}
                />
                {errors.workZone && <span className="admin-error">{errors.workZone}</span>}
              </label>
            </div>
          </div>

          <div className="mt-8 rounded-[1.5rem] border border-[var(--color-border)] bg-[color-mix(in_srgb,var(--color-surface)_75%,transparent)] p-6">
            <div className="mb-5">
              <h2 className="text-lg font-semibold tracking-tight">Schedule</h2>
              <p className="mt-1 text-sm text-[var(--color-muted-foreground)]">
                Working hours and active weekdays for courier availability planning.
              </p>
            </div>

            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <label className="admin-field">
                <span className="admin-label">Start time</span>
                <input
                  className="admin-input"
                  type="time"
                  value={values.workStart}
                  onChange={(event) => updateField('workStart', event.target.value)}
                />
                {errors.workStart && <span className="admin-error">{errors.workStart}</span>}
              </label>

              <label className="admin-field">
                <span className="admin-label">End time</span>
                <input
                  className="admin-input"
                  type="time"
                  value={values.workEnd}
                  onChange={(event) => updateField('workEnd', event.target.value)}
                />
                {errors.workEnd && <span className="admin-error">{errors.workEnd}</span>}
              </label>
            </div>

            <div className="admin-field mt-4">
              <span className="admin-label">Work days</span>
              <div className="admin-checkbox-grid">
                {WEEKDAYS.map((day) => (
                  <label key={day.value} className="admin-checkbox">
                    <input
                      type="checkbox"
                      checked={selectedWorkDays.has(day.value)}
                      onChange={() => toggleWorkDay(day.value)}
                    />
                    <span>{day.label}</span>
                  </label>
                ))}
              </div>
              {errors.workDays && <span className="admin-error">{errors.workDays}</span>}
            </div>
          </div>

          <div className="mt-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <FeedbackPanel
              variant="empty"
              eyebrow="Validation"
              title="Review data before creating"
              message="The courier registration flow now uses plain React state and Apollo, with no Ant Design form layer."
              size="sm"
              className="w-full max-w-xl"
            />

            <div className="flex flex-wrap gap-3">
              <Button as={Link} asProps={{ href: '/couriers' }} variant="secondary">
                Cancel
              </Button>
              <Button loading={loading} type="submit">
                Register
              </Button>
            </div>
          </div>
        </form>
      </section>
    </div>
  );
}
