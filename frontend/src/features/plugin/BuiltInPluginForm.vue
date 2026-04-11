<script setup lang="ts">
import { computed, watch } from 'vue';
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons-vue';
import type { HeaderModifyConfig } from '@/interfaces/route';
import { useI18n } from 'vue-i18n';

type HeaderRow = {
  headerType: 'request' | 'response';
  actionType: 'add' | 'set' | 'remove';
  key: string;
  value?: string;
};

const props = defineProps<{
  pluginName: string;
  targetDetail: Record<string, any> | null;
  state: Record<string, any>;
}>();

const { t } = useI18n();

function normalizeRewriteSource(source: any) {
  return {
    enabled: Boolean(source?.enabled),
    path: source?.path || '',
    host: source?.host || '',
    matchType: props.targetDetail?.path?.matchType || props.targetDetail?.pathPredicate?.matchType || '',
    originPath: props.targetDetail?.path?.matchValue || props.targetDetail?.pathPredicate?.matchValue || '',
    originHost: Array.isArray(props.targetDetail?.domains) ? props.targetDetail?.domains.join(', ') : '',
  };
}

function headerRowsFromConfig(config?: HeaderModifyConfig | any): HeaderRow[] {
  const source = config || {};
  const rows: HeaderRow[] = [];
  ['request', 'response'].forEach((headerType) => {
    const stage = source?.[headerType] || {};
    (stage.add || []).forEach((item: any) => rows.push({ headerType: headerType as HeaderRow['headerType'], actionType: 'add', key: item.key, value: item.value }));
    (stage.set || []).forEach((item: any) => rows.push({ headerType: headerType as HeaderRow['headerType'], actionType: 'set', key: item.key, value: item.value }));
    (stage.remove || stage.delete || []).forEach((item: string) => rows.push({ headerType: headerType as HeaderRow['headerType'], actionType: 'remove', key: item, value: '' }));
  });
  return rows.length ? rows : [{ headerType: 'request', actionType: 'add', key: '', value: '' }];
}

function normalizeHeaderSource(source?: any) {
  return {
    enabled: Boolean(source?.enabled),
    rows: headerRowsFromConfig(source),
  };
}

function normalizeCorsSource(source?: any) {
  return {
    enabled: Boolean(source?.enabled),
    allowOrigins: Array.isArray(source?.allowOrigins) ? source.allowOrigins.join(';') : '*',
    allowMethods: Array.isArray(source?.allowMethods) ? source.allowMethods : ['GET', 'PUT', 'POST', 'HEAD', 'DELETE', 'PATCH', 'OPTIONS'],
    allowHeaders: Array.isArray(source?.allowHeaders) ? source.allowHeaders.join(';') : '*',
    exposeHeaders: Array.isArray(source?.exposeHeaders || source?.exposeHeader) ? (source.exposeHeaders || source.exposeHeader).join(';') : '*',
    allowCredentials: Boolean(source?.allowCredentials),
    maxAge: source?.maxAge ?? source?.maxAgent ?? 86400,
  };
}

function normalizeRetriesSource(source?: any) {
  const retrySource = source || {};
  return {
    enabled: Boolean(retrySource?.enabled),
    attempts: retrySource?.attempts ?? retrySource?.attempt ?? 3,
    conditions: retrySource?.conditions || (retrySource?.retryOn ? String(retrySource.retryOn).split(',') : ['error', 'timeout']),
    timeout: retrySource?.timeout ?? 5,
  };
}

function syncState() {
  if (props.pluginName === 'rewrite') {
    Object.assign(props.state, normalizeRewriteSource(props.targetDetail?.rewrite));
    return;
  }
  if (props.pluginName === 'headerModify') {
    Object.assign(props.state, normalizeHeaderSource(props.targetDetail?.headerModify || props.targetDetail?.headerControl));
    return;
  }
  if (props.pluginName === 'cors') {
    Object.assign(props.state, normalizeCorsSource(props.targetDetail?.cors));
    return;
  }
  if (props.pluginName === 'retries') {
    Object.assign(props.state, normalizeRetriesSource(props.targetDetail?.proxyNextUpstream || props.targetDetail?.retries));
  }
}

watch(() => [props.pluginName, props.targetDetail], syncState, { immediate: true, deep: true });

const methodOptions = ['GET', 'POST', 'PUT', 'DELETE', 'HEAD', 'OPTIONS', 'PATCH'];

const showHeaderRows = computed(() => Array.isArray(props.state.rows) ? props.state.rows : []);

