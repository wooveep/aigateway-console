<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import PageSection from '@/components/common/PageSection.vue';
import ListToolbar from '@/components/common/ListToolbar.vue';
import DeleteConfirmModal from '@/components/common/DeleteConfirmModal.vue';
import StatusTag from '@/components/common/StatusTag.vue';
import { showSuccess } from '@/lib/feedback';
import {
  createWasmPlugin,
  deleteWasmPlugin,
  getDomainPluginInstance,
  getDomainPluginInstances,
  getGlobalPluginInstance,
  getRoutePluginInstance,
  getRoutePluginInstances,
  getWasmPlugins,
  getWasmPluginsConfig,
  updateDomainPluginInstance,
  updateGlobalPluginInstance,
  updateRoutePluginInstance,
  updateWasmPlugin,
} from '@/services/plugin';
import { getAiRoute, updateAiRoute } from '@/services/ai-route';
import { getGatewayRouteDetail, updateRouteConfig } from '@/services/route';
import { BUILTIN_ROUTE_PLUGIN_LIST } from '@/plugins/constants';
import {
  QueryType,
  filterVisiblePlugins,
  resolvePluginVisibilityScope,
} from '@/plugins/visibility';

const WasmPluginDrawer = defineAsyncComponent(() => import('@/features/plugin/WasmPluginDrawer.vue'));
const PluginConfigDrawer = defineAsyncComponent(() => import('@/features/plugin/PluginConfigDrawer.vue'));

const route = useRoute();
const router = useRouter();
const { locale } = useI18n();

const loading = ref(false);
const search = ref('');
const rows = ref<any[]>([]);
const targetDetail = ref<any>(null);

const wasmDrawerOpen = ref(false);
const configDrawerOpen = ref(false);
const deleteOpen = ref(false);
const configLoading = ref(false);
const instanceLoading = ref(false);

const editingWasm = ref<any>(null);
const deleting = ref<any>(null);
const configuring = ref<any>(null);
const currentConfigData = ref<any>(null);
const currentInstanceData = ref<any>(null);

const queryType = computed(() => String(route.query.type || ''));
const queryName = computed(() => String(route.query.name || ''));
const isTargetMode = computed(() => Boolean(queryType.value && queryName.value));
const visibilityScope = computed(() => resolvePluginVisibilityScope(queryType.value));

const backPath = computed(() => {
  if (queryType.value === QueryType.ROUTE) return '/route';
  if (queryType.value === QueryType.DOMAIN) return '/domain';
  if (queryType.value === QueryType.AI_ROUTE) return '/ai/route';
  return '';
});

const filteredRows = computed(() => rows.value.filter((item) => {
  const keyword = search.value.trim().toLowerCase();
  if (!keyword) {
    return true;
  }
  return [item.name, item.title, item.category, item.description]
    .some((value) => String(value || '').toLowerCase().includes(keyword));
}));

function isBuiltInPlugin(name: string) {
  return BUILTIN_ROUTE_PLUGIN_LIST.some((item) => item.key === name);
}

function getBuiltInEnabled(name: string) {
  if (name === 'rewrite') {
    return Boolean(targetDetail.value?.rewrite?.enabled);
  }
  if (name === 'headerModify') {
    return Boolean(targetDetail.value?.headerModify?.enabled || targetDetail.value?.headerControl?.enabled);
  }
  if (name === 'cors') {
    return Boolean(targetDetail.value?.cors?.enabled);
  }
  return Boolean(targetDetail.value?.retries?.enabled || targetDetail.value?.proxyNextUpstream?.enabled);
}

function getBuiltInRows() {
  if (queryType.value !== QueryType.ROUTE && queryType.value !== QueryType.AI_ROUTE) {
    return [];
  }
  const list = queryType.value === QueryType.AI_ROUTE
    ? BUILTIN_ROUTE_PLUGIN_LIST.filter((item) => item.enabledInAiRoute !== false)
    : BUILTIN_ROUTE_PLUGIN_LIST;

  return list.map((item) => ({
    ...item,
    name: item.key,
    enabled: getBuiltInEnabled(item.key),
    boundStatus: getBuiltInEnabled(item.key) ? '已绑定' : '未绑定',
  }));
}

