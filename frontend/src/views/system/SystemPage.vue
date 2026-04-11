<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import PageSection from '@/components/common/PageSection.vue';
import { getAIGatewayConfig, getSystemInfo, updateAIGatewayConfig } from '@/services/system';
import { showSuccess } from '@/lib/feedback';

const { t } = useI18n();
const loading = ref(false);
const saving = ref(false);
const systemInfo = ref<Record<string, any>>({});
const configText = ref('');

async function load() {
  loading.value = true;
  try {
    const [info, config] = await Promise.all([
      getSystemInfo().catch(() => ({})),
      getAIGatewayConfig().catch(() => ''),
    ]);
    systemInfo.value = info || {};
    configText.value = typeof config === 'string' ? config : JSON.stringify(config, null, 2);
  } finally {
    loading.value = false;
  }
}

async function saveConfig() {
  saving.value = true;
  try {
    await updateAIGatewayConfig(configText.value);
    showSuccess(t('misc.save'));
  } finally {
    saving.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="system-page">
    <PageSection :title="t('menu.systemSettings')">
      <template #actions>
        <a-button @click="load">{{ t('misc.refresh') }}</a-button>
      </template>
      <a-skeleton v-if="loading" active />
      <div v-else class="system-page__overview">
        <article
          v-for="(value, key) in systemInfo"
          :key="key"
          class="system-page__stat"
        >
          <span>{{ key }}</span>
          <strong>{{ typeof value === 'object' ? JSON.stringify(value) : String(value) }}</strong>
        </article>
      </div>
    </PageSection>

    <PageSection title="AIGateway Config">
      <a-textarea v-model:value="configText" :rows="24" spellcheck="false" />
      <div class="system-page__actions">
        <a-button @click="load">{{ t('misc.refresh') }}</a-button>
        <a-button type="primary" :loading="saving" @click="saveConfig">{{ t('misc.save') }}</a-button>
      </div>
    </PageSection>
  </div>
</template>

<style scoped>
.system-page {
  display: grid;
  gap: 18px;
}

.system-page__overview {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.system-page__stat {
  min-width: 0;
  padding: 16px;
  border: 1px solid var(--portal-border);
  border-radius: 16px;
  background: var(--portal-surface-soft);
}

.system-page__stat span {
  display: block;
  margin-bottom: 8px;
  color: var(--portal-text-soft);
  font-size: 12px;
}

.system-page__stat strong {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
}

.system-page__actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 18px;
}

@media (max-width: 1023px) {
  .system-page__overview {
    grid-template-columns: 1fr;
  }
}
</style>
