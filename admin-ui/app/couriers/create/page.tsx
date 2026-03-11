'use client';

import { useCreate } from '@refinedev/core';
import { useRouter } from 'next/navigation';
import { 
  Card, 
  Form, 
  Input, 
  Select, 
  InputNumber, 
  Button, 
  TimePicker, 
  Checkbox,
  message,
  Space,
} from 'antd';
import { ArrowLeftOutlined, SaveOutlined } from '@ant-design/icons';
import Link from 'next/link';
import dayjs from 'dayjs';

import type { TransportType } from '@/types/courier';

const { Option } = Select;

interface FormValues {
  name: string;
  phone: string;
  email: string;
  transportType: TransportType;
  maxDistanceKm: number;
  workZone: string;
  workStart: dayjs.Dayjs;
  workEnd: dayjs.Dayjs;
  workDays: number[];
}

const WEEKDAYS = [
  { value: 0, label: 'Sunday' },
  { value: 1, label: 'Monday' },
  { value: 2, label: 'Tuesday' },
  { value: 3, label: 'Wednesday' },
  { value: 4, label: 'Thursday' },
  { value: 5, label: 'Friday' },
  { value: 6, label: 'Saturday' },
];

export default function CreateCourierPage() {
  const router = useRouter();
  const [form] = Form.useForm<FormValues>();

  const { mutate: createCourier, mutation } = useCreate();
  const isLoading = mutation.isPending;

  const onFinish = (values: FormValues) => {
    createCourier(
      {
        resource: 'couriers',
        values: {
          name: values.name,
          phone: values.phone,
          email: values.email,
          transportType: values.transportType,
          maxDistanceKm: values.maxDistanceKm,
          workZone: values.workZone,
          workHours: {
            startTime: values.workStart.format('HH:mm'),
            endTime: values.workEnd.format('HH:mm'),
            workDays: values.workDays,
          },
        },
      },
      {
        onSuccess: (data) => {
          message.success('Courier registered');
          router.push(`/couriers/${data.data.id}`);
        },
        onError: (error) => {
          message.error(`Error: ${error.message}`);
        },
      }
    );
  };

  return (
    <div className="p-6">
      <div className="mb-4">
        <Space>
          <Link href="/couriers">
            <Button icon={<ArrowLeftOutlined />}>Back</Button>
          </Link>
          <h1 className="text-2xl font-bold m-0">Courier registration</h1>
        </Space>
      </div>

      <Card>
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          initialValues={{
            transportType: 'BICYCLE',
            maxDistanceKm: 10,
            workDays: [1, 2, 3, 4, 5],
            workStart: dayjs('09:00', 'HH:mm'),
            workEnd: dayjs('18:00', 'HH:mm'),
          }}
        >
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Personal Info */}
            <div>
              <h3 className="text-lg font-semibold mb-4">Personal information</h3>
              
              <Form.Item
                name="name"
                label="Full name"
                rules={[{ required: true, message: 'Enter the courier name' }]}
              >
                <Input placeholder="John Doe" />
              </Form.Item>

              <Form.Item
                name="phone"
                label="Phone"
                rules={[
                  { required: true, message: 'Enter a phone number' },
                  { pattern: /^\+?[0-9]{10,15}$/, message: 'Invalid phone number format' },
                ]}
              >
                <Input placeholder="+79001234567" />
              </Form.Item>

              <Form.Item
                name="email"
                label="Email"
                rules={[
                  { required: true, message: 'Enter an email address' },
                  { type: 'email', message: 'Invalid email format' },
                ]}
              >
                <Input placeholder="courier@example.com" />
              </Form.Item>
            </div>

            {/* Work Info */}
            <div>
              <h3 className="text-lg font-semibold mb-4">Work information</h3>
              
              <Form.Item
                name="transportType"
                label="Transport type"
                rules={[{ required: true, message: 'Select a transport type' }]}
              >
                <Select>
                  <Option value="WALKING">Walking</Option>
                  <Option value="BICYCLE">Bicycle</Option>
                  <Option value="MOTORCYCLE">Motorcycle</Option>
                  <Option value="CAR">Car</Option>
                </Select>
              </Form.Item>

              <Form.Item
                name="maxDistanceKm"
                label="Maximum distance (km)"
                rules={[{ required: true, message: 'Enter the maximum distance' }]}
              >
                <InputNumber min={1} max={100} style={{ width: '100%' }} />
              </Form.Item>

              <Form.Item
                name="workZone"
                label="Work zone"
                rules={[{ required: true, message: 'Enter the work zone' }]}
              >
                <Input placeholder="Center, North, South..." />
              </Form.Item>
            </div>

            {/* Schedule */}
            <div className="md:col-span-2">
              <h3 className="text-lg font-semibold mb-4">Schedule</h3>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Form.Item
                  name="workStart"
                  label="Start time"
                  rules={[{ required: true, message: 'Select the start time' }]}
                >
                  <TimePicker format="HH:mm" style={{ width: '100%' }} />
                </Form.Item>

                <Form.Item
                  name="workEnd"
                  label="End time"
                  rules={[{ required: true, message: 'Select the end time' }]}
                >
                  <TimePicker format="HH:mm" style={{ width: '100%' }} />
                </Form.Item>
              </div>

              <Form.Item
                name="workDays"
                label="Work days"
                rules={[{ required: true, message: 'Select the work days' }]}
              >
                <Checkbox.Group options={WEEKDAYS} />
              </Form.Item>
            </div>
          </div>

          <div className="flex justify-end gap-4 mt-6">
            <Link href="/couriers">
              <Button>Cancel</Button>
            </Link>
            <Button 
              type="primary" 
              htmlType="submit" 
              icon={<SaveOutlined />}
              loading={isLoading}
            >
              Register
            </Button>
          </div>
        </Form>
      </Card>
    </div>
  );
}
