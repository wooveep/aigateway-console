export interface DashboardInfo {
  builtIn: boolean;
  uid?: string;
  url: string;
}

export enum DashboardType {
  MAIN = "MAIN",
  AI = "AI",
  LOG = "LOG",
}

export interface NativeDashboardVariableState {
  value: string;
  options: string[];
}

export interface NativeDashboardGridPos {
  h: number;
  w: number;
  x: number;
  y: number;
}

export interface NativeDashboardPoint {
  time: number;
  value: number;
}

export interface NativeDashboardSeries {
  name: string;
  labels: Record<string, string>;
  points: NativeDashboardPoint[];
}

export interface NativeDashboardStat {
  value?: number | null;
}

export interface NativeDashboardTableColumn {
  key: string;
  title: string;
}

export interface NativeDashboardTable {
  columns: NativeDashboardTableColumn[];
  rows: Record<string, string | number | null>[];
}

export interface NativeDashboardPanel {
  id: number;
  title: string;
  type: 'stat' | 'timeseries' | 'table';
  unit: string;
  gridPos: NativeDashboardGridPos;
  error?: string;
  stat?: NativeDashboardStat;
  series?: NativeDashboardSeries[];
  table?: NativeDashboardTable;
}

export interface NativeDashboardRow {
  title: string;
  collapsed: boolean;
  panels: NativeDashboardPanel[];
}

export interface NativeDashboardData {
  title: string;
  type: DashboardType;
  from: number;
  to: number;
  defaultRangeMs: number;
  variables: {
    gateway: NativeDashboardVariableState;
    namespace: NativeDashboardVariableState;
  };
  rows: NativeDashboardRow[];
}