async function loadTargetDetail() {
  if (!isTargetMode.value) {
    targetDetail.value = null;
    return;
  }
  if (queryType.value === QueryType.DOMAIN) {
    targetDetail.value = { name: queryName.value };
    return;
  }
  if (queryType.value === QueryType.AI_ROUTE) {
    targetDetail.value = await getAiRoute(queryName.value).catch(() => null);
    return;
  }
  targetDetail.value = await getGatewayRouteDetail(queryName.value).catch(() => null);
}

function getPluginTargetName() {
  if (queryType.value === QueryType.AI_ROUTE) {
    return `ai-route-${queryName.value}.internal`;
  }
  return queryName.value;
}

async function load() {
  loading.value = true;
  try {
    await loadTargetDetail();
    const plugins = await getWasmPlugins(locale.value).catch(() => []);
    const visiblePlugins = filterVisiblePlugins(plugins || [], visibilityScope.value);
    let merged = visiblePlugins;

    if (isTargetMode.value) {
      let enabledList: any[] = [];
      if (queryType.value === QueryType.DOMAIN) {
        enabledList = await getDomainPluginInstances(queryName.value).catch(() => []);
      } else {
        enabledList = await getRoutePluginInstances(getPluginTargetName()).catch(() => []);
      }

      merged = visiblePlugins.map((item: any) => {
        const enabledInstance = enabledList.find((plugin: any) => plugin.pluginName === item.name);
        return {
          ...item,
          enabled: Boolean(enabledInstance?.enabled),
          boundStatus: enabledInstance ? (enabledInstance.enabled ? '已绑定' : '已创建') : '未绑定',
        };
      });
      merged = [...getBuiltInRows(), ...merged];
    } else {
      merged = visiblePlugins.map((item: any) => ({
        ...item,
        boundStatus: item.internal ? '内置' : '可配置',
      }));
    }

    rows.value = merged;
  } finally {
    loading.value = false;
  }
}

function openWasmDrawer(record?: any) {
  editingWasm.value = record || null;
  wasmDrawerOpen.value = true;
}

async function submitWasm(payload: any, isEdit: boolean) {
  if (isEdit && editingWasm.value) {
    await updateWasmPlugin(editingWasm.value.name, payload);
  } else {
    await createWasmPlugin(payload);
  }
  wasmDrawerOpen.value = false;
  await load();
  showSuccess('插件已保存');
}

async function openConfig(record: any) {
  configuring.value = { ...record, queryType: queryType.value };
  currentConfigData.value = null;
  currentInstanceData.value = null;
  configDrawerOpen.value = true;

  if (record.builtIn) {
    return;
  }

  instanceLoading.value = true;
  configLoading.value = true;
  try {
    const [instanceData, configData] = await Promise.all([
      loadPluginInstance(record),
      getWasmPluginsConfig(record.name).catch(() => null),
    ]);
    currentInstanceData.value = instanceData;
    currentConfigData.value = configData;
  } finally {
    instanceLoading.value = false;
    configLoading.value = false;
  }
}

async function loadPluginInstance(record: any) {
  try {
    if (!isTargetMode.value) {
      return await getGlobalPluginInstance(record.name);
    }
    if (queryType.value === QueryType.DOMAIN) {
      return await getDomainPluginInstance({ name: queryName.value, pluginName: record.name });
    }
    return await getRoutePluginInstance({ name: getPluginTargetName(), pluginName: record.name });
  } catch {
    return null;
  }
}

