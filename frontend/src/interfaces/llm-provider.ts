export interface LlmProvider {
  key?: string;
  name: string;
  type: string;
  protocol?: string;
  proxyName?: string;
  tokens: string[];
  tokenFailoverConfig?: TokeFailoverConfig;
  rawConfigs?: LlmProviderRawConfigs;
}

export interface TokeFailoverConfig {
  enabled?: boolean;
  failureThreshold?: number;
  successThreshold?: number;
  healthCheckInterval?: number;
  healthCheckTimeout?: number;
  healthCheckModel?: string;
}

export enum LlmProviderProtocol {
  OPENAI_V1 = 'openai/v1',
}

export interface LlmProviderRawConfigs {
  portalModelMeta?: PortalModelMeta;
  [prop: string]: any;
}

export interface PortalModelMeta {
  intro?: string;
  tags?: string[];
  capabilities?: {
    modalities?: string[];
    features?: string[];
  };
  pricing?: {
    currency?: 'CNY';
    inputPer1K?: number;
    outputPer1K?: number;
  };
  limits?: {
    rpm?: number;
    tpm?: number;
    contextWindow?: number;
  };
}
