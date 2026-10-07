/**
 * ASESAMO — API Convenience Façade
 * Location: asesamo-web/lib/api.ts
 * Re-exports singleton methods and types for standard import paths (@/lib/api).
 */

import { apiClient } from './api/client';
import { AccessLogsQueryParams } from './api/types';

export * from './api/types';
export * from './api/client';

export const login = (email: string, pass: string) => apiClient.login(email, pass);
export const getAccessPoints = () => apiClient.getAccessPoints();
export const checkIn = (workerId: string, accessPointId: string, notes?: string) =>
  apiClient.checkIn(workerId, accessPointId, notes);
export const checkOut = (workerId: string, accessPointId: string, plantId?: string) =>
  apiClient.checkOut(workerId, accessPointId, plantId);
export const getAccessLogs = (params?: AccessLogsQueryParams) => apiClient.getAccessLogs(params);
export const logout = () => apiClient.logout();

export default apiClient;
