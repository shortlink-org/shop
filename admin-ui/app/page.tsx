'use client';

import { Card, Col, Row, Statistic } from 'antd';
import { TeamOutlined, CheckCircleOutlined, ClockCircleOutlined, CarOutlined } from '@ant-design/icons';

export default function DashboardPage() {
  // TODO: Fetch real data from GraphQL
  const stats = {
    totalCouriers: 45,
    activeCouriers: 32,
    busyCouriers: 18,
    deliveriesToday: 156,
  };

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6">Dashboard</h1>
      
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total couriers"
              value={stats.totalCouriers}
              prefix={<TeamOutlined />}
            />
          </Card>
        </Col>
        
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Active"
              value={stats.activeCouriers}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="On delivery"
              value={stats.busyCouriers}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Deliveries today"
              value={stats.deliveriesToday}
              prefix={<CarOutlined />}
            />
          </Card>
        </Col>
      </Row>
      
      <div className="mt-8">
        <Card title="Quick actions">
          <p className="text-gray-500">
            Go to the Couriers section to manage couriers.
          </p>
        </Card>
      </div>
    </div>
  );
}
