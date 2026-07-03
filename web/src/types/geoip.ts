export interface GeoIPConfig {
  block_countries: string[];
  allow_mode: boolean;
}

export interface GeoIPUpdateRequest {
  block_countries: string[];
  allow_mode: boolean;
}
