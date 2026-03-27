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
  quotaUnit?: string;
  scheduleRuleCount: number;
}

export interface AiQuotaConsumerQuota {
  consumerName: string;
  quota: number;
}

export interface AiQuotaUserPolicy {
  consumerName: string;
  limitTotal: number;
  limit5h: number;
  limitDaily: number;
  dailyResetMode: string;
  dailyResetTime: string;
  limitWeekly: number;
  limitMonthly: number;
  costResetAt?: string;
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

export interface AiQuotaUserPolicyRequest {
  limitTotal: number;
  limit5h: number;
  limitDaily: number;
  dailyResetMode?: string;
  dailyResetTime?: string;
  limitWeekly: number;
  limitMonthly: number;
  costResetAt?: string;
}
