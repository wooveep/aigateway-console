import { Consumer } from '@/interfaces/consumer';
import { AutoComplete, Form, Input, Select } from 'antd';
import React, { forwardRef, useEffect, useImperativeHandle } from 'react';
import { useTranslation } from 'react-i18next';

interface Props {
  value?: Consumer | null;
  departments?: string[];
  presetDepartment?: string;
}

const ConsumerForm: React.FC<Props> = forwardRef((props, ref) => {
  const { t } = useTranslation();
  const { value, departments = [], presetDepartment } = props;
  const [form] = Form.useForm();

  useEffect(() => {
    if (value) {
      form.setFieldsValue({
        department: value.department,
        name: value.name,
        portalDisplayName: value.portalDisplayName,
        portalEmail: value.portalEmail,
        portalUserLevel: value.portalUserLevel || 'normal',
      });
    } else {
      form.resetFields();
      form.setFieldValue('portalUserLevel', 'normal');
      if (presetDepartment) {
        form.setFieldValue('department', presetDepartment);
      }
    }
  }, [form, presetDepartment, value]);

  useImperativeHandle(ref, () => ({
    reset: () => {
      form.resetFields();
    },
    handleSubmit: async () => {
      const values = await form.validateFields();
      return {
        ...values,
        credentials: [],
      };
    },
  }));

  return (
    <Form form={form} layout="vertical">
      <Form.Item label={t('consumer.consumerForm.department')} name="department">
        <AutoComplete
          options={departments.map((department) => ({ value: department }))}
          filterOption={(inputValue, option) =>
            (option?.value || '').toUpperCase().includes(inputValue.toUpperCase())
          }
        >
          <Input
            showCount
            allowClear
            maxLength={63}
            placeholder={t('consumer.consumerForm.departmentPlaceholder') || ''}
          />
        </AutoComplete>
      </Form.Item>
      <Form.Item
        label={t('consumer.consumerForm.name')}
        required
        name="name"
        rules={[
          {
            required: true,
            message: t('consumer.consumerForm.nameRequired') || '',
          },
        ]}
      >
        <Input
          showCount
          allowClear
          maxLength={63}
          placeholder={t('consumer.consumerForm.namePlaceholder') || ''}
          disabled={!!value}
        />
      </Form.Item>
      <Form.Item label="Portal显示名" name="portalDisplayName">
        <Input showCount allowClear maxLength={63} placeholder="可选，默认与用户名一致" />
      </Form.Item>
      <Form.Item label="Portal邮箱" name="portalEmail">
        <Input showCount allowClear maxLength={128} placeholder="可选" />
      </Form.Item>
      <Form.Item
        label={t('consumer.consumerForm.portalUserLevel')}
        name="portalUserLevel"
        rules={[{ required: true, message: t('consumer.consumerForm.portalUserLevelRequired') || '' }]}
      >
        <Select placeholder={t('consumer.consumerForm.portalUserLevelPlaceholder') || ''}>
          <Select.Option value="normal">{t('consumer.userLevel.normal')}</Select.Option>
          <Select.Option value="plus">{t('consumer.userLevel.plus')}</Select.Option>
          <Select.Option value="pro">{t('consumer.userLevel.pro')}</Select.Option>
          <Select.Option value="ultra">{t('consumer.userLevel.ultra')}</Select.Option>
        </Select>
      </Form.Item>
      <Form.Item label="Portal密码" name="portalPassword">
        <Input.Password placeholder={value ? '留空则不修改密码' : '留空将由系统生成临时密码'} />
      </Form.Item>
    </Form>
  );
});

export default ConsumerForm;
