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
  { value: 0, label: 'Воскресенье' },
  { value: 1, label: 'Понедельник' },
  { value: 2, label: 'Вторник' },
  { value: 3, label: 'Среда' },
  { value: 4, label: 'Четверг' },
  { value: 5, label: 'Пятница' },
  { value: 6, label: 'Суббота' },
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
          message.success('Курьер зарегистрирован');
          router.push(`/couriers/${data.data.id}`);
        },
        onError: (error) => {
          message.error(`Ошибка: ${error.message}`);
        },
      }
    );
  };

  return (
    <div className="p-6">
      <div className="mb-4">
        <Space>
          <Link href="/couriers">
            <Button icon={<ArrowLeftOutlined />}>Назад</Button>
          </Link>
          <h1 className="text-2xl font-bold m-0">Регистрация курьера</h1>
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
              <h3 className="text-lg font-semibold mb-4">Личная информация</h3>
              
              <Form.Item
                name="name"
                label="ФИО"
                rules={[{ required: true, message: 'Введите имя курьера' }]}
              >
                <Input placeholder="Иван Иванов" />
              </Form.Item>

              <Form.Item
                name="phone"
                label="Телефон"
                rules={[
                  { required: true, message: 'Введите телефон' },
                  { pattern: /^\+?[0-9]{10,15}$/, message: 'Неверный формат телефона' },
                ]}
              >
                <Input placeholder="+79001234567" />
              </Form.Item>

              <Form.Item
                name="email"
                label="Email"
                rules={[
                  { required: true, message: 'Введите email' },
                  { type: 'email', message: 'Неверный формат email' },
                ]}
              >
                <Input placeholder="courier@example.com" />
              </Form.Item>
            </div>

            {/* Work Info */}
            <div>
              <h3 className="text-lg font-semibold mb-4">Рабочая информация</h3>
              
              <Form.Item
                name="transportType"
                label="Тип транспорта"
                rules={[{ required: true, message: 'Выберите тип транспорта' }]}
              >
                <Select>
                  <Option value="WALKING">Пешком</Option>
                  <Option value="BICYCLE">Велосипед</Option>
                  <Option value="MOTORCYCLE">Мотоцикл</Option>
                  <Option value="CAR">Автомобиль</Option>
                </Select>
              </Form.Item>

              <Form.Item
                name="maxDistanceKm"
                label="Максимальная дистанция (км)"
                rules={[{ required: true, message: 'Введите максимальную дистанцию' }]}
              >
                <InputNumber min={1} max={100} style={{ width: '100%' }} />
              </Form.Item>

              <Form.Item
                name="workZone"
                label="Рабочая зона"
                rules={[{ required: true, message: 'Введите рабочую зону' }]}
              >
                <Input placeholder="Центр, Север, Юг..." />
              </Form.Item>
            </div>

            {/* Schedule */}
            <div className="md:col-span-2">
              <h3 className="text-lg font-semibold mb-4">Расписание</h3>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Form.Item
                  name="workStart"
                  label="Начало работы"
                  rules={[{ required: true, message: 'Выберите время начала' }]}
                >
                  <TimePicker format="HH:mm" style={{ width: '100%' }} />
                </Form.Item>

                <Form.Item
                  name="workEnd"
                  label="Конец работы"
                  rules={[{ required: true, message: 'Выберите время окончания' }]}
                >
                  <TimePicker format="HH:mm" style={{ width: '100%' }} />
                </Form.Item>
              </div>

              <Form.Item
                name="workDays"
                label="Рабочие дни"
                rules={[{ required: true, message: 'Выберите рабочие дни' }]}
              >
                <Checkbox.Group options={WEEKDAYS} />
              </Form.Item>
            </div>
          </div>

          <div className="flex justify-end gap-4 mt-6">
            <Link href="/couriers">
              <Button>Отмена</Button>
            </Link>
            <Button 
              type="primary" 
              htmlType="submit" 
              icon={<SaveOutlined />}
              loading={isLoading}
            >
              Зарегистрировать
            </Button>
          </div>
        </Form>
      </Card>
    </div>
  );
}
