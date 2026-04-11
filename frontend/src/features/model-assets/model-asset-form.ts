import type {
  ModelAsset,
  ModelAssetBinding,
  ModelAssetOptions,
  ModelBindingPricing,
} from '@/interfaces/model-asset';
import type { AssetGrantRecord, OrgDepartmentNode } from '@/interfaces/org';
import { MODEL_ASSET_PRESET_TAGS } from '@/interfaces/model-asset';

export type AssetFormState = {
  assetId: string;
  canonicalName: string;
  displayName: string;
  intro: string;
  tags: string[];
  modalities: string[];
  features: string[];
  requestKinds: string[];
};

export type BindingFormState = {
  bindingId: string;
  modelId: string;
  providerName: string;
  targetModel: string;
  protocol: string;
  endpoint: string;
  rpm?: number;
  tpm?: number;
  contextWindow?: number;
  currency: string;
  supportsPromptCaching: boolean;
  inputCostPerToken?: number;
  outputCostPerToken?: number;
  inputCostPerRequest?: number;
  cacheCreationInputTokenCost?: number;
  cacheCreationInputTokenCostAbove1hr?: number;
  cacheReadInputTokenCost?: number;
  inputCostPerTokenAbove200kTokens?: number;
  outputCostPerTokenAbove200kTokens?: number;
  cacheCreationInputTokenCostAbove200kTokens?: number;
  cacheReadInputTokenCostAbove200kTokens?: number;
  outputCostPerImage?: number;
  outputCostPerImageToken?: number;
  inputCostPerImage?: number;
  inputCostPerImageToken?: number;
};

type PricingUnitKind = 'token' | 'request' | 'image';

export const pricingFieldGroups: Array<{
  title: string;
  fields: Array<{ name: keyof BindingFormState; label: string; step?: number; unitKind?: PricingUnitKind }>;
}> = [
  {
    title: '基础价',
    fields: [
      { name: 'inputCostPerToken', label: '输入 Token 单价', step: 0.000001, unitKind: 'token' },
      { name: 'outputCostPerToken', label: '输出 Token 单价', step: 0.000001, unitKind: 'token' },
      { name: 'inputCostPerRequest', label: '按请求计价', step: 0.000001, unitKind: 'request' },
      { name: 'cacheCreationInputTokenCost', label: 'Cache 写入 Token 单价', step: 0.000001, unitKind: 'token' },
      { name: 'cacheCreationInputTokenCostAbove1hr', label: 'Cache 写入 Token 单价（>1h）', step: 0.000001, unitKind: 'token' },
      { name: 'cacheReadInputTokenCost', label: 'Cache 读取 Token 单价', step: 0.000001, unitKind: 'token' },
      { name: 'outputCostPerImage', label: '输出图片单价', step: 0.000001, unitKind: 'image' },
      { name: 'outputCostPerImageToken', label: '输出图片 Token 单价', step: 0.000001, unitKind: 'token' },
      { name: 'inputCostPerImage', label: '输入图片单价', step: 0.000001, unitKind: 'image' },
      { name: 'inputCostPerImageToken', label: '输入图片 Token 单价', step: 0.000001, unitKind: 'token' },
    ],
  },
  {
    title: 'above_200k',
    fields: [
      { name: 'inputCostPerTokenAbove200kTokens', label: '输入 Token 单价（>200k）', step: 0.000001, unitKind: 'token' },
      { name: 'outputCostPerTokenAbove200kTokens', label: '输出 Token 单价（>200k）', step: 0.000001, unitKind: 'token' },
      { name: 'cacheCreationInputTokenCostAbove200kTokens', label: 'Cache 写入 Token 单价（>200k）', step: 0.000001, unitKind: 'token' },
      { name: 'cacheReadInputTokenCostAbove200kTokens', label: 'Cache 读取 Token 单价（>200k）', step: 0.000001, unitKind: 'token' },
    ],
  },
];

