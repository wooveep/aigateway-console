export interface OrgDepartmentNode {
  departmentId: string;
  name: string;
  parentDepartmentId?: string;
  adminConsumerName?: string;
  adminDisplayName?: string;
  createdAdminTempPassword?: string;
  level?: number;
  memberCount?: number;
  children?: OrgDepartmentNode[];
}

export interface OrgAccountRecord {
  consumerName: string;
  displayName?: string;
  email?: string;
  status?: 'active' | 'disabled' | 'pending' | string;
  userLevel?: 'normal' | 'plus' | 'pro' | 'ultra' | string;
  source?: string;
  departmentId?: string;
  departmentName?: string;
  departmentPath?: string;
  isDepartmentAdmin?: boolean;
  lastLoginAt?: string;
  tempPassword?: string;
}

export interface OrgAccountMutation {
  consumerName: string;
  displayName?: string;
  email?: string;
  userLevel?: string;
  password?: string;
  status?: string;
  departmentId?: string;
}

export interface OrgDepartmentMutation {
  name?: string;
  parentDepartmentId?: string;
  adminMode?: 'existing' | 'create' | string;
  adminConsumerName?: string;
  adminDisplayName?: string;
  adminEmail?: string;
  adminUserLevel?: string;
  adminPassword?: string;
}

export interface OrgAccountSSORebindRequest {
  targetConsumerName: string;
}

export interface OrgDepartmentMoveRequest {
  parentDepartmentId?: string;
}

export interface AssetGrantRecord {
  assetType?: string;
  assetId?: string;
  subjectType?: 'consumer' | 'department' | 'user_level' | string;
  subjectId?: string;
}

export interface OrgImportResult {
  createdDepartments: number;
  updatedDepartments: number;
  createdAccounts: number;
  updatedAccounts: number;
}