function addHeaderRow() {
  if (!Array.isArray(props.state.rows)) {
    props.state.rows = [];
  }
  props.state.rows.push({ headerType: 'request', actionType: 'add', key: '', value: '' });
}

function removeHeaderRow(index: number) {
  props.state.rows.splice(index, 1);
}

defineExpose({
  serialize() {
    if (props.pluginName === 'rewrite') {
      return {
        rewrite: {
          enabled: Boolean(props.state.enabled),
          path: props.state.path || '',
          host: props.state.host || '',
        },
      };
    }

    if (props.pluginName === 'headerModify') {
      const headerModify = {
        enabled: Boolean(props.state.enabled),
        request: { add: [] as any[], set: [] as any[], remove: [] as string[] },
        response: { add: [] as any[], set: [] as any[], remove: [] as string[] },
      };
      showHeaderRows.value.forEach((item) => {
        if (!item.key) {
          return;
        }
        const target = headerModify[item.headerType][item.actionType];
        if (item.actionType === 'remove') {
          target.push(item.key);
          return;
        }
        target.push({
          key: item.key,
          value: item.value || '',
        });
      });
      return {
        headerModify,
        headerControl: headerModify,
      };
    }

    if (props.pluginName === 'cors') {
      return {
        cors: {
          enabled: Boolean(props.state.enabled),
          allowOrigins: String(props.state.allowOrigins || '').split(';').map((item) => item.trim()).filter(Boolean),
          allowMethods: Array.isArray(props.state.allowMethods) ? props.state.allowMethods : [],
          allowHeaders: String(props.state.allowHeaders || '').split(';').map((item) => item.trim()).filter(Boolean),
          exposeHeaders: String(props.state.exposeHeaders || '').split(';').map((item) => item.trim()).filter(Boolean),
          allowCredentials: Boolean(props.state.allowCredentials),
          maxAge: props.state.maxAge || 86400,
        },
      };
    }

    const retries = {
      enabled: Boolean(props.state.enabled),
      attempts: props.state.attempts || 3,
      conditions: Array.isArray(props.state.conditions) ? props.state.conditions : [],
      timeout: props.state.timeout || 5,
    };

    return {
      retries,
      proxyNextUpstream: retries,
    };
  },
});
</script>