export const statusColorMap: Record<string, string> = {
  draft: 'default',
  published: 'green',
  unpublished: 'orange',
  active: 'green',
  inactive: 'default',
  disabled: 'red',
};

export function toAssetFormState(asset?: ModelAsset): AssetFormState {
  return {
    assetId: asset?.assetId || '',
    canonicalName: asset?.canonicalName || '',
    displayName: asset?.displayName || '',
    intro: asset?.intro || '',
    tags: asset?.tags || [],
    modalities: asset?.capabilities?.modalities || [],
    features: asset?.capabilities?.features || [],
    requestKinds: asset?.capabilities?.requestKinds || [],
  };
}

export function toBindingFormState(binding?: ModelAssetBinding): BindingFormState {
  const pricing = binding?.pricing || {};
  return {
    bindingId: binding?.bindingId || '',
    modelId: binding?.modelId || '',
    providerName: binding?.providerName || '',
    targetModel: binding?.targetModel || '',
    protocol: binding?.protocol || 'openai/v1',
    endpoint: binding?.endpoint || '',
    rpm: binding?.limits?.rpm,
    tpm: binding?.limits?.tpm,
    contextWindow: binding?.limits?.contextWindow,
    currency: pricing.currency || 'CNY',
    supportsPromptCaching: Boolean(pricing.supportsPromptCaching),
    inputCostPerToken: pricing.inputCostPerToken,
    outputCostPerToken: pricing.outputCostPerToken,
    inputCostPerRequest: pricing.inputCostPerRequest,
    cacheCreationInputTokenCost: pricing.cacheCreationInputTokenCost,
    cacheCreationInputTokenCostAbove1hr: pricing.cacheCreationInputTokenCostAbove1hr,
    cacheReadInputTokenCost: pricing.cacheReadInputTokenCost,
    inputCostPerTokenAbove200kTokens: pricing.inputCostPerTokenAbove200kTokens,
    outputCostPerTokenAbove200kTokens: pricing.outputCostPerTokenAbove200kTokens,
    cacheCreationInputTokenCostAbove200kTokens: pricing.cacheCreationInputTokenCostAbove200kTokens,
    cacheReadInputTokenCostAbove200kTokens: pricing.cacheReadInputTokenCostAbove200kTokens,
    outputCostPerImage: pricing.outputCostPerImage,
    outputCostPerImageToken: pricing.outputCostPerImageToken,
    inputCostPerImage: pricing.inputCostPerImage,
    inputCostPerImageToken: pricing.inputCostPerImageToken,
  };
}

export function buildPricing(values: BindingFormState): ModelBindingPricing {
  const pricing: ModelBindingPricing = {
    currency: values.currency || 'CNY',
    supportsPromptCaching: values.supportsPromptCaching,
  };
  const numericFields: Array<keyof BindingFormState> = [
    'inputCostPerToken',
    'outputCostPerToken',
    'inputCostPerRequest',
    'cacheCreationInputTokenCost',
    'cacheCreationInputTokenCostAbove1hr',
    'cacheReadInputTokenCost',
    'inputCostPerTokenAbove200kTokens',
    'outputCostPerTokenAbove200kTokens',
    'cacheCreationInputTokenCostAbove200kTokens',
    'cacheReadInputTokenCostAbove200kTokens',
    'outputCostPerImage',
    'outputCostPerImageToken',
    'inputCostPerImage',
    'inputCostPerImageToken',
  ];
  numericFields.forEach((field) => {
    const value = values[field];
    if (typeof value === 'number') {
      (pricing as any)[field] = value;
    }
  });
  return pricing;
}

