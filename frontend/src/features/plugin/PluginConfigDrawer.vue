<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import DrawerFooter from '@/components/common/DrawerFooter.vue';
import BuiltInPluginForm from './BuiltInPluginForm.vue';
import PluginSchemaEditor from './PluginSchemaEditor.vue';
import { showError, showWarning } from '@/lib/feedback';
import { BUILTIN_ROUTE_PLUGIN_LIST } from '@/plugins/constants';
import { PluginPhase } from '@/interfaces/wasm-plugin';
import {
  AI_DATA_MASKING_PLUGIN_NAME,
  cloneDeep,
  dumpYamlObject,
  getExampleRaw,
  omitAiDataMaskingManagedKeys,
  omitManagedSchema,
  parseYamlObject,
  resolvePluginSchema,
  sanitizeSchemaValue,
  validateSchemaValue,
} from './plugin-config';
import { QueryType } from '@/plugins/visibility';

const props = defineProps<{
  open: boolean;
  record?: any | null;
  targetDetail?: Record<string, any> | null;
  loading?: boolean;
  instanceLoading?: boolean;
  deleting?: boolean;
  allowDelete?: boolean;
  configData?: any;
  instanceData?: any;
}>();

const emit = defineEmits<{
  'update:open': [value: boolean];
  submitBuiltIn: [payload: Record<string, any>];
  submitPlugin: [payload: {
    enabled: boolean;
    rawConfigurations: string;
    phase: string;
    priority: number;
  }];
  deletePlugin: [];
}>();

const { locale, t } = useI18n();
const activeTab = ref<'form' | 'yaml'>('form');
const builtInRef = ref<InstanceType<typeof BuiltInPluginForm> | null>(null);

const schemaState = reactive<Record<string, any>>({});
const yamlState = ref('');
const enabledState = ref(false);
const phaseState = ref(PluginPhase.UNSPECIFIED);
const priorityState = ref(100);
const isRouteBuiltInPlugin = computed(() => BUILTIN_ROUTE_PLUGIN_LIST.some((item) => item.key === props.record?.name));
const requiredPlugins = computed(() => Array.isArray(props.record?.requiredPlugins) ? props.record.requiredPlugins : []);
const dependentBy = computed(() => Array.isArray(props.record?.dependentBy) ? props.record.dependentBy : []);
const isDisableLocked = computed(() => props.record?.canDisable === false && dependentBy.value.length > 0);

const currentConfigData = computed(() => {
  if (props.record?.name === AI_DATA_MASKING_PLUGIN_NAME) {
    return omitManagedSchema(props.configData);
  }
  return props.configData;
});

const isAiDataMaskingRouteBindingOnly = computed(() => (
  props.record?.name === AI_DATA_MASKING_PLUGIN_NAME
  && props.record?.queryType === QueryType.AI_ROUTE
));
const shouldUseBuiltInEditor = computed(() => (
  isRouteBuiltInPlugin.value && !isAiDataMaskingRouteBindingOnly.value
));
const currentSchema = computed(() => resolvePluginSchema(currentConfigData.value));
const canRenderSchemaForm = computed(() => (
  !isAiDataMaskingRouteBindingOnly.value
  && Boolean(currentSchema.value?.properties)
));

watch(
  () => [props.open, props.record, props.instanceData, currentConfigData.value],
  () => {
    if (!props.open || !props.record || shouldUseBuiltInEditor.value) {
      activeTab.value = 'form';
      return;
    }

    enabledState.value = Boolean(
      props.instanceData?.enabled
      ?? props.instanceData?.runtimeEnabled
      ?? props.record?.enabled
    );
    const nextPhase = String(props.record?.phase || PluginPhase.UNSPECIFIED);
    phaseState.value = Object.values(PluginPhase).includes(nextPhase as PluginPhase)
      ? nextPhase as PluginPhase
      : PluginPhase.UNSPECIFIED;
    priorityState.value = Number(props.record?.priority || 100);
    const exampleRaw = getExampleRaw(currentConfigData.value, !props.record?.queryType && props.record?.category === 'auth');
    const raw = props.instanceData?.rawConfigurations || exampleRaw || '';
    yamlState.value = raw;

    let nextSchema = {};
    try {
      nextSchema = parseYamlObject(raw);
    } catch {
      nextSchema = {};
    }
    Object.keys(schemaState).forEach((key) => delete schemaState[key]);
    Object.assign(schemaState, cloneDeep(nextSchema));
    activeTab.value = isAiDataMaskingRouteBindingOnly.value || canRenderSchemaForm.value ? 'form' : 'yaml';
  },
  { immediate: true, deep: true },
);

watch(activeTab, (nextTab) => {
  if (shouldUseBuiltInEditor.value) {
    return;
  }
  if (nextTab === 'yaml') {
    yamlState.value = dumpYamlObject(sanitizeSchemaValue(cloneDeep(schemaState)));
    return;
  }
  try {
    const parsed = parseYamlObject(yamlState.value);
    Object.keys(schemaState).forEach((key) => delete schemaState[key]);
    Object.assign(schemaState, cloneDeep(parsed));
  } catch {
        showWarning('YAML 解析失败，继续保留当前表单值');
  }
});

function close() {
  emit('update:open', false);
}

function syncYamlFromForm() {
  if (activeTab.value !== 'form' || shouldUseBuiltInEditor.value) {
    return;
  }
  yamlState.value = dumpYamlObject(sanitizeSchemaValue(cloneDeep(schemaState)));
}

watch(schemaState, syncYamlFromForm, { deep: true });

