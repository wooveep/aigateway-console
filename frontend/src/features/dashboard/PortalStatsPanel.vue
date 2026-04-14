<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue';
import PortalUnavailableState from '@/components/common/PortalUnavailableState.vue';
import { usePortalAvailability } from '@/composables/usePortalAvailability';
import type {
  PortalDepartmentBillRecord,
  PortalUsageEventRecord,
  PortalUsageStatRecord,
} from '@/interfaces/portal-stats';
import {
  getPortalDepartmentBills,
  getPortalUsageEvents,
  getPortalUsageStats,
} from '@/services/portal-stats';
import { formatDateTimeDisplay } from '@/utils/time';

const { portalUnavailable } = usePortalAvailability();

const activeTab = ref('usage');
const loading = ref(false);
const usageRows = ref<PortalUsageStatRecord[]>([]);
const usageEventRows = ref<PortalUsageEventRecord[]>([]);
const departmentBillRows = ref<PortalDepartmentBillRecord[]>([]);

const usageEventsQuery = reactive({
  consumerName: '',
  departmentId: '',
  apiKeyId: '',
  modelId: '',
  routeName: '',
  requestStatus: '',
  usageStatus: '',
  includeChildren: true,
  pageNum: 1,
  pageSize: 50,
});

const departmentBillsQuery = reactive({
  departmentId: '',
  includeChildren: true,
});

async function loadUsage() {
  usageRows.value = await getPortalUsageStats().catch(() => []);
}

async function loadUsageEvents() {
  usageEventRows.value = await getPortalUsageEvents(usageEventsQuery).catch(() => []);
}

async function loadDepartmentBills() {
  departmentBillRows.value = await getPortalDepartmentBills(departmentBillsQuery).catch(() => []);
}

async function loadActiveTab() {
  if (portalUnavailable.value) {
    usageRows.value = [];
    usageEventRows.value = [];
    departmentBillRows.value = [];
    return;
  }
  loading.value = true;
  try {
    if (activeTab.value === 'usage') {
      await loadUsage();
      return;
    }
    if (activeTab.value === 'usage-events') {
      await loadUsageEvents();
      return;
    }
    await loadDepartmentBills();
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void loadActiveTab();
});

watch(activeTab, () => {
  void loadActiveTab();
});
</script>

