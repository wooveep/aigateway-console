import { ReloadOutlined } from '@ant-design/icons';
import { Line } from '@ant-design/charts';
import { DashboardInfo, DashboardType, NativeDashboardPanel } from '@/interfaces/dashboard';
import { getNativeDashboard } from '@/services';
import { useRequest } from 'ahooks';
import { Alert, Button, Card, Collapse, Empty, Select, Spin, Statistic, Table } from 'antd';
import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import styles from './index.module.css';

const { Panel } = Collapse;

const DEFAULT_RANGE_MS = 5 * 60 * 1000;
const RANGE_OPTIONS = [5 * 60 * 1000, 15 * 60 * 1000, 60 * 60 * 1000, 6 * 60 * 60 * 1000, 24 * 60 * 60 * 1000];
const REFRESH_OPTIONS = [0, 15 * 1000, 30 * 1000, 60 * 1000];

interface NativeDashboardProps {
  type: DashboardType;
  dashboardInfo: DashboardInfo;
}

const NativeDashboard: React.FC<NativeDashboardProps> = ({ type, dashboardInfo }) => {
  const { t } = useTranslation();
  const [rangeMs, setRangeMs] = useState(DEFAULT_RANGE_MS);
  const [refreshMs, setRefreshMs] = useState(30 * 1000);
  const [activeRows, setActiveRows] = useState<string[]>([]);
  const [gateway, setGateway] = useState<string>();
  const [namespace, setNamespace] = useState<string>();

  const {
    data,
    error,
    loading,
    refresh,
  } = useRequest(() => {
    const to = Date.now();
    return getNativeDashboard(type, {
      from: to - rangeMs,
      to,
      gateway,
      namespace,
    });
  }, {
    pollingInterval: refreshMs > 0 ? refreshMs : undefined,
    refreshDeps: [type, rangeMs, gateway, namespace],
  });

  useEffect(() => {
    if (!data) {
      return;
    }
    if (activeRows.length === 0) {
      setActiveRows(data.rows.filter((row) => !row.collapsed).map((row) => row.title));
    }
  }, [activeRows.length, data]);

  if (loading && !data) {
    return (
      <div style={{ width: '100%', height: '50vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Spin />
      </div>
    );
  }

  if (error || !data) {
    return (
      <Alert
        message={t('dashboard.loadFailed')}
        type="error"
        showIcon
      />
    );
  }

  const currentGateway = gateway || data.variables.gateway.value;
  const currentNamespace = namespace || data.variables.namespace.value;
  const grafanaUrl = dashboardInfo.builtIn && dashboardInfo.url ? `${location.origin}${dashboardInfo.url}` : dashboardInfo.url;

  return (
    <div className={styles.wrapper}>
      <div className={styles.toolbar}>
        <div className={styles.toolbarMeta}>
          <div className={styles.control}>
            <span className={styles.controlLabel}>{t('dashboard.native.gateway')}</span>
            <Select
              value={currentGateway || undefined}
              options={data.variables.gateway.options.map((option) => ({ label: option, value: option }))}
              onChange={(value) => {
                setGateway(value);
                setNamespace(undefined);
              }}
            />
          </div>
          <div className={styles.control}>
            <span className={styles.controlLabel}>{t('dashboard.native.namespace')}</span>
            <span className={styles.namespace}>{currentNamespace || '-'}</span>
          </div>
          <div className={styles.control}>
            <span className={styles.controlLabel}>{t('dashboard.native.range')}</span>
            <Select
              value={rangeMs}
              options={RANGE_OPTIONS.map((option) => ({
                label: t(`dashboard.native.rangeOptions.${option}`),
                value: option,
              }))}
              onChange={(value) => setRangeMs(value)}
            />
          </div>
          <div className={styles.control}>
            <span className={styles.controlLabel}>{t('dashboard.native.refreshEvery')}</span>
            <Select
              value={refreshMs}
              options={REFRESH_OPTIONS.map((option) => ({
                label: t(`dashboard.native.refreshOptions.${option}`),
                value: option,
              }))}
              onChange={(value) => setRefreshMs(value)}
            />
          </div>
          <span className={styles.status}>
            {t('dashboard.native.lastUpdated', { time: formatDateTime(data.to) })}
          </span>
        </div>
        <div className={styles.toolbarAction}>
          <Button icon={<ReloadOutlined />} onClick={() => refresh()}>
            {t('dashboard.native.refresh')}
          </Button>
          {grafanaUrl && (
            <a href={grafanaUrl} target="_blank" rel="noreferrer">
              {t('dashboard.openInNewPage')}
            </a>
          )}
        </div>
      </div>

      <Collapse
        className={styles.collapse}
        activeKey={activeRows}
        onChange={(keys) => setActiveRows(Array.isArray(keys) ? keys as string[] : [keys as string])}
      >
        {data.rows.map((row) => (
          <Panel header={translateNativeText(t, 'rows', row.title)} key={row.title}>
            <div className={styles.grid}>
              {row.panels.map((panel) => (
                <div
                  className={styles.panelCell}
                  key={panel.id}
                  style={{
                    gridColumn: `${panel.gridPos.x + 1} / span ${Math.max(1, panel.gridPos.w)}`,
                  }}
                >
                  <DashboardPanelCard panel={panel} />
                </div>
              ))}
            </div>
          </Panel>
        ))}
      </Collapse>
    </div>
  );
};

const DashboardPanelCard: React.FC<{ panel: NativeDashboardPanel }> = ({ panel }) => {
  const { t } = useTranslation();
  const cardHeight = Math.max(panel.type === 'stat' ? 180 : 240, panel.gridPos.h * 38);
  const lineData = [];
  for (const series of panel.series || []) {
    for (const point of series.points) {
      lineData.push({
        time: point.time,
        series: translateNativeText(t, 'series', series.name),
        value: point.value,
      });
    }
  }

  return (
    <Card
      className={styles.panelCard}
      title={translateNativeText(t, 'titles', panel.title)}
      bordered={false}
      style={{ height: cardHeight }}
      bodyStyle={{ height: cardHeight - 56 }}
    >
      <div className={styles.panelBody}>
        {panel.error && (
          <Alert
            className={styles.panelError}
            message={panel.error}
            type="warning"
            showIcon
          />
        )}
        {panel.type === 'stat' && (
          <div className={styles.statWrap}>
            <Statistic
              value={panel.stat?.value ?? null}
              formatter={(value) => formatValue(value as number | null | undefined, panel.unit)}
            />
          </div>
        )}
        {panel.type === 'timeseries' && (
          <div className={styles.chartWrap}>
            {lineData.length > 0 ? (
              <Line
                data={lineData}
                xField="time"
                yField="value"
                seriesField="series"
                autoFit
                smooth={false}
                animation={false}
                padding="auto"
                xAxis={{
                  type: 'time',
                  label: {
                    formatter: (value) => formatTime(value),
                  },
                }}
                yAxis={{
                  label: {
                    formatter: (value) => formatAxisValue(Number(value), panel.unit),
                  },
                }}
                legend={{
                  position: 'top',
                }}
                tooltip={{
                  formatter: (datum) => ({
                    name: datum.series,
                    value: formatValue(datum.value as number, panel.unit),
                  }),
                }}
              />
            ) : (
              <div className={styles.emptyWrap}>
                <Empty description={t('dashboard.native.noData')} />
              </div>
            )}
          </div>
        )}
        {panel.type === 'table' && (
          <div className={styles.tableWrap}>
            {panel.table && panel.table.rows.length > 0 ? (
              <Table
                size="small"
                pagination={false}
                rowKey={(_, index) => `${panel.id}-${index}`}
                scroll={{ x: 'max-content' }}
                columns={panel.table.columns.map((column) => ({
                  title: translateColumnTitle(t, column.title || column.key),
                  dataIndex: column.key,
                  key: column.key,
                  render: (value: string | number | null) => formatTableValue(value),
                }))}
                dataSource={panel.table.rows}
              />
            ) : (
              <div className={styles.emptyWrap}>
                <Empty description={t('dashboard.native.noData')} />
              </div>
            )}
          </div>
        )}
      </div>
    </Card>
  );
};

function formatValue(value: number | null | undefined, unit: string) {
  if (value === null || value === undefined || Number.isNaN(value)) {
    return '-';
  }

  switch (unit) {
    case 'percentunit':
      return `${formatNumber(value * 100)}%`;
    case 'percent':
      return `${formatNumber(value)}%`;
    case 'reqps':
      return `${formatNumber(value)} req/s`;
    case 'Bps':
      return `${formatBytes(value)}/s`;
    case 'bytes':
      return formatBytes(value);
    case 'dtdurationms':
    case 'ms':
      return formatDuration(value);
    case 'ops':
      return `${formatNumber(value)} ops`;
    case 'short':
      return formatCompactNumber(value);
    default:
      return formatNumber(value);
  }
}

function formatAxisValue(value: number, unit: string) {
  if (Number.isNaN(value)) {
    return '';
  }
  const rendered = formatValue(value, unit);
  return typeof rendered === 'string' ? rendered : String(rendered);
}

function formatTableValue(value: string | number | null) {
  if (value === null || value === undefined) {
    return '-';
  }
  if (typeof value === 'number') {
    return formatNumber(value);
  }
  return value;
}

function formatNumber(value: number) {
  if (Math.abs(value) >= 1000) {
    return formatCompactNumber(value);
  }
  if (Math.abs(value) >= 100) {
    return value.toFixed(1);
  }
  if (Math.abs(value) >= 10) {
    return value.toFixed(2);
  }
  return value.toFixed(3).replace(/\.?0+$/, '');
}

function formatCompactNumber(value: number) {
  return new Intl.NumberFormat(undefined, {
    notation: 'compact',
    maximumFractionDigits: 1,
  }).format(value);
}

function formatBytes(value: number) {
  if (value === 0) {
    return '0 B';
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let current = value;
  let unitIndex = 0;
  while (current >= 1024 && unitIndex < units.length - 1) {
    current /= 1024;
    unitIndex += 1;
  }
  return `${formatNumber(current)} ${units[unitIndex]}`;
}

function formatDuration(value: number) {
  if (value >= 1000) {
    return `${formatNumber(value / 1000)} s`;
  }
  return `${formatNumber(value)} ms`;
}

function formatDateTime(timestamp: number) {
  return new Date(timestamp).toLocaleString();
}

function formatTime(value: string | number) {
  const date = typeof value === 'number' ? new Date(value) : new Date(value);
  if (Number.isNaN(date.getTime())) {
    return typeof value === 'number' ? String(value) : value;
  }
  return date.toLocaleTimeString();
}

function translateNativeText(t: (key: string, options?: any) => string, group: string, value?: string) {
  if (!value) {
    return value || '';
  }
  const key = `dashboard.native.${group}.${value}`;
  const translated = t(key);
  return translated === key ? value : translated;
}

function translateColumnTitle(t: (key: string, options?: any) => string, value?: string) {
  if (!value) {
    return t('dashboard.native.columns.defaultDimension');
  }
  return translateNativeText(t, 'columns', value);
}

export default NativeDashboard;