export function describePricing(pricing?: ModelBindingPricing) {
  if (!pricing) {
    return '-';
  }
  const currency = pricing.currency || 'CNY';
  const items: string[] = [];
  if (typeof pricing.inputCostPerToken === 'number') items.push(`输入 ${pricing.inputCostPerToken} ${formatPricingUnit(currency, 'token')}`);
  if (typeof pricing.outputCostPerToken === 'number') items.push(`输出 ${pricing.outputCostPerToken} ${formatPricingUnit(currency, 'token')}`);
  if (typeof pricing.inputCostPerRequest === 'number') items.push(`请求 ${pricing.inputCostPerRequest} ${formatPricingUnit(currency, 'request')}`);
  if (typeof pricing.inputCostPerTokenAbove200kTokens === 'number') items.push(`输入>200k ${pricing.inputCostPerTokenAbove200kTokens} ${formatPricingUnit(currency, 'token')}`);
  if (typeof pricing.outputCostPerTokenAbove200kTokens === 'number') items.push(`输出>200k ${pricing.outputCostPerTokenAbove200kTokens} ${formatPricingUnit(currency, 'token')}`);
  if (pricing.supportsPromptCaching) items.push('支持缓存');
  return items.length ? items.join(' / ') : '-';
}

export function getPricingFieldExtra(
  field: { unitKind?: PricingUnitKind },
  currency?: string,
) {
  if (!field.unitKind) {
    return undefined;
  }
  const resolvedCurrency = currency || 'CNY';
  return `按 ${formatPricingUnit(resolvedCurrency, field.unitKind)} 计价`;
}

function formatPricingUnit(currency: string, unitKind: PricingUnitKind) {
  if (unitKind === 'request') {
    return `${currency} / request`;
  }
  if (unitKind === 'image') {
    return `${currency} / image`;
  }
  return `${currency} / token`;
}

export function flattenDepartmentOptions(nodes: OrgDepartmentNode[], level = 0): Array<{ label: string; value: string }> {
  return (nodes || []).flatMap((node) => {
    const prefix = level > 0 ? `${'  '.repeat(level)}- ` : '';
    return [
      { label: `${prefix}${node.name}`, value: node.departmentId },
      ...flattenDepartmentOptions(node.children || [], level + 1),
    ];
  });
}

export function splitGrantAssignments(grants: AssetGrantRecord[]) {
  return {
    consumers: (grants || []).filter((item) => item.subjectType === 'consumer').map((item) => item.subjectId || ''),
    departments: (grants || []).filter((item) => item.subjectType === 'department').map((item) => item.subjectId || ''),
    userLevels: (grants || []).filter((item) => item.subjectType === 'user_level').map((item) => item.subjectId || ''),
  };
}

export function buildGrantAssignments(bindingId: string, consumers: string[], departments: string[], userLevels: string[]) {
  const next: AssetGrantRecord[] = [];
  consumers.forEach((subjectId) => next.push({ assetType: 'model_binding', assetId: bindingId, subjectType: 'consumer', subjectId }));
  departments.forEach((subjectId) => next.push({ assetType: 'model_binding', assetId: bindingId, subjectType: 'department', subjectId }));
  userLevels.forEach((subjectId) => next.push({ assetType: 'model_binding', assetId: bindingId, subjectType: 'user_level', subjectId }));
  return next;
}

export function hasLegacyAssetValues(asset: ModelAsset | undefined, assetOptions: ModelAssetOptions) {
  if (!asset) {
    return { tags: false, capabilities: false };
  }
  const capabilitySets = {
    modalities: new Set(assetOptions.capabilities?.modalities || []),
    features: new Set(assetOptions.capabilities?.features || []),
    requestKinds: new Set(assetOptions.capabilities?.requestKinds || []),
  };
  return {
    tags: !!asset.tags?.some((tag) => !MODEL_ASSET_PRESET_TAGS.includes(tag as any)),
    capabilities:
      !!asset.capabilities?.modalities?.some((item) => !capabilitySets.modalities.has(item))
      || !!asset.capabilities?.features?.some((item) => !capabilitySets.features.has(item))
      || !!asset.capabilities?.requestKinds?.some((item) => !capabilitySets.requestKinds.has(item)),
  };
}
