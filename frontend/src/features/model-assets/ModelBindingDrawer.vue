<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import type {
  ModelAssetBinding,
  ModelAssetOptions,
  ModelBindingPriceVersion,
  ProviderModelOption,
} from '@/interfaces/model-asset';
import type { LlmProvider } from '@/interfaces/llm-provider';
import DrawerFooter from '@/components/common/DrawerFooter.vue';
import { buildPricing, describePricing, getPricingFieldExtra, pricingFieldGroups, toBindingFormState } from './model-asset-form';

const props = defineProps<{
  open: boolean;
  binding?: ModelAssetBinding | null;
  providers: LlmProvider[];
  assetOptions: ModelAssetOptions;
  activePriceVersion?: ModelBindingPriceVersion | null;
}>();

const emit = defineEmits<{
  'update:open': [value: boolean];
  submit: [payload: ModelAssetBinding, isEdit: boolean];
  'open-history': [];
}>();

const formState = reactive(toBindingFormState());
const formRef = ref();

watch(() => [props.open, props.binding], () => {
  Object.assign(formState, toBindingFormState(props.binding || undefined));
}, { immediate: true });

const providerModelCatalog = computed(() =>
  (props.assetOptions.providerModels || []).reduce<Record<string, ProviderModelOption[]>>((accumulator, item) => {
    accumulator[item.providerName] = item.models || [];
    return accumulator;
  }, {}),
);

const currentProviderModels = computed(() => providerModelCatalog.value[formState.providerName] || []);
const currentProviderUsesCatalog = computed(() => currentProviderModels.value.length > 0);

const currentModelIdOptions = computed(() => {
  const options = currentProviderModels.value.map((item) => ({
    label: item.modelId === item.targetModel ? item.modelId : `${item.modelId} / ${item.targetModel}`,
    value: item.modelId,
  }));
  if (formState.modelId && !currentProviderModels.value.some((item) => item.modelId === formState.modelId)) {
    options.unshift({ label: `历史值 / ${formState.modelId}`, value: formState.modelId });
  }
  return options;
});

const currentTargetModelOptions = computed(() => {
  const options = currentProviderModels.value.map((item) => ({
    label: item.targetModel === item.modelId ? item.targetModel : `${item.targetModel} / ${item.modelId}`,
    value: item.targetModel,
  }));
  if (formState.targetModel && !currentProviderModels.value.some((item) => item.targetModel === formState.targetModel)) {
    options.unshift({ label: `历史值 / ${formState.targetModel}`, value: formState.targetModel });
  }
  return options;
});

const hasLegacyCatalogValue = computed(() =>
  currentProviderUsesCatalog.value
  && (
    (Boolean(formState.modelId) && !currentProviderModels.value.some((item) => item.modelId === formState.modelId))
    || (Boolean(formState.targetModel) && !currentProviderModels.value.some((item) => item.targetModel === formState.targetModel))
  ),
);

function syncBindingModelPair(field: 'modelId' | 'targetModel', selectedValue?: string) {
  if (!currentProviderUsesCatalog.value) {
    return;
  }
  const matched = currentProviderModels.value.find((item) =>
    field === 'modelId' ? item.modelId === selectedValue : item.targetModel === selectedValue);
  if (matched) {
    formState.modelId = matched.modelId;
    formState.targetModel = matched.targetModel;
  }
}

function handleProviderChange(providerName: string) {
  formState.providerName = providerName;
  const providerModels = providerModelCatalog.value[providerName] || [];
  if (!providerModels.length) {
    return;
  }
  const matched = providerModels.find((item) =>
    item.modelId === formState.modelId || item.targetModel === formState.targetModel);
  if (matched) {
    formState.modelId = matched.modelId;
    formState.targetModel = matched.targetModel;
    return;
  }
  formState.modelId = '';
  formState.targetModel = '';
}

function close() {
  emit('update:open', false);
}

async function submit() {
  await formRef.value?.validate();
  emit('submit', {
    ...(props.binding || {}),
    bindingId: formState.bindingId.trim(),
    modelId: formState.modelId.trim(),
    providerName: formState.providerName.trim(),
    targetModel: formState.targetModel.trim(),
    protocol: formState.protocol.trim() || 'openai/v1',
    endpoint: formState.endpoint.trim(),
    pricing: buildPricing(formState),
    limits: {
      rpm: formState.rpm,
      tpm: formState.tpm,
      contextWindow: formState.contextWindow,
    },
  }, Boolean(props.binding));
}
</script>

