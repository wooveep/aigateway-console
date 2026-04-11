<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import PageSection from '@/components/common/PageSection.vue';
import ListToolbar from '@/components/common/ListToolbar.vue';
import DrawerFooter from '@/components/common/DrawerFooter.vue';
import DeleteConfirmModal from '@/components/common/DeleteConfirmModal.vue';
import StrategyLink from '@/components/common/StrategyLink.vue';
import { addGatewayRouteCompat, deleteGatewayRouteCompat, getGatewayRoutesCompat, updateGatewayRouteCompat } from '@/services/route-compat';
import { safeParseJson, splitLines, stringifyPretty } from '@/lib/portal';
import { showSuccess } from '@/lib/feedback';

const loading = ref(false);
const search = ref('');
const rows = ref<any[]>([]);
const drawerOpen = ref(false);
const deleteOpen = ref(false);
const editing = ref<any>(null);
const deleting = ref<any>(null);

const formState = reactive({
  name: '',
  domainsText: '',
  methodsText: 'GET\nPOST',
  pathMatchType: 'PRE',
  pathMatchValue: '/',
  servicesJson: '[\n  {\n    "name": "",\n    "weight": 100\n  }\n]',
  allowedConsumerLevels: ['normal'] as string[],
});

const filtered = computed(() => rows.value.filter((item) => {
  const keyword = search.value.trim().toLowerCase();
  if (!keyword) {
    return true;
  }
  return [item.name, (item.domains || []).join(',')].some((value) => String(value || '').toLowerCase().includes(keyword));
}));

async function load() {
  loading.value = true;
  try {
    const result = await getGatewayRoutesCompat().catch(() => ({ data: [] }));
    rows.value = Array.isArray(result) ? result : (result.data || []);
  } finally {
    loading.value = false;
  }
}

function openDrawer(record?: any) {
  editing.value = record || null;
  Object.assign(formState, {
    name: record?.name || '',
    domainsText: (record?.domains || []).join('\n'),
    methodsText: (record?.methods || ['GET', 'POST']).join('\n'),
    pathMatchType: record?.path?.matchType || 'PRE',
    pathMatchValue: record?.path?.matchValue || '/',
    servicesJson: stringifyPretty(record?.services || []),
    allowedConsumerLevels: record?.authConfig?.allowedConsumerLevels || ['normal'],
  });
  drawerOpen.value = true;
}

async function submit() {
  const payload = {
    ...editing.value,
    name: formState.name,
    domains: splitLines(formState.domainsText),
    methods: splitLines(formState.methodsText),
    path: {
      matchType: formState.pathMatchType,
      matchValue: formState.pathMatchValue,
    },
    services: safeParseJson(formState.servicesJson, []),
    authConfig: {
      enabled: formState.allowedConsumerLevels.length > 0,
      allowedConsumerLevels: formState.allowedConsumerLevels,
    },
  };
  if (editing.value) {
    await updateGatewayRouteCompat(payload as any);
  } else {
    await addGatewayRouteCompat(payload as any);
  }
  drawerOpen.value = false;
  await load();
  showSuccess('保存成功');
}

async function confirmDelete() {
  if (!deleting.value) {
    return;
  }
  await deleteGatewayRouteCompat(deleting.value.name);
  deleteOpen.value = false;
  await load();
  showSuccess('删除成功');
}

onMounted(load);
</script>

<template>
  <PageSection title="路由配置">
    <ListToolbar v-model:search="search" search-placeholder="搜索路由名或域名" create-text="新增路由" @refresh="load" @create="openDrawer()" />
    <a-table :data-source="filtered" :loading="loading" row-key="name" :scroll="{ x: 980 }">
      <a-table-column key="name" data-index="name" title="名称" />
      <a-table-column key="domains" title="域名">
        <template #default="{ record }">{{ (record.domains || []).join(', ') || '-' }}</template>
      </a-table-column>
      <a-table-column key="path" title="路径匹配">
        <template #default="{ record }">{{ record.path?.matchType }} | {{ record.path?.matchValue }}</template>
      </a-table-column>
      <a-table-column key="services" title="目标服务">
        <template #default="{ record }">{{ (record.services || []).map((item: any) => item.port ? `${item.name}:${item.port}` : item.name).join(', ') || '-' }}</template>
      </a-table-column>
      <a-table-column key="actions" title="操作" width="240">
        <template #default="{ record }">
          <StrategyLink :path="`/route/config?type=route&name=${encodeURIComponent(record.name)}`" />
          <a-button type="link" size="small" @click="openDrawer(record)">编辑</a-button>
          <a-button type="link" size="small" danger @click="deleting = record; deleteOpen = true">删除</a-button>
        </template>
      </a-table-column>
    </a-table>

    <a-drawer v-model:open="drawerOpen" width="760" :title="editing ? '编辑路由' : '新增路由'">
      <a-form layout="vertical">
        <a-form-item label="名称"><a-input v-model:value="formState.name" :disabled="Boolean(editing)" /></a-form-item>
        <a-form-item label="域名（一行一个）"><a-textarea v-model:value="formState.domainsText" :rows="4" /></a-form-item>
        <a-form-item label="Methods（一行一个）"><a-textarea v-model:value="formState.methodsText" :rows="3" /></a-form-item>
        <a-form-item label="路径匹配方式"><a-input v-model:value="formState.pathMatchType" /></a-form-item>
        <a-form-item label="路径匹配值"><a-input v-model:value="formState.pathMatchValue" /></a-form-item>
        <a-form-item label="目标服务(JSON)"><a-textarea v-model:value="formState.servicesJson" :rows="10" /></a-form-item>
        <a-form-item label="允许用户等级">
          <a-select v-model:value="formState.allowedConsumerLevels" mode="multiple">
            <a-select-option value="normal">normal</a-select-option>
            <a-select-option value="plus">plus</a-select-option>
            <a-select-option value="pro">pro</a-select-option>
            <a-select-option value="ultra">ultra</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
      <DrawerFooter @cancel="drawerOpen = false" @confirm="submit" />
    </a-drawer>

    <DeleteConfirmModal v-model:open="deleteOpen" :content="deleting ? `确认删除 ${deleting.name} 吗？` : ''" @confirm="confirmDelete" />
  </PageSection>
</template>
