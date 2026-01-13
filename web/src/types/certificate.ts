export interface Certificate {
  id: string;
  name: string;
  description?: string;
  publicKey: string;
  privateKey: string;
  expireDate: string;
  fingerPrint: string;
  issuerName: string;
  domains: string[];
  createdAt: string;
  updatedAt: string;
}

export interface CertificateListResponse {
  items: Certificate[];
  total: number;
}

export interface CertificateCreateRequest {
  name: string;
  description?: string;
  publicKey: string;
  privateKey: string;
  expireDate?: string;
  fingerPrint?: string;
  issuerName?: string;
  domains?: string[];
}

export interface CertificateUpdateRequest {
  name?: string;
  description?: string;
  publicKey?: string;
  privateKey?: string;
  expireDate?: string;
  fingerPrint?: string;
  issuerName?: string;
  domains?: string[];
}

export interface ParsedCertificate {
  expireDate: string;
  fingerPrint: string;
  issuerName: string;
  domains: string[];
}

export interface CertificateFile {
  content: string;
  filename: string;
} 