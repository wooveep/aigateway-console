export interface PortalUsageStatRecord {
  consumerName: string;
  modelName: string;
  requestCount: number;
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
  cacheCreationInputTokens: number;
  cacheCreation5mInputTokens: number;
  cacheCreation1hInputTokens: number;
  cacheReadInputTokens: number;
  inputImageTokens: number;
  outputImageTokens: number;
  inputImageCount: number;
  outputImageCount: number;
}

export interface PortalUsageEventRecord {
  eventId?: string;
  requestId?: string;
  traceId?: string;
  consumerName?: string;
  departmentId?: string;
  departmentPath?: string;
  apiKeyId?: string;
  modelId?: string;
  priceVersionId?: number;
  routeName?: string;
  requestKind?: string;
  requestStatus?: string;
  usageStatus?: string;
  httpStatus?: number;
  inputTokens?: number;
  outputTokens?: number;
  totalTokens?: number;
  requestCount?: number;
  costMicroYuan?: number;
  occurredAt?: string;
}

export interface PortalDepartmentBillRecord {
  departmentId?: string;
  departmentName?: string;
  departmentPath?: string;
  requestCount?: number;
  totalTokens?: number;
  totalCost?: number;
  activeConsumers?: number;
}

export interface PortalUsageEventsQuery {
  from?: number;
  to?: number;
  consumerName?: string;
  departmentId?: string;
  includeChildren?: boolean;
  apiKeyId?: string;
  modelId?: string;
  routeName?: string;
  requestStatus?: string;
  usageStatus?: string;
  pageNum?: number;
  pageSize?: number;
}

export interface PortalDepartmentBillsQuery {
  from?: number;
  to?: number;
  departmentId?: string;
  includeChildren?: boolean;
}