<template>
  <div class="built-in-plugin-form">
    <template v-if="pluginName === 'rewrite'">
      <a-form layout="vertical">
        <a-form-item :label="t('plugins.configForm.enableStatus')">
          <a-switch v-model:checked="state.enabled" />
        </a-form-item>
        <div class="built-in-plugin-form__grid">
          <article class="built-in-plugin-form__card">
            <div class="built-in-plugin-form__card-title">{{ t('plugins.builtIns.rewrite.originalPath') }}</div>
            <div class="built-in-plugin-form__static">{{ state.matchType || '-' }} {{ state.originPath || '-' }}</div>
          </article>
          <article class="built-in-plugin-form__card">
            <div class="built-in-plugin-form__card-title">{{ t('plugins.builtIns.rewrite.rewritePath') }}</div>
            <a-input v-model:value="state.path" :placeholder="t('plugins.builtIns.rewrite.rewritePathPlaceholder')" />
          </article>
          <article class="built-in-plugin-form__card">
            <div class="built-in-plugin-form__card-title">{{ t('plugins.builtIns.rewrite.originalHost') }}</div>
            <div class="built-in-plugin-form__static">{{ state.originHost || '-' }}</div>
          </article>
          <article class="built-in-plugin-form__card">
            <div class="built-in-plugin-form__card-title">{{ t('plugins.builtIns.rewrite.rewriteHost') }}</div>
            <a-input v-model:value="state.host" :placeholder="t('plugins.builtIns.rewrite.rewriteHostPlaceholder')" />
          </article>
        </div>
      </a-form>
    </template>

    <template v-else-if="pluginName === 'headerModify'">
      <a-form layout="vertical">
        <div class="built-in-plugin-form__header">
          <a-form-item :label="t('plugins.configForm.enableStatus')">
            <a-switch v-model:checked="state.enabled" />
          </a-form-item>
          <a-button type="link" @click="addHeaderRow">
            <template #icon><PlusOutlined /></template>
            {{ t('plugins.builtIns.headerControl.addNewRule') }}
          </a-button>
        </div>

        <div class="built-in-plugin-form__rows">
          <article v-for="(row, index) in showHeaderRows" :key="`${row.headerType}-${index}`" class="built-in-plugin-form__row">
            <a-select v-model:value="row.headerType" style="width: 132px">
              <a-select-option value="request">{{ t('plugins.builtIns.headerControl.request') }}</a-select-option>
              <a-select-option value="response">{{ t('plugins.builtIns.headerControl.response') }}</a-select-option>
            </a-select>
            <a-select v-model:value="row.actionType" style="width: 128px">
              <a-select-option value="add">{{ t('plugins.builtIns.headerControl.add') }}</a-select-option>
              <a-select-option value="set">{{ t('plugins.builtIns.headerControl.set') }}</a-select-option>
              <a-select-option value="remove">{{ t('plugins.builtIns.headerControl.remove') }}</a-select-option>
            </a-select>
            <a-input v-model:value="row.key" placeholder="Header Key" />
            <a-input v-if="row.actionType !== 'remove'" v-model:value="row.value" placeholder="Header Value" />
            <div v-else class="built-in-plugin-form__empty">删除项不需要值</div>
            <a-button type="text" danger @click="removeHeaderRow(index)">
              <template #icon><MinusCircleOutlined /></template>
            </a-button>
          </article>
        </div>
      </a-form>
    </template>

    <template v-else-if="pluginName === 'cors'">
      <a-form layout="vertical">
        <a-form-item :label="t('plugins.configForm.enableStatus')">
          <a-switch v-model:checked="state.enabled" />
        </a-form-item>
        <a-form-item :label="t('plugins.builtIns.cors.allowOrigins')">
          <a-textarea v-model:value="state.allowOrigins" :rows="3" />
        </a-form-item>
        <a-form-item :label="t('plugins.builtIns.cors.allowMethods')">
          <a-checkbox-group v-model:value="state.allowMethods" :options="methodOptions" />
        </a-form-item>
        <a-form-item :label="t('plugins.builtIns.cors.allowHeaders')">
          <a-textarea v-model:value="state.allowHeaders" :rows="3" />
        </a-form-item>
        <a-form-item :label="t('plugins.builtIns.cors.exposeHeaders')">
          <a-textarea v-model:value="state.exposeHeaders" :rows="3" />
        </a-form-item>
        <div class="built-in-plugin-form__grid">
          <a-form-item :label="t('plugins.builtIns.cors.allowCredentials')">
            <a-switch v-model:checked="state.allowCredentials" />
          </a-form-item>
          <a-form-item :label="t('plugins.builtIns.cors.maxAge')">
            <a-input-number v-model:value="state.maxAge" style="width: 100%" />
          </a-form-item>
        </div>
      </a-form>
    </template>

    <template v-else>
      <a-form layout="vertical">
        <a-form-item :label="t('plugins.configForm.enableStatus')">
          <a-switch v-model:checked="state.enabled" />
        </a-form-item>
        <div class="built-in-plugin-form__grid">
          <a-form-item :label="t('plugins.builtIns.retries.attempts')">
            <a-input-number v-model:value="state.attempts" style="width: 100%" :min="1" :max="10" />
          </a-form-item>
          <a-form-item :label="t('plugins.builtIns.retries.timeout')">
            <a-input-number v-model:value="state.timeout" style="width: 100%" :min="1" :max="600" />
          </a-form-item>
        </div>
        <a-form-item :label="t('plugins.builtIns.retries.conditions')">
          <a-select v-model:value="state.conditions" mode="multiple">
            <a-select-option value="error">{{ t('plugins.builtIns.retries.condition.error') }}</a-select-option>
            <a-select-option value="timeout">{{ t('plugins.builtIns.retries.condition.timeout') }}</a-select-option>
            <a-select-option value="non_idempotent">{{ t('plugins.builtIns.retries.condition.non_idempotent') }}</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </template>
  </div>
</template>

<style scoped>
.built-in-plugin-form {
  display: grid;
  gap: 16px;
}

.built-in-plugin-form__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.built-in-plugin-form__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.built-in-plugin-form__card,
.built-in-plugin-form__row {
  padding: 14px;
  border: 1px solid var(--portal-border);
  border-radius: 14px;
  background: var(--portal-surface);
}

.built-in-plugin-form__card-title {
  margin-bottom: 8px;
  font-size: 12px;
  color: var(--portal-text-soft);
}

.built-in-plugin-form__static {
  color: var(--portal-text);
  font-weight: 600;
}

.built-in-plugin-form__rows {
  display: grid;
  gap: 10px;
}

.built-in-plugin-form__row {
  display: grid;
  grid-template-columns: 132px 128px minmax(0, 1fr) minmax(0, 1fr) auto;
  gap: 8px;
  align-items: center;
}

.built-in-plugin-form__empty {
  color: var(--portal-text-soft);
  font-size: 12px;
}

@media (max-width: 960px) {
  .built-in-plugin-form__grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .built-in-plugin-form__row {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
