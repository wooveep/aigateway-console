export interface AiQuotaMenuState {
  enabled: boolean;
  enabledRouteCount: number;
}

export interface AiQuotaRouteSummary {
  routeName: string;
  domains?: string[];
  path?: string;
  redisKeyPrefix: string;
  adminConsumer: string;
  adminPath: string;
  scheduleRuleCount: number;
}

export interface AiQuotaConsumerQuota {
  consumerName: string;
  quota: number;
}

export type AiQuotaScheduleAction = 'REFRESH' | 'DELTA';

export interface AiQuotaScheduleRule {
  id: string;
  consumerName: string;
  action: AiQuotaScheduleAction;
  cron: string;
  value: number;
  enabled: boolean;
  createdAt?: number;
  updatedAt?: number;
  lastAppliedAt?: number;
  lastError?: string;
}

export interface AiQuotaValueRequest {
  value: number;
}

export interface AiQuotaScheduleRuleRequest {
  id?: string;
  consumerName: string;
  action: AiQuotaScheduleAction;
  cron: string;
  value: number;
  enabled?: boolean;
}