<template>
  <a-drawer
    :open="open"
    width="860"
    :title="binding ? '编辑发布绑定' : '新建发布绑定'"
    destroy-on-close
    @update:open="(value) => emit('update:open', value)"
  >
    <a-alert
      v-if="binding"
      type="info"
      show-icon
      style="margin-bottom: 16px"
      :message="`当前状态：${binding.status || 'draft'}`"
      :description="`发布时间：${binding.publishedAt || '-'}；下架时间：${binding.unpublishedAt || '-'}`"
    />

    <a-card v-if="activePriceVersion" size="small" title="当前生效价格版本" style="margin-bottom: 16px">
      <div class="model-binding-drawer__active">
        <span>版本 #{{ activePriceVersion.versionId }}</span>
        <span>生效时间 {{ activePriceVersion.effectiveFrom || '-' }}</span>
        <span>{{ describePricing(activePriceVersion.pricing) }}</span>
      </div>
    </a-card>

    <a-alert
      v-if="formState.providerName && !currentProviderUsesCatalog"
      type="info"
      show-icon
      style="margin-bottom: 16px"
      message="当前 Provider 未配置系统预置模型目录，模型 ID 和目标模型继续手填即可。"
    />

    <a-alert
      v-if="hasLegacyCatalogValue"
      type="warning"
      show-icon
      style="margin-bottom: 16px"
      message="当前绑定包含历史模型值，建议重新选择预置目录中的模型。"
    />

    <a-form ref="formRef" layout="vertical" :model="formState">
      <div class="model-binding-drawer__grid">
        <a-form-item
          label="绑定 ID"
          name="bindingId"
          :rules="[{ required: true, message: '请输入绑定 ID' }]"
        >
          <a-input v-model:value="formState.bindingId" :disabled="Boolean(binding)" />
        </a-form-item>
        <a-form-item
          label="Provider"
          name="providerName"
          :rules="[{ required: true, message: '请选择 Provider' }]"
        >
          <a-select
            :value="formState.providerName"
            show-search
            :options="providers.map((item) => ({ label: item.name, value: item.name }))"
            @update:value="handleProviderChange"
          />
        </a-form-item>
        <a-form-item
          label="可展示模型 ID"
          name="modelId"
          :rules="[{ required: true, message: '请输入模型 ID' }]"
        >
          <a-select
            v-if="currentProviderUsesCatalog"
            :value="formState.modelId"
            show-search
            :options="currentModelIdOptions"
            @update:value="(value) => syncBindingModelPair('modelId', String(value || ''))"
          />
          <a-input v-else v-model:value="formState.modelId" />
        </a-form-item>
        <a-form-item
          label="目标模型"
          name="targetModel"
          :rules="[{ required: true, message: '请输入目标模型' }]"
        >
          <a-select
            v-if="currentProviderUsesCatalog"
            :value="formState.targetModel"
            show-search
            :options="currentTargetModelOptions"
            @update:value="(value) => syncBindingModelPair('targetModel', String(value || ''))"
          />
          <a-input v-else v-model:value="formState.targetModel" />
        </a-form-item>
        <a-form-item label="协议">
          <a-input v-model:value="formState.protocol" />
        </a-form-item>
        <a-form-item label="入口地址">
          <a-input v-model:value="formState.endpoint" />
        </a-form-item>
      </div>

      <a-divider orientation="left">限制</a-divider>
      <div class="model-binding-drawer__grid model-binding-drawer__grid--compact">
        <a-form-item label="RPM">
          <a-input-number v-model:value="formState.rpm" style="width: 100%" :min="0" />
        </a-form-item>
        <a-form-item label="TPM">
          <a-input-number v-model:value="formState.tpm" style="width: 100%" :min="0" />
        </a-form-item>
        <a-form-item label="Context Window">
          <a-input-number v-model:value="formState.contextWindow" style="width: 100%" :min="0" />
        </a-form-item>
      </div>

      <a-divider orientation="left">价格</a-divider>
      <div class="model-binding-drawer__grid model-binding-drawer__grid--compact">
        <a-form-item label="币种">
          <a-input v-model:value="formState.currency" />
        </a-form-item>
        <a-form-item label="支持 Prompt Cache">
          <a-switch v-model:checked="formState.supportsPromptCaching" />
        </a-form-item>
      </div>

      <section v-for="group in pricingFieldGroups" :key="group.title" class="model-binding-drawer__group">
        <h4>{{ group.title }}</h4>
        <div class="model-binding-drawer__grid">
          <a-form-item
            v-for="field in group.fields"
            :key="String(field.name)"
            :label="field.label"
            :extra="getPricingFieldExtra(field, formState.currency)"
          >
            <a-input-number
              :value="typeof formState[field.name] === 'number' ? formState[field.name] as number : undefined"
              style="width: 100%"
              :min="0"
              :step="field.step || 0.000001"
              @update:value="(value) => ((formState as any)[field.name] = value ?? undefined)"
            />
          </a-form-item>
        </div>
      </section>
    </a-form>

    <DrawerFooter @cancel="close" @confirm="submit">
      <template #extra>
        <a-button v-if="binding" @click="emit('open-history')">价格历史</a-button>
      </template>
    </DrawerFooter>
  </a-drawer>
</template>

<style scoped>
.model-binding-drawer__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 14px;
}

.model-binding-drawer__grid--compact {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.model-binding-drawer__group {
  margin-top: 12px;
}

.model-binding-drawer__group h4 {
  margin: 0 0 12px;
  font-size: 13px;
}

.model-binding-drawer__active {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  color: var(--portal-text-soft);
  font-size: 12px;
}

@media (max-width: 960px) {
  .model-binding-drawer__grid,
  .model-binding-drawer__grid--compact {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