function submit() {
  if (!props.record) {
    return;
  }

  if (!enabledState.value && isDisableLocked.value) {
    showError(`当前插件仍被以下插件依赖，暂不允许关闭：${dependentBy.value.join('、')}`);
    return;
  }

  if (shouldUseBuiltInEditor.value) {
    const payload = builtInRef.value?.serialize?.();
    if (!payload) {
      return;
    }
    emit('submitBuiltIn', payload);
    return;
  }

  if (isAiDataMaskingRouteBindingOnly.value) {
    emit('submitPlugin', {
      enabled: enabledState.value,
      rawConfigurations: props.instanceData?.rawConfigurations || '',
      phase: phaseState.value,
      priority: Number(priorityState.value || 0),
    });
    return;
  }

  if (activeTab.value === 'form' && canRenderSchemaForm.value) {
    const errors = validateSchemaValue(currentSchema.value, schemaState, locale.value);
    if (errors.length) {
      showError(`请补全必填项：${errors[0]}`);
      return;
    }
    yamlState.value = dumpYamlObject(sanitizeSchemaValue(cloneDeep(schemaState)));
  } else {
    try {
      parseYamlObject(yamlState.value);
    } catch {
      showError('YAML 格式不正确');
      return;
    }
  }

  let rawConfigurations = yamlState.value;
  if (props.record.name === AI_DATA_MASKING_PLUGIN_NAME) {
    rawConfigurations = dumpYamlObject(omitAiDataMaskingManagedKeys(parseYamlObject(rawConfigurations)));
  }

  emit('submitPlugin', {
    enabled: enabledState.value,
    rawConfigurations,
    phase: phaseState.value,
    priority: Number(priorityState.value || 0),
  });
}
</script>

<template>
  <a-drawer
    :open="open"
    width="820"
    :title="record ? `配置 · ${record.title || record.name}` : '插件配置'"
    destroy-on-close
    @update:open="(value) => emit('update:open', value)"
  >
    <a-skeleton :loading="Boolean(loading || instanceLoading)" active>
      <div v-if="shouldUseBuiltInEditor">
        <BuiltInPluginForm
          ref="builtInRef"
          :plugin-name="record.name"
          :target-detail="targetDetail || null"
          :state="schemaState"
        />
      </div>

      <div v-else class="plugin-config-drawer">
        <a-alert
          v-if="record?.queryType === QueryType.AI_ROUTE && record?.enableStateText"
          type="info"
          show-icon
          :message="record.enableStateText"
        />
        <a-alert
          v-if="record?.queryType === QueryType.AI_ROUTE && requiredPlugins.length"
          type="info"
          show-icon
          :message="`启用当前插件时会自动启用依赖插件：${requiredPlugins.join('、')}`"
        />
        <a-alert
          v-if="record?.queryType === QueryType.AI_ROUTE && dependentBy.length"
          type="warning"
          show-icon
          :message="`当前插件被以下插件依赖：${dependentBy.join('、')}`"
        />
        <a-alert
          v-if="record?.name === AI_DATA_MASKING_PLUGIN_NAME"
          type="info"
          show-icon
          :message="isAiDataMaskingRouteBindingOnly
            ? 'AI脱敏规则在独立页面统一维护，这里只控制当前 AI 路由是否启用该插件。'
            : '托管敏感词规则不在此处编辑，这里只保留可直接下发的插件配置。'"
        />

        <a-form layout="vertical">
          <a-form-item label="启用状态">
            <a-switch v-model:checked="enabledState" :disabled="isDisableLocked && enabledState" />
          </a-form-item>
          <a-form-item v-if="record?.queryType === QueryType.AI_ROUTE" label="执行阶段">
            <a-select v-model:value="phaseState">
              <a-select-option :value="PluginPhase.UNSPECIFIED">{{ t('plugin.phases.unspecified') }}</a-select-option>
              <a-select-option :value="PluginPhase.AUTHN">{{ t('plugin.phases.authn') }}</a-select-option>
              <a-select-option :value="PluginPhase.AUTHZ">{{ t('plugin.phases.authz') }}</a-select-option>
              <a-select-option :value="PluginPhase.STATS">{{ t('plugin.phases.stats') }}</a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item v-if="record?.queryType === QueryType.AI_ROUTE" label="执行优先级">
            <a-input-number v-model:value="priorityState" style="width: 100%" :min="0" :max="1000" />
          </a-form-item>
        </a-form>

        <a-tabs v-if="!isAiDataMaskingRouteBindingOnly" v-model:activeKey="activeTab">
          <a-tab-pane key="form" tab="表单配置">
            <a-empty v-if="!canRenderSchemaForm" description="当前插件未提供可渲染的结构化 Schema，请切换到 YAML。" />
            <PluginSchemaEditor
              v-else
              :schema="currentSchema"
              :state="schemaState"
              :locale="locale"
              :allow-custom-fields="record?.name === AI_DATA_MASKING_PLUGIN_NAME"
            />
          </a-tab-pane>
          <a-tab-pane key="yaml" tab="YAML">
            <a-textarea v-model:value="yamlState" :rows="24" spellcheck="false" />
          </a-tab-pane>
        </a-tabs>
      </div>
    </a-skeleton>
    <DrawerFooter :loading="deleting" @cancel="close" @confirm="submit">
      <template v-if="allowDelete" #extra>
        <a-button danger :loading="deleting" @click="emit('deletePlugin')">删除当前绑定</a-button>
      </template>
    </DrawerFooter>
  </a-drawer>
</template>

<style scoped>
.plugin-config-drawer {
  display: grid;
  gap: 14px;
}
</style>