async function submitBuiltIn(payload: Record<string, any>) {
  if (!configuring.value) {
    return;
  }

  const nextPayload = {
    ...(targetDetail.value || {}),
    ...payload,
  };

  if (queryType.value === QueryType.AI_ROUTE) {
    await updateAiRoute(nextPayload);
  } else {
    await updateRouteConfig(targetDetail.value.name, nextPayload);
  }

  configDrawerOpen.value = false;
  await load();
  showSuccess('配置已保存');
}

async function submitPlugin(payload: { enabled: boolean; rawConfigurations: string }) {
  if (!configuring.value) {
    return;
  }

  const nextPayload = {
    enabled: payload.enabled,
    pluginName: configuring.value.name,
    rawConfigurations: payload.rawConfigurations,
  };

  if (!isTargetMode.value) {
    await updateGlobalPluginInstance(configuring.value.name, nextPayload);
  } else if (queryType.value === QueryType.DOMAIN) {
    await updateDomainPluginInstance({ name: queryName.value, pluginName: configuring.value.name }, nextPayload);
  } else {
    await updateRoutePluginInstance({ name: getPluginTargetName(), pluginName: configuring.value.name }, nextPayload);
  }

  configDrawerOpen.value = false;
  await load();
  showSuccess('配置已保存');
}

async function confirmDelete() {
  if (!deleting.value) {
    return;
  }
  await deleteWasmPlugin(deleting.value.name);
  deleteOpen.value = false;
  await load();
  showSuccess('插件已删除');
}

watch(() => [route.fullPath, locale.value], load);
onMounted(load);
</script>

<template>
  <PageSection :title="isTargetMode ? `插件配置 · ${queryType} / ${queryName}` : '插件配置'">
    <ListToolbar
      v-model:search="search"
      search-placeholder="搜索插件名、标题、分类"
      :create-text="isTargetMode ? '' : '新增插件'"
      @refresh="load"
      @create="openWasmDrawer()"
    >
      <template #left>
        <a-button v-if="backPath" @click="router.push(backPath)">返回</a-button>
      </template>
    </ListToolbar>

    <a-table :data-source="filteredRows" :loading="loading" row-key="name" :scroll="{ x: 1180 }">
      <a-table-column key="name" data-index="name" title="插件名" width="220" />
      <a-table-column key="title" data-index="title" title="标题" width="180" />
      <a-table-column key="category" data-index="category" title="分类" width="120" />
      <a-table-column key="enabled" title="状态" width="120">
        <template #default="{ record }">
          <StatusTag :value="record.enabled ? 'enabled' : 'disabled'" />
        </template>
      </a-table-column>
      <a-table-column key="boundStatus" data-index="boundStatus" title="绑定状态" width="120" />
      <a-table-column key="description" data-index="description" title="描述" />
      <a-table-column key="actions" title="操作" width="260" fixed="right">
        <template #default="{ record }">
          <a-button type="link" size="small" @click="openConfig(record)">配置</a-button>
          <a-button
            v-if="!record.builtIn && !isTargetMode"
            type="link"
            size="small"
            @click="openWasmDrawer(record)"
          >
            编辑
          </a-button>
          <a-button
            v-if="!record.builtIn && !isTargetMode"
            type="link"
            size="small"
            danger
            @click="deleting = record; deleteOpen = true"
          >
            删除
          </a-button>
        </template>
      </a-table-column>
    </a-table>

    <WasmPluginDrawer
      v-model:open="wasmDrawerOpen"
      :record="editingWasm"
      @submit="submitWasm"
    />

    <PluginConfigDrawer
      v-model:open="configDrawerOpen"
      :record="configuring"
      :target-detail="targetDetail"
      :loading="configLoading"
      :instance-loading="instanceLoading"
      :config-data="currentConfigData"
      :instance-data="currentInstanceData"
      @submit-built-in="submitBuiltIn"
      @submit-plugin="submitPlugin"
    />

    <DeleteConfirmModal
      v-model:open="deleteOpen"
      title="删除插件"
      :content="deleting ? `确定删除插件 ${deleting.name} 吗？` : ''"
      @confirm="confirmDelete"
    />
  </PageSection>
</template>
