import { PageQuery } from './common';

export interface McpServerPageQuery extends PageQuery {
  mcpServerName?: string;
  type?: string;
}

export interface McpPresetTemplate {
  id: string;
  name: string;
}

export interface McpServer {
  name: string;
  type: string; // OPEN_API/DATABASE/REDIRECT_ROUTE
  description?: string;
  domains?: string[];
  services?: UpstreamService[];
  rawConfigurations?: string;
  dsn?: string;
  dbType?: string; // MYSQL/PostgreSQL/Sqlite/Clickhouse
  upstreamPathPrefix?: string;
  consumerAuthInfo?: ConsumerAuthInfo;
}

export interface McpServerConsumerDetail {
  mcpServerName: string;
  consumerName: string;
  type: string;
}

interface UpstreamService {
  // 根据实际服务结构补充
  name: string;
  port: number;
  version: string;
  weight: number;
}

interface ConsumerAuthInfo {
  // 根据实际认证结构补充
  type: string;
  enable: boolean;
  allowedConsumers: string[];
  allowedConsumerLevels?: Array<'normal' | 'plus' | 'pro' | 'ultra' | string>;
}
