import { InitParams } from '@/interfaces/system';
import request from './request';

export async function getSystemInfo(): Promise<any> {
  return await request.get('/system/info');
}

export async function getConfigs(): Promise<any> {
  return await request.get('/system/config');
}

export async function initialize(payload: InitParams): Promise<any> {
  return request.post<any, any>('/system/init', payload);
}

export async function getAIGatewayConfig(): Promise<any> {
  return await request.get('/system/aigateway-config');
}

export async function updateAIGatewayConfig(config: string): Promise<any> {
  return request.put<any, any>('/system/aigateway-config', { config });
}
