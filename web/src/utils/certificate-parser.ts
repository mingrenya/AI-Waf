import * as forge from 'node-forge'

/**
 * 解析后的证书信息接口
 */
export interface ParsedCertificate {
    /** 证书过期时间 */
    expireDate: string
    /** 证书指纹 (SHA-256) */
    fingerPrint: string
    /** 证书颁发者名称 */
    issuerName: string
    /** 证书包含的域名列表 */
    domains: string[]
}

/**
 * 从PEM格式的证书中解析关键信息
 * @param certPem PEM格式的证书
 * @returns 解析后的证书信息
 */
export function parseCertificate(certPem: string): ParsedCertificate {
    try {
        // 从PEM格式解析证书
        const cert = forge.pki.certificateFromPem(certPem)

        // 解析颁发者信息
        const issuerCN = cert.issuer.getField('CN')
        const issuerName = issuerCN ? issuerCN.value : '未知'

        // 解析过期时间
        const expireDate = cert.validity.notAfter.toISOString()

        // 计算证书指纹 (SHA-256)
        const md = forge.md.sha256.create()
        const derBytes = forge.asn1.toDer(forge.pki.certificateToAsn1(cert)).getBytes()
        md.update(derBytes)
        const fingerPrint = md.digest().toHex().match(/.{2}/g)?.join(':') || ''

        // 解析域名
        const domains: string[] = []

        // 尝试从主题别名中获取域名 (SAN扩展)
        try {
            const altNamesExt = cert.getExtension('subjectAltName')
            if (altNamesExt && 'altNames' in altNamesExt) {
                const altNames = altNamesExt.altNames as Array<{ type: number; value: string }>
                if (Array.isArray(altNames)) {
                    altNames.forEach((altName) => {
                        if (altName.type === 2) { // DNS类型
                            domains.push(altName.value)
                        }
                    })
                }
            }
        } catch (e) {
            console.error('解析域名扩展失败', e instanceof Error ? e.message : String(e))
        }

        // 如果没有从扩展中获取到域名，尝试从CN中获取
        if (domains.length === 0) {
            const commonName = cert.subject.getField('CN')
            if (commonName) {
                domains.push(commonName.value)
            }
        }

        return {
            expireDate,
            fingerPrint,
            issuerName,
            domains
        }
    } catch (error) {
        console.error('证书解析失败:', error instanceof Error ? error.message : String(error))
        return {
            expireDate: '',
            fingerPrint: '',
            issuerName: '解析失败',
            domains: []
        }
    }
}

/**
 * 从文件中读取文本内容
 * @param file 文件对象
 * @returns Promise，解析为文件内容
 */
export function readFileAsText(file: File): Promise<string> {
    return new Promise<string>((resolve, reject) => {
        const reader = new FileReader()

        reader.onload = (event: ProgressEvent<FileReader>) => {
            if (event.target && typeof event.target.result === 'string') {
                resolve(event.target.result)
            } else {
                reject(new Error('读取文件失败：文件内容无效'))
            }
        }

        reader.onerror = (event: ProgressEvent<FileReader>) => {
            reject(new Error(`读取文件错误: ${event.target?.error?.message || '未知错误'}`))
        }

        reader.readAsText(file)
    })
}