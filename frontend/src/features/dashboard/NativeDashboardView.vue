<script setup lang="ts">
import { computed, onBeforeUnmount, shallowRef, watch } from 'vue';
import { ReloadOutlined } from '@ant-design/icons-vue';
import NativeDashboardPanelCard from '@/features/dashboard/NativeDashboardPanelCard.vue';
import { DEFAULT_RANGE_MS, panelHasData, RANGE_OPTIONS, REFRESH_OPTIONS } from '@/features/dashboard/dashboard-native';
import { DashboardType, type NativeDashboardData } from '@/interfaces/dashboard';
import { getNativeDashboard } from '@/services/dashboard';
import { formatDateTimeDisplay } from '@/utils/time';
import { useI18n } from 'vue-i18n';

const props = defineProps<{
  type: DashboardType;
}>();

const { t } = useI18n();

const rangeMs = shallowRef(DEFAULT_RANGE_MS);
const refreshMs = shallowRef(30 * 1000);
const loading = shallowRef(false);
const errorMessage = shallowRef('');
const lastUpdated = shallowRef('');
const data = shallowRef<NativeDashboardData | null>(null);
const activeRows = shallowRef<string[]>([]);

let refreshTimer: number | null = null;

const hasAnyData = computed(() => data.value?.rows.some((row) => row.panels.some((panel) => panelHasData(panel))) ?? false);

async function load() {
  loading.value = true;
  errorMessage.value = '';
  try {
    const to = Date.now();
    const result = await getNativeDashboard(props.type, {
      from: to - rangeMs.value,
      to,
    });
    data.value = result;
    lastUpdated.value = formatDateTimeDisplay(Date.now());
    if (!activeRows.value.length) {
      activeRows.value = result.rows.filter((row) => !row.collapsed).map((row) => row.title);
    }
  } catch (error: any) {
    data.value = null;
    errorMessage.value = String(error?.response?.data?.message || error?.message || t('dashboard.loadFailed'));
  } finally {
    loading.value = false;
  }
}

function setupAutoRefresh() {
  if (refreshTimer) {
    window.clearInterval(refreshTimer);
    refreshTimer = null;
  }
  if (refreshMs.value > 0) {
    refreshTimer = window.setInterval(() => {
      void load();
    }, refreshMs.value);
  }
}

function translateText(group: 'titles' | 'series' | 'columns', value?: string) {
  if (!value) {
    return value || '';
  }
  const key = `dashboard.native.${group}.${value}`;
  const translated = t(key);
  return translated === key ? value : translated;
}

watch(() => props.type, () => {
  activeRows.value = [];
  void load();
}, { immediate: true });

watch(rangeMs, () => {
  void load();
});

watch(refreshMs, setupAutoRefresh, { immediate: true });

onBeforeUnmount(() => {
  if (refreshTimer) {
    window.clearInterval(refreshTimer);
  }
});
</script>

<template>
  <div class="native-dashboard">
    <div class="native-dashboard__toolbar">
      <div class="native-dashboard__controls">
        <div class="native-dashboard__control">
          <span class="native-dashboard__label">{{ t('dashboard.native.range') }}</span>
          <a-select v-model:value="rangeMs" :options="RANGE_OPTIONS.map((option) => ({ value: option, label: t(`dashboard.native.rangeOptions.${option}`) }))" />
        </div>
        <div class="native-dashboard__control">
          <span class="native-dashboard__label">{{ t('dashboard.native.refreshEvery') }}</span>
          <a-select v-model:value="refreshMs" :options="REFRESH_OPTIONS.map((option) => ({ value: option, label: t(`dashboard.native.refreshOptions.${option}`) }))" />
        </div>
      </div>
      <div class="native-dashboard__actions">
        <span class="native-dashboard__updated" v-if="lastUpdated">
          {{ t('dashboard.native.lastUpdated', { time: lastUpdated }) }}
        </span>
        <a-button @click="load">
          <template #icon>
            <ReloadOutlined />
          </template>
          {{ t('dashboard.native.refresh') }}
        </a-button>
      </div>
    </div>

    <a-skeleton v-if="loading && !data" active />

    <a-alert
      v-else-if="errorMessage"
      type="warning"
      show-icon
      :message="errorMessage"
    />

    <a-empty v-else-if="!hasAnyData" :description="t('dashboard.native.noData')" />

    <a-collapse
      v-else
      class="native-dashboard__collapse"
      v-model:active-key="activeRows"
    >
      <a-collapse-panel
        v-for="row in data?.rows || []"
        :key="row.title"
        :header="t(`dashboard.native.rows.${row.title}`) === `dashboard.native.rows.${row.title}` ? row.title : t(`dashboard.native.rows.${row.title}`)"
      >
        <div class="native-dashboard__grid">
          <div
            v-for="panel in row.panels"
            :key="panel.id"
            class="native-dashboard__panel-cell"
            :style="{ gridColumn: `${panel.gridPos.x + 1} / span ${Math.max(1, panel.gridPos.w)}` }"
          >
            <NativeDashboardPanelCard :panel="panel" :range-ms="rangeMs" :translate-text="translateText" />
          </div>
        </div>
      </a-collapse-panel>
    </a-collapse>
  </div>
</template>

<style scoped>
.native-dashboard {
  display: grid;
  gap: 16px;
}

.native-dashboard__toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.native-dashboard__controls,
.native-dashboard__actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.native-dashboard__control {
  display: grid;
  gap: 6px;
  min-width: 180px;
}

.native-dashboard__label,
.native-dashboard__updated {
  color: var(--portal-text-soft);
  font-size: 12px;
}

.native-dashboard__collapse :deep(.ant-collapse-item) {
  border-radius: 18px;
  border: 1px solid var(--portal-border);
  overflow: hidden;
  background: rgba(255, 255, 255, 0.9);
}

.native-dashboard__grid {
  display: grid;
  grid-template-columns: repeat(24, minmax(0, 1fr));
  gap: 14px;
}

.native-dashboard__panel-cell {
  min-width: 0;
}

@media (max-width: 1023px) {
  .native-dashboard__grid {
    grid-template-columns: repeat(12, minmax(0, 1fr));
  }
}

@media (max-width: 767px) {
  .native-dashboard__grid {
    grid-template-columns: 1fr;
  }

  .native-dashboard__panel-cell {
    grid-column: 1 / -1 !important;
  }
}
</style>
