import api from './index';
import type { GeoIPConfig, GeoIPUpdateRequest } from '@/types/geoip';

const BASE = '/geoip';

export const getGeoIPConfig = () =>
  api.get<{ data: GeoIPConfig }>(`${BASE}/config`);

export const updateGeoIPConfig = (req: GeoIPUpdateRequest) =>
  api.put(`${BASE}/config`, req);