<template>
  <div class="portal-stats-panel">
    <PortalUnavailableState v-if="portalUnavailable" />

    <template v-else>
      <div class="portal-stats-panel__actions">
        <a-button @click="loadActiveTab">刷新</a-button>
      </div>

      <a-tabs v-model:activeKey="activeTab">
        <a-tab-pane key="usage" tab="Usage">
          <a-table :data-source="usageRows" :loading="loading" row-key="consumerName" :scroll="{ x: 1180 }" size="small">
            <a-table-column key="consumerName" data-index="consumerName" title="Consumer" width="180" />
            <a-table-column key="modelName" data-index="modelName" title="Model" width="180" />
            <a-table-column key="requestCount" data-index="requestCount" title="Requests" width="120" />
            <a-table-column key="inputTokens" data-index="inputTokens" title="Input Tokens" width="120" />
            <a-table-column key="outputTokens" data-index="outputTokens" title="Output Tokens" width="120" />
            <a-table-column key="totalTokens" data-index="totalTokens" title="Total Tokens" width="120" />
            <a-table-column key="cacheReadInputTokens" data-index="cacheReadInputTokens" title="Cache Read" width="120" />
            <a-table-column key="inputImageCount" data-index="inputImageCount" title="Input Images" width="120" />
            <a-table-column key="outputImageCount" data-index="outputImageCount" title="Output Images" width="120" />
          </a-table>
        </a-tab-pane>

        <a-tab-pane key="usage-events" tab="Usage Events">
          <a-form layout="inline" class="portal-stats-panel__filters">
            <a-form-item label="Consumer"><a-input v-model:value="usageEventsQuery.consumerName" allow-clear /></a-form-item>
            <a-form-item label="Department"><a-input v-model:value="usageEventsQuery.departmentId" allow-clear /></a-form-item>
            <a-form-item label="API Key"><a-input v-model:value="usageEventsQuery.apiKeyId" allow-clear /></a-form-item>
            <a-form-item label="Model"><a-input v-model:value="usageEventsQuery.modelId" allow-clear /></a-form-item>
            <a-form-item label="Route"><a-input v-model:value="usageEventsQuery.routeName" allow-clear /></a-form-item>
            <a-form-item label="Request Status"><a-input v-model:value="usageEventsQuery.requestStatus" allow-clear /></a-form-item>
            <a-form-item label="Usage Status"><a-input v-model:value="usageEventsQuery.usageStatus" allow-clear /></a-form-item>
            <a-form-item label="Include Children"><a-switch v-model:checked="usageEventsQuery.includeChildren" /></a-form-item>
            <a-form-item><a-button type="primary" @click="loadActiveTab">查询</a-button></a-form-item>
          </a-form>

          <a-table :data-source="usageEventRows" :loading="loading" row-key="eventId" :scroll="{ x: 1800 }" size="small">
            <a-table-column key="occurredAt" title="Occurred At" width="180">
              <template #default="{ record }">{{ formatDateTimeDisplay(record.occurredAt) }}</template>
            </a-table-column>
            <a-table-column key="consumerName" data-index="consumerName" title="Consumer" width="160" />
            <a-table-column key="departmentPath" data-index="departmentPath" title="Department Path" width="220" />
            <a-table-column key="apiKeyId" data-index="apiKeyId" title="API Key" width="180" />
            <a-table-column key="modelId" data-index="modelId" title="Model" width="180" />
            <a-table-column key="routeName" data-index="routeName" title="Route" width="180" />
            <a-table-column key="requestStatus" data-index="requestStatus" title="Request Status" width="140" />
            <a-table-column key="usageStatus" data-index="usageStatus" title="Usage Status" width="140" />
            <a-table-column key="requestCount" data-index="requestCount" title="Requests" width="110" />
            <a-table-column key="totalTokens" data-index="totalTokens" title="Total Tokens" width="120" />
            <a-table-column key="costMicroYuan" data-index="costMicroYuan" title="Cost(μ¥)" width="120" />
          </a-table>
        </a-tab-pane>

        <a-tab-pane key="department-bills" tab="Department Bills">
          <a-form layout="inline" class="portal-stats-panel__filters">
            <a-form-item label="Department"><a-input v-model:value="departmentBillsQuery.departmentId" allow-clear /></a-form-item>
            <a-form-item label="Include Children"><a-switch v-model:checked="departmentBillsQuery.includeChildren" /></a-form-item>
            <a-form-item><a-button type="primary" @click="loadActiveTab">查询</a-button></a-form-item>
          </a-form>

          <a-table :data-source="departmentBillRows" :loading="loading" row-key="departmentId" :scroll="{ x: 980 }" size="small">
            <a-table-column key="departmentId" data-index="departmentId" title="Department ID" width="180" />
            <a-table-column key="departmentName" data-index="departmentName" title="Department" width="180" />
            <a-table-column key="departmentPath" data-index="departmentPath" title="Department Path" width="240" />
            <a-table-column key="requestCount" data-index="requestCount" title="Requests" width="120" />
            <a-table-column key="totalTokens" data-index="totalTokens" title="Total Tokens" width="140" />
            <a-table-column key="totalCost" data-index="totalCost" title="Total Cost" width="120" />
            <a-table-column key="activeConsumers" data-index="activeConsumers" title="Active Consumers" width="140" />
          </a-table>
        </a-tab-pane>
      </a-tabs>
    </template>
  </div>
</template>

<style scoped>
.portal-stats-panel {
  display: grid;
  gap: 16px;
}

.portal-stats-panel__actions {
  display: flex;
  justify-content: flex-end;
}

.portal-stats-panel__filters {
  margin-bottom: 16px;
  gap: 8px 0;
}
</style>
