import {
  AiQuotaConsumerQuota,
  AiQuotaRouteSummary,
  AiQuotaScheduleAction,
  AiQuotaScheduleRule,
  AiQuotaScheduleRuleRequest,
} from '@/interfaces/ai-quota';
import {
  deleteAiQuotaScheduleRule,
  deltaAiQuota,
  getAiQuotaConsumers,
  getAiQuotaRoutes,
  getAiQuotaScheduleRules,
  refreshAiQuota,
  saveAiQuotaScheduleRule,
} from '@/services/ai-quota';
import { RedoOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-layout';
import { useRequest } from 'ahooks';
import {
  Button,
  Descriptions,
  Drawer,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  Typography,
  message,
} from 'antd';
import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';

const { Text } = Typography;

type QuotaModalType = 'refresh' | 'delta' | null;

const AiQuotaPage: React.FC = () => {
  const { t } = useTranslation();
  const [routes, setRoutes] = useState<AiQuotaRouteSummary[]>([]);
  const [selectedRouteName, setSelectedRouteName] = useState<string>();
  const [quotaList, setQuotaList] = useState<AiQuotaConsumerQuota[]>([]);
  const [keyword, setKeyword] = useState('');
  const [quotaModalType, setQuotaModalType] = useState<QuotaModalType>(null);
  const [currentConsumer, setCurrentConsumer] = useState<AiQuotaConsumerQuota | null>(null);
  const [scheduleDrawerOpen, setScheduleDrawerOpen] = useState(false);
  const [scheduleRules, setScheduleRules] = useState<AiQuotaScheduleRule[]>([]);
  const [scheduleConsumer, setScheduleConsumer] = useState<AiQuotaConsumerQuota | null>(null);
  const [editingScheduleRule, setEditingScheduleRule] = useState<AiQuotaScheduleRule | null>(null);

  const [quotaForm] = Form.useForm();
  const [scheduleForm] = Form.useForm();

  const selectedRoute = useMemo(
    () => routes.find((route) => route.routeName === selectedRouteName),
    [routes, selectedRouteName],
  );

  const { loading: routesLoading, run: loadRoutes } = useRequest(getAiQuotaRoutes, {
    manual: true,
    onSuccess: (result = []) => {
      setRoutes(result);
      if (!result.length) {
        setSelectedRouteName(undefined);
        setQuotaList([]);
        return;
      }
      if (!selectedRouteName || !result.some((route) => route.routeName === selectedRouteName)) {
        setSelectedRouteName(result[0].routeName);
      }
    },
  });

  const { loading: quotaLoading, run: loadConsumers } = useRequest(getAiQuotaConsumers, {
    manual: true,
    onSuccess: (result = []) => {
      setQuotaList(result);
    },
  });

  const { loading: scheduleLoading, run: loadSchedules } = useRequest(getAiQuotaScheduleRules, {
    manual: true,
    onSuccess: (result = []) => {
      setScheduleRules(result);
    },
  });

  useEffect(() => {
    loadRoutes();
  }, []);

  useEffect(() => {
    if (selectedRouteName) {
      loadConsumers(selectedRouteName);
    }
  }, [selectedRouteName]);

  const filteredQuotaList = useMemo(() => {
    return quotaList.filter((item) => {
      if (!keyword) {
        return true;
      }
      return item.consumerName.toLowerCase().includes(keyword.toLowerCase());
    });
  }, [keyword, quotaList]);

  const refreshAll = async () => {
    await loadRoutes();
    if (selectedRouteName) {
      await loadConsumers(selectedRouteName);
    }
  };

  const openQuotaModal = (type: QuotaModalType, consumer: AiQuotaConsumerQuota) => {
    setQuotaModalType(type);
    setCurrentConsumer(consumer);
    quotaForm.setFieldsValue({
      value: type === 'refresh' ? consumer.quota : 0,
    });
  };

  const closeQuotaModal = () => {
    setQuotaModalType(null);
    setCurrentConsumer(null);
    quotaForm.resetFields();
  };

  const submitQuotaModal = async () => {
    if (!selectedRouteName || !currentConsumer || !quotaModalType) {
      return;
    }
    const values = await quotaForm.validateFields();
    if (quotaModalType === 'refresh') {
      await refreshAiQuota(selectedRouteName, currentConsumer.consumerName, values.value);
      message.success(t('aiQuota.messages.refreshSuccess'));
    } else {
      await deltaAiQuota(selectedRouteName, currentConsumer.consumerName, values.value);
      message.success(t('aiQuota.messages.deltaSuccess'));
    }
    closeQuotaModal();
    await loadConsumers(selectedRouteName);
  };

  const openScheduleDrawer = async (consumer: AiQuotaConsumerQuota) => {
    if (!selectedRouteName) {
      return;
    }
    setScheduleConsumer(consumer);
    setEditingScheduleRule(null);
    scheduleForm.setFieldsValue({
      action: 'REFRESH',
      cron: '0 0 0 * * *',
      value: consumer.quota,
      enabled: true,
    });
    setScheduleDrawerOpen(true);
    await loadSchedules(selectedRouteName, consumer.consumerName);
  };

  const closeScheduleDrawer = () => {
    setScheduleDrawerOpen(false);
    setScheduleConsumer(null);
    setEditingScheduleRule(null);
    setScheduleRules([]);
    scheduleForm.resetFields();
  };

  const submitScheduleRule = async () => {
    if (!selectedRouteName || !scheduleConsumer) {
      return;
    }
    const values = await scheduleForm.validateFields();
    const payload: AiQuotaScheduleRuleRequest = {
      id: editingScheduleRule?.id,
      consumerName: scheduleConsumer.consumerName,
      action: values.action as AiQuotaScheduleAction,
      cron: values.cron,
      value: values.value,
      enabled: values.enabled,
    };
    await saveAiQuotaScheduleRule(selectedRouteName, payload);
    message.success(t('aiQuota.messages.scheduleSaved'));
    setEditingScheduleRule(null);
    scheduleForm.setFieldsValue({
      action: 'REFRESH',
      cron: '0 0 0 * * *',
      value: scheduleConsumer.quota,
      enabled: true,
    });
    await loadSchedules(selectedRouteName, scheduleConsumer.consumerName);
    await loadRoutes();
  };

  const editScheduleRule = (rule: AiQuotaScheduleRule) => {
    setEditingScheduleRule(rule);
    scheduleForm.setFieldsValue({
      action: rule.action,
      cron: rule.cron,
      value: rule.value,
      enabled: rule.enabled,
    });
  };

  const removeScheduleRule = async (rule: AiQuotaScheduleRule) => {
    if (!selectedRouteName || !scheduleConsumer) {
      return;
    }
    await deleteAiQuotaScheduleRule(selectedRouteName, rule.id);
    message.success(t('aiQuota.messages.scheduleDeleted'));
    if (editingScheduleRule?.id === rule.id) {
      setEditingScheduleRule(null);
      scheduleForm.setFieldsValue({
        action: 'REFRESH',
        cron: '0 0 0 * * *',
        value: scheduleConsumer.quota,
        enabled: true,
      });
    }
    await loadSchedules(selectedRouteName, scheduleConsumer.consumerName);
    await loadRoutes();
  };

  const resetScheduleForm = () => {
    setEditingScheduleRule(null);
    scheduleForm.setFieldsValue({
      action: 'REFRESH',
      cron: '0 0 0 * * *',
      value: scheduleConsumer?.quota ?? 0,
      enabled: true,
    });
  };

  const quotaColumns = [
    {
      title: t('aiQuota.columns.consumer'),
      dataIndex: 'consumerName',
      key: 'consumerName',
    },
    {
      title: t('aiQuota.columns.quota'),
      dataIndex: 'quota',
      key: 'quota',
      render: (value: number) => value ?? 0,
    },
    {
      title: t('aiQuota.columns.actions'),
      key: 'actions',
      width: 220,
      render: (_: unknown, record: AiQuotaConsumerQuota) => (
        <Space size="small">
          <a onClick={() => openQuotaModal('refresh', record)}>{t('aiQuota.actions.refresh')}</a>
          <a onClick={() => openQuotaModal('delta', record)}>{t('aiQuota.actions.delta')}</a>
          <a onClick={() => openScheduleDrawer(record)}>{t('aiQuota.actions.schedule')}</a>
        </Space>
      ),
    },
  ];

  const scheduleColumns = [
    {
      title: t('aiQuota.schedule.columns.action'),
      dataIndex: 'action',
      key: 'action',
      render: (value: AiQuotaScheduleAction) => (
        <Tag color={value === 'REFRESH' ? 'blue' : 'green'}>
          {value === 'REFRESH' ? t('aiQuota.schedule.actions.refresh') : t('aiQuota.schedule.actions.delta')}
        </Tag>
      ),
    },
    {
      title: t('aiQuota.schedule.columns.cron'),
      dataIndex: 'cron',
      key: 'cron',
    },
    {
      title: t('aiQuota.schedule.columns.value'),
      dataIndex: 'value',
      key: 'value',
    },
    {
      title: t('aiQuota.schedule.columns.enabled'),
      dataIndex: 'enabled',
      key: 'enabled',
      render: (value: boolean) => (value ? t('misc.enabled') : t('misc.disabled')),
    },
    {
      title: t('aiQuota.schedule.columns.lastAppliedAt'),
      dataIndex: 'lastAppliedAt',
      key: 'lastAppliedAt',
      render: (value?: number) => (value ? new Date(value).toLocaleString() : '-'),
    },
    {
      title: t('aiQuota.schedule.columns.lastError'),
      dataIndex: 'lastError',
      key: 'lastError',
      ellipsis: true,
      render: (value?: string) => value || '-',
    },
    {
      title: t('aiQuota.columns.actions'),
      key: 'actions',
      width: 140,
      render: (_: unknown, record: AiQuotaScheduleRule) => (
        <Space size="small">
          <a onClick={() => editScheduleRule(record)}>{t('misc.edit')}</a>
          <a onClick={() => removeScheduleRule(record)}>{t('misc.delete')}</a>
        </Space>
      ),
    },
  ];

  if (!routesLoading && routes.length === 0) {
    return (
      <PageContainer>
        <div style={{ background: '#fff', padding: 24 }}>
          <Empty description={t('aiQuota.empty')}>
            <Button icon={<RedoOutlined />} onClick={() => loadRoutes()}>
              {t('misc.refresh')}
            </Button>
          </Empty>
        </div>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <div style={{ background: '#fff', padding: 24, marginBottom: 16 }}>
        <Space wrap style={{ width: '100%', justifyContent: 'space-between' }}>
          <Space wrap size={16}>
            <div>
              <div style={{ marginBottom: 8 }}>{t('aiQuota.route')}</div>
              <Select
                style={{ width: 320 }}
                value={selectedRouteName}
                onChange={setSelectedRouteName}
                options={routes.map((route) => ({
                  label: route.routeName,
                  value: route.routeName,
                }))}
              />
            </div>
            <div>
              <div style={{ marginBottom: 8 }}>{t('aiQuota.search')}</div>
              <Input
                style={{ width: 260 }}
                allowClear
                value={keyword}
                placeholder={t('aiQuota.searchPlaceholder') as string}
                onChange={(event) => setKeyword(event.target.value)}
              />
            </div>
          </Space>
          <Button icon={<RedoOutlined />} onClick={refreshAll}>
            {t('misc.refresh')}
          </Button>
        </Space>
      </div>

      {selectedRoute && (
        <div style={{ background: '#fff', padding: 24, marginBottom: 16 }}>
          <Descriptions column={2} size="small">
            <Descriptions.Item label={t('aiQuota.summary.route')}>
              {selectedRoute.routeName}
            </Descriptions.Item>
            <Descriptions.Item label={t('aiQuota.summary.path')}>
              {selectedRoute.path || '-'}
            </Descriptions.Item>
            <Descriptions.Item label={t('aiQuota.summary.domains')}>
              {selectedRoute.domains?.length ? selectedRoute.domains.join(', ') : '-'}
            </Descriptions.Item>
            <Descriptions.Item label={t('aiQuota.summary.redisKeyPrefix')}>
              {selectedRoute.redisKeyPrefix}
            </Descriptions.Item>
            <Descriptions.Item label={t('aiQuota.summary.adminConsumer')}>
              {selectedRoute.adminConsumer}
            </Descriptions.Item>
            <Descriptions.Item label={t('aiQuota.summary.adminPath')}>
              {selectedRoute.adminPath}
            </Descriptions.Item>
          </Descriptions>
        </div>
      )}

      <div style={{ background: '#fff', padding: 24 }}>
        <Table
          rowKey="consumerName"
          loading={quotaLoading}
          dataSource={filteredQuotaList}
          columns={quotaColumns}
          pagination={{
            showSizeChanger: true,
            showTotal: (total) => `${t('misc.total')} ${total}`,
          }}
        />
      </div>

      <Modal
        title={
          quotaModalType === 'refresh' ? t('aiQuota.modals.refreshTitle') : t('aiQuota.modals.deltaTitle')
        }
        open={!!quotaModalType}
        onCancel={closeQuotaModal}
        onOk={submitQuotaModal}
        destroyOnClose
      >
        <Form form={quotaForm} layout="vertical">
          <Form.Item label={t('aiQuota.columns.consumer')}>
            <Text>{currentConsumer?.consumerName}</Text>
          </Form.Item>
          <Form.Item
            name="value"
            label={quotaModalType === 'refresh' ? t('aiQuota.modals.refreshValue') : t('aiQuota.modals.deltaValue')}
            rules={[
              {
                required: true,
                message:
                  (quotaModalType === 'refresh'
                    ? t('aiQuota.validation.refreshValueRequired')
                    : t('aiQuota.validation.deltaValueRequired')) || '',
              },
            ]}
          >
            <InputNumber style={{ width: '100%' }} precision={0} />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        width={760}
        title={t('aiQuota.schedule.title')}
        open={scheduleDrawerOpen}
        onClose={closeScheduleDrawer}
        destroyOnClose
      >
        <div style={{ marginBottom: 16 }}>
          <Text strong>{t('aiQuota.columns.consumer')}:</Text>{' '}
          <Text>{scheduleConsumer?.consumerName || '-'}</Text>
        </div>
        <Form form={scheduleForm} layout="vertical">
          <Form.Item
            name="action"
            label={t('aiQuota.schedule.form.action')}
            rules={[{ required: true, message: t('aiQuota.validation.scheduleActionRequired') || '' }]}
          >
            <Select
              options={[
                { label: t('aiQuota.schedule.actions.refresh'), value: 'REFRESH' },
                { label: t('aiQuota.schedule.actions.delta'), value: 'DELTA' },
              ]}
            />
          </Form.Item>
          <Form.Item
            name="cron"
            label={t('aiQuota.schedule.form.cron')}
            extra={t('aiQuota.schedule.form.cronHelp')}
            rules={[{ required: true, message: t('aiQuota.validation.scheduleCronRequired') || '' }]}
          >
            <Input placeholder="0 0 0 * * *" />
          </Form.Item>
          <Form.Item
            name="value"
            label={t('aiQuota.schedule.form.value')}
            rules={[{ required: true, message: t('aiQuota.validation.scheduleValueRequired') || '' }]}
          >
            <InputNumber style={{ width: '100%' }} precision={0} />
          </Form.Item>
          <Form.Item name="enabled" label={t('aiQuota.schedule.form.enabled')} valuePropName="checked">
            <Switch />
          </Form.Item>
          <Space style={{ marginBottom: 24 }}>
            <Button type="primary" onClick={submitScheduleRule}>
              {editingScheduleRule ? t('misc.save') : t('misc.create')}
            </Button>
            <Button onClick={resetScheduleForm}>{t('misc.reset')}</Button>
          </Space>
        </Form>

        <Table
          rowKey="id"
          loading={scheduleLoading}
          dataSource={scheduleRules}
          columns={scheduleColumns}
          pagination={false}
        />
      </Drawer>
    </PageContainer>
  );
};

export default AiQuotaPage;
