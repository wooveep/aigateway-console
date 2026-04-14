import type {
  PortalDepartmentBillRecord,
  PortalDepartmentBillsQuery,
  PortalUsageEventRecord,
  PortalUsageEventsQuery,
  PortalUsageStatRecord,
} from '@/interfaces/portal-stats';
import request, { type RequestOptions } from './request';

const QUIET_PORTAL_STATS_OPTIONS: RequestOptions = {
  skipErrorModal: true,
};

export const getPortalUsageStats = (params?: { from?: number; to?: number }) => {
  return request.get<any, PortalUsageStatRecord[]>('/v1/portal/stats/usage', {
    ...QUIET_PORTAL_STATS_OPTIONS,
    params,
  });
};

export const getPortalUsageEvents = (params?: PortalUsageEventsQuery) => {
  return request.get<any, PortalUsageEventRecord[]>('/v1/portal/stats/usage-events', {
    ...QUIET_PORTAL_STATS_OPTIONS,
    params,
  });
};

export const getPortalDepartmentBills = (params?: PortalDepartmentBillsQuery) => {
  return request.get<any, PortalDepartmentBillRecord[]>('/v1/portal/stats/department-bills', {
    ...QUIET_PORTAL_STATS_OPTIONS,
    params,
  });
};
