import React, { useState, useEffect, forwardRef, useImperativeHandle, useRef } from 'react';
import { Table, message, Input, Form, Row, Col } from 'antd';
import { useTranslation } from 'react-i18next';
import { listMcpConsumers } from '@/services/mcp';
import { useSearchParams } from 'ice';

const ConsumerTable = forwardRef<any>((_, ref) => {
  const { t } = useTranslation();
  const [consumers, setConsumers] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchParams] = useSearchParams();
  const name = searchParams.get('name');

  const [form] = Form.useForm();
  const debounceRef = useRef<any | null>(null);

  const fetchConsumers = async (consumerName?: string) => {
    setLoading(true);
    try {
      const res = await listMcpConsumers({
        mcpServerName: name,
        consumerName,
      });
      setConsumers(res || []);
    } catch (error) {
      message.error(t('mcp.detail.fetchConsumersError'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchConsumers('');
  }, []);

  const columns = [
    {
      title: t('mcp.detail.consumerName'),
      dataIndex: 'consumerName',
      key: 'consumerName',
    },
  ];

  useImperativeHandle(ref, () => ({
    fetchConsumers,
  }));

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { value } = e.target;
    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }
    debounceRef.current = setTimeout(() => {
      fetchConsumers(value);
    }, 800);
  };

  return (
    <div>
      <Form
        form={form}
      >
        <Row gutter={24}>
          <Col span={12}>
            <Form.Item name="consumerName" label={t('mcp.detail.consumerName')}>
              <Input
                allowClear
                placeholder={t('mcp.detail.consumerNameSearchPlaceholder') as string}
                onChange={handleInputChange}
              />
            </Form.Item>
          </Col>
        </Row>
      </Form>
      <Table
        columns={columns}
        dataSource={consumers.map((consumer, index) => ({ ...consumer, key: consumer.consumerName || index }))}
        loading={loading}
        rowKey="key"
        pagination={false}
        locale={{ emptyText: t('mcp.detail.noData') }}
      />
    </div>
  );
});

export default ConsumerTable;
