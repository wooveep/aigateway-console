<script setup lang="ts">
import { computed } from 'vue';
import { buildLineData, formatValue } from '@/features/dashboard/dashboard-native';
import type { NativeDashboardSeries } from '@/interfaces/dashboard';

const props = defineProps<{
  series: NativeDashboardSeries[];
  rangeMs: number;
  unit: string;
}>();

const VIEWBOX_WIDTH = 720;
const VIEWBOX_HEIGHT = 220;
const PADDING = {
  top: 18,
  right: 18,
  bottom: 30,
  left: 18,
};

const lineData = computed(() => buildLineData(props.series, props.rangeMs));

const xDomain = computed(() => {
  const timestamps = lineData.value.map((item) => item.timeValue);
  if (!timestamps.length) {
    return { min: 0, max: 1 };
  }
  return {
    min: timestamps[0],
    max: timestamps[timestamps.length - 1] || timestamps[0] + 1,
  };
});

const yDomain = computed(() => {
  const values = lineData.value.map((item) => item.value);
  if (!values.length) {
    return { min: 0, max: 1 };
  }
  const min = Math.min(...values);
  const max = Math.max(...values);
  if (min === max) {
    return { min: min - 1, max: max + 1 };
  }
  return { min, max };
});

const seriesGroups = computed(() => {
  const groups = new Map<string, { color: string; items: typeof lineData.value }>();
  lineData.value.forEach((item) => {
    const existed = groups.get(item.series);
    if (existed) {
      existed.items.push(item);
      return;
    }
    groups.set(item.series, { color: item.color, items: [item] });
  });
  return Array.from(groups.entries()).map(([series, payload]) => ({
    series,
    color: payload.color,
    items: payload.items,
    path: payload.items.map((item) => `${toX(item.timeValue)},${toY(item.value)}`).join(' '),
  }));
});

const xTicks = computed(() => {
  if (!lineData.value.length) {
    return [];
  }
  const first = lineData.value[0];
  const middle = lineData.value[Math.floor(lineData.value.length / 2)];
  const last = lineData.value[lineData.value.length - 1];
  const candidates = [first, middle, last].filter(Boolean);
  const seen = new Set<number>();
  return candidates.filter((item) => {
    if (seen.has(item.timeValue)) {
      return false;
    }
    seen.add(item.timeValue);
    return true;
  }).map((item) => ({
    x: toX(item.timeValue),
    label: item.timeLabel,
  }));
});

const yTicks = computed(() => {
  const min = yDomain.value.min;
  const max = yDomain.value.max;
  const middle = (min + max) / 2;
  return [
    { y: toY(max), label: formatValue(max, props.unit) },
    { y: toY(middle), label: formatValue(middle, props.unit) },
    { y: toY(min), label: formatValue(min, props.unit) },
  ];
});

function toX(value: number) {
  const width = VIEWBOX_WIDTH - PADDING.left - PADDING.right;
  const span = Math.max(1, xDomain.value.max - xDomain.value.min);
  return PADDING.left + ((value - xDomain.value.min) / span) * width;
}

function toY(value: number) {
  const height = VIEWBOX_HEIGHT - PADDING.top - PADDING.bottom;
  const span = Math.max(1, yDomain.value.max - yDomain.value.min);
  return PADDING.top + (1 - (value - yDomain.value.min) / span) * height;
}
</script>

<template>
  <div class="native-line-chart">
    <div class="native-line-chart__legend">
      <span v-for="item in seriesGroups" :key="item.series" class="native-line-chart__legend-item">
        <span class="native-line-chart__legend-dot" :style="{ backgroundColor: item.color }" />
        <span class="native-line-chart__legend-text">{{ item.series }}</span>
      </span>
    </div>

    <svg
      class="native-line-chart__svg"
      :viewBox="`0 0 ${VIEWBOX_WIDTH} ${VIEWBOX_HEIGHT}`"
      preserveAspectRatio="none"
      role="img"
      aria-label="Native dashboard line chart"
    >
      <line
        v-for="tick in yTicks"
        :key="tick.y"
        class="native-line-chart__grid"
        :x1="PADDING.left"
        :x2="VIEWBOX_WIDTH - PADDING.right"
        :y1="tick.y"
        :y2="tick.y"
      />

      <polyline
        v-for="item in seriesGroups"
        :key="item.series"
        :points="item.path"
        :stroke="item.color"
        class="native-line-chart__polyline"
      />

      <g v-for="tick in xTicks" :key="tick.x">
        <text class="native-line-chart__axis native-line-chart__axis--x" :x="tick.x" :y="VIEWBOX_HEIGHT - 8">
          {{ tick.label }}
        </text>
      </g>

      <g v-for="tick in yTicks" :key="`${tick.y}-${tick.label}`">
        <text class="native-line-chart__axis native-line-chart__axis--y" :x="VIEWBOX_WIDTH - 8" :y="tick.y - 6">
          {{ tick.label }}
        </text>
      </g>
    </svg>
  </div>
</template>

<style scoped>
.native-line-chart {
  display: grid;
  gap: 10px;
  height: 100%;
}

.native-line-chart__legend {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
}

.native-line-chart__legend-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--portal-text-soft);
  font-size: 12px;
}

.native-line-chart__legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
}

.native-line-chart__legend-text {
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.native-line-chart__svg {
  width: 100%;
  height: 100%;
  min-height: 150px;
}

.native-line-chart__grid {
  stroke: rgba(16, 35, 63, 0.08);
  stroke-width: 1;
}

.native-line-chart__polyline {
  fill: none;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.native-line-chart__axis {
  fill: var(--portal-text-muted);
  font-size: 11px;
}

.native-line-chart__axis--x {
  text-anchor: middle;
}

.native-line-chart__axis--y {
  text-anchor: end;
}
</style>
