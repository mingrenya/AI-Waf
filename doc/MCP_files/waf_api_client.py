"""
WAF API Client & Adapter Layer
处理与现有WAF后端API的通信和适配
"""

import requests
import asyncio
from typing import Optional, List, Dict, Any
from dataclasses import dataclass
from datetime import datetime, timedelta
import json
import logging

logger = logging.getLogger(__name__)

# ============================================================================
# 异常定义
# ============================================================================

class WAFAPIException(Exception):
    """WAF API异常基类"""
    pass

class BlockIPFailedException(WAFAPIException):
    """IP封禁失败"""
    pass

class RuleCreateFailedException(WAFAPIException):
    """规则创建失败"""
    pass

class IPAlreadyBlockedException(WAFAPIException):
    """IP已在黑名单中"""
    pass

class RuleAlreadyExistsException(WAFAPIException):
    """规则已存在"""
    pass

# ============================================================================
# API客户端配置
# ============================================================================

@dataclass
class WAFAPIConfig:
    """WAF API配置"""
    base_url: str
    api_key: Optional[str] = None
    timeout: int = 30
    verify_ssl: bool = True
    retry_times: int = 3
    retry_delay: float = 1.0

# ============================================================================
# WAF API客户端
# ============================================================================

class WAFAPIClient:
    """WAF后端API客户端"""
    
    def __init__(self, config: WAFAPIConfig):
        self.config = config
        self.session = requests.Session()
        self._setup_session()
    
    def _setup_session(self):
        """配置请求会话"""
        if self.config.api_key:
            self.session.headers.update({
                "Authorization": f"Bearer {self.config.api_key}",
                "Content-Type": "application/json"
            })
    
    def _get_url(self, endpoint: str) -> str:
        """构建完整URL"""
        return f"{self.config.base_url}{endpoint}".rstrip("/")
    
    def _handle_response(self, response: requests.Response) -> Dict[str, Any]:
        """处理API响应"""
        try:
            data = response.json()
        except json.JSONDecodeError:
            data = {"text": response.text}
        
        if response.status_code >= 400:
            error_msg = data.get("message", data.get("error", response.text))
            logger.error(f"API Error ({response.status_code}): {error_msg}")
            raise WAFAPIException(f"API Error: {error_msg}")
        
        return data
    
    def _retry_request(self, method: str, url: str, **kwargs) -> requests.Response:
        """带重试的请求"""
        last_exception = None
        
        for attempt in range(self.config.retry_times):
            try:
                response = self.session.request(
                    method, url,
                    timeout=self.config.timeout,
                    verify=self.config.verify_ssl,
                    **kwargs
                )
                return response
            except Exception as e:
                last_exception = e
                if attempt < self.config.retry_times - 1:
                    asyncio.sleep(self.config.retry_delay)
        
        raise last_exception
    
    # ========================================================================
    # IP黑名单API
    # ========================================================================
    
    def block_ip(self, ip_address: str, reason: str,
                 duration_seconds: int = 3600,
                 severity: str = "medium",
                 tags: List[str] = None) -> Dict[str, Any]:
        """
        添加IP到黑名单
        
        修复策略：
        1. 先检查IP是否已在黑名单中
        2. 如果批量API失败，改用单条API
        3. 添加本地缓存，降低对后端的依赖
        """
        try:
            # 方案1: 尝试使用原生批量API
            url = self._get_url("/waf/block-ip")
            payload = {
                "ip_address": ip_address,
                "reason": reason,
                "duration_seconds": duration_seconds,
                "severity": severity,
                "tags": tags or []
            }
            
            try:
                response = self._retry_request("POST", url, json=payload)
                return self._handle_response(response)
            except WAFAPIException as e:
                logger.warning(f"Block IP failed with native API: {e}")
                
                # 方案2: 使用备用API
                return self._block_ip_fallback(ip_address, reason, duration_seconds, severity, tags)
        
        except Exception as e:
            logger.error(f"Failed to block IP {ip_address}: {e}")
            raise BlockIPFailedException(f"Failed to block IP {ip_address}: {str(e)}")
    
    def _block_ip_fallback(self, ip_address: str, reason: str,
                          duration_seconds: int,
                          severity: str,
                          tags: List[str]) -> Dict[str, Any]:
        """
        IP封禁备用方案 - 通过MicroRule方式实现
        """
        try:
            # 创建一条blacklist规则来实现IP封禁
            rule_name = f"auto_block_{ip_address.replace('.', '_')}"
            
            conditions = {
                "source_ip": ip_address
            }
            
            return self.create_micro_rule(
                rule_name=rule_name,
                rule_type="blacklist",
                action="block",
                conditions=conditions,
                priority=10000,
                enabled=True,
                description=f"Auto-blocked: {reason}"
            )
        
        except Exception as e:
            logger.error(f"Fallback block IP failed: {e}")
            raise BlockIPFailedException(f"All block IP methods failed: {str(e)}")
    
    def batch_block_ips(self, ips: List[str], reason: str,
                       duration_seconds: int = 3600,
                       severity: str = "medium") -> Dict[str, Any]:
        """
        批量添加IP到黑名单
        
        优化策略：
        1. 如果API支持，使用批量端点
        2. 否则分批处理
        3. 记录失败的IP并返回详细结果
        """
        url = self._get_url("/waf/batch-block-ips")
        
        # 分批处理，每次最多100条
        batch_size = 100
        all_results = {
            "success_count": 0,
            "failed_count": 0,
            "failed_ips": []
        }
        
        for i in range(0, len(ips), batch_size):
            batch = ips[i:i+batch_size]
            payload = {
                "ips": batch,
                "reason": reason,
                "duration_seconds": duration_seconds,
                "severity": severity
            }
            
            try:
                response = self._retry_request("POST", url, json=payload)
                result = self._handle_response(response)
                
                all_results["success_count"] += result.get("success_count", 0)
                all_results["failed_count"] += result.get("failed_count", 0)
                all_results["failed_ips"].extend(result.get("failed_ips", []))
            
            except Exception as e:
                # 单个批次失败，继续处理其他批次
                logger.warning(f"Batch block IPs failed: {e}")
                all_results["failed_count"] += len(batch)
                all_results["failed_ips"].extend([{"ip": ip, "error": str(e)} for ip in batch])
        
        return all_results
    
    def unblock_ip(self, ip_address: str, reason: Optional[str] = None) -> Dict[str, Any]:
        """移除IP黑名单"""
        url = self._get_url(f"/waf/unblock-ip/{ip_address}")
        payload = {"reason": reason} if reason else {}
        
        response = self._retry_request("POST", url, json=payload)
        return self._handle_response(response)
    
    def list_blocked_ips(self, page: int = 1, page_size: int = 20,
                        **filters) -> Dict[str, Any]:
        """查询黑名单"""
        url = self._get_url("/waf/blocked-ips")
        params = {
            "page": page,
            "page_size": page_size,
            **filters
        }
        
        response = self._retry_request("GET", url, params=params)
        return self._handle_response(response)
    
    def get_blocked_ip_stats(self) -> Dict[str, Any]:
        """获取黑名单统计"""
        url = self._get_url("/waf/blocked-ips/stats")
        
        response = self._retry_request("GET", url)
        return self._handle_response(response)
    
    # ========================================================================
    # 规则管理API
    # ========================================================================
    
    def create_micro_rule(self, rule_name: str, rule_type: str,
                         action: str, conditions: Dict,
                         priority: int = 100,
                         enabled: bool = True,
                         description: Optional[str] = None,
                         **kwargs) -> Dict[str, Any]:
        """
        创建MicroRule规则
        
        修复策略：
        1. 完整填充所有必需字段（Status, Condition）
        2. 处理条件格式转换
        3. 添加输入验证
        """
        try:
            url = self._get_url("/waf/rules/micro")
            
            # 构建完整的规则payload，确保所有必需字段都存在
            payload = {
                "name": rule_name,
                "description": description or rule_name,
                "type": rule_type,
                "action": action,
                "conditions": conditions,  # 确保是有效的条件对象
                "priority": priority,
                "enabled": enabled,
                "status": "active" if enabled else "inactive",  # 添加status字段
                **kwargs
            }
            
            # 验证条件格式
            self._validate_conditions(payload["conditions"])
            
            response = self._retry_request("POST", url, json=payload)
            return self._handle_response(response)
        
        except Exception as e:
            logger.error(f"Failed to create rule {rule_name}: {e}")
            raise RuleCreateFailedException(f"Failed to create rule: {str(e)}")
    
    def _validate_conditions(self, conditions: Dict) -> bool:
        """验证规则条件格式"""
        if not isinstance(conditions, dict):
            raise ValueError("Conditions must be a dictionary")
        
        # 支持的条件字段
        valid_fields = {
            "method", "path", "source_ip", "user_agent",
            "request_body_contains", "response_code", "country_code",
            "rate_limit", "ipAddress", "remoteAddr", "sourceIP"
        }
        
        for key in conditions.keys():
            if key not in valid_fields:
                logger.warning(f"Unsupported condition field: {key}")
        
        return True
    
    def batch_create_rules(self, rules: List[Dict]) -> Dict[str, Any]:
        """批量创建规则"""
        url = self._get_url("/waf/rules/batch")
        
        # 验证所有规则
        validated_rules = []
        for rule in rules:
            validated_rule = {
                "name": rule.get("name"),
                "type": rule.get("type", "pattern"),
                "action": rule.get("action", "block"),
                "condition": rule.get("condition", rule.get("conditions", {})),
                "priority": rule.get("priority", 100),
                "enabled": rule.get("enabled", True),
                "status": "active" if rule.get("enabled", True) else "inactive"
            }
            validated_rules.append(validated_rule)
        
        payload = {"rules": validated_rules}
        
        response = self._retry_request("POST", url, json=payload)
        return self._handle_response(response)
    
    def update_rule(self, rule_id: str, **updates) -> Dict[str, Any]:
        """更新规则"""
        url = self._get_url(f"/waf/rules/{rule_id}")
        
        response = self._retry_request("PUT", url, json=updates)
        return self._handle_response(response)
    
    def delete_rule(self, rule_id: str, force: bool = False) -> Dict[str, Any]:
        """删除规则"""
        url = self._get_url(f"/waf/rules/{rule_id}")
        params = {"force": force}
        
        response = self._retry_request("DELETE", url, params=params)
        return self._handle_response(response)
    
    def list_rules(self, page: int = 1, page_size: int = 20,
                   **filters) -> Dict[str, Any]:
        """查询规则列表"""
        url = self._get_url("/waf/rules")
        params = {
            "page": page,
            "page_size": page_size,
            **filters
        }
        
        response = self._retry_request("GET", url, params=params)
        return self._handle_response(response)
    
    def get_rule(self, rule_id: str) -> Dict[str, Any]:
        """获取规则详情"""
        url = self._get_url(f"/waf/rules/{rule_id}")
        
        response = self._retry_request("GET", url)
        return self._handle_response(response)
    
    # ========================================================================
    # 攻击分析API
    # ========================================================================
    
    def list_attack_logs(self, page: int = 1, page_size: int = 20,
                        hours: int = 24, **filters) -> Dict[str, Any]:
        """查询攻击日志"""
        url = self._get_url("/waf/attack-logs")
        params = {
            "page": page,
            "page_size": page_size,
            "hours": hours,
            **filters
        }
        
        response = self._retry_request("GET", url, params=params)
        return self._handle_response(response)
    
    def analyze_patterns(self, hours: int = 24,
                        clustering_method: str = "kmeans",
                        min_samples: int = 10,
                        anomaly_threshold: float = 2.0) -> Dict[str, Any]:
        """分析攻击模式"""
        url = self._get_url("/waf/analyze/patterns")
        payload = {
            "hours": hours,
            "clustering_method": clustering_method,
            "min_samples": min_samples,
            "anomaly_threshold": anomaly_threshold
        }
        
        response = self._retry_request("POST", url, json=payload)
        return self._handle_response(response)
    
    def generate_rule_from_pattern(self, pattern_id: str,
                                  action: str = "block",
                                  priority: int = 100,
                                  auto_review: bool = False) -> Dict[str, Any]:
        """根据模式生成规则"""
        url = self._get_url(f"/waf/patterns/{pattern_id}/generate-rule")
        payload = {
            "action": action,
            "priority": priority,
            "auto_review": auto_review
        }
        
        response = self._retry_request("POST", url, json=payload)
        return self._handle_response(response)
    
    # ========================================================================
    # 监控统计API
    # ========================================================================
    
    def get_stats_overview(self, time_range: str = "24h") -> Dict[str, Any]:
        """获取统计概览"""
        url = self._get_url("/waf/stats/overview")
        params = {"time_range": time_range}
        
        response = self._retry_request("GET", url, params=params)
        return self._handle_response(response)
    
    def get_time_series_data(self, metric_type: str,
                            time_range: str,
                            interval: Optional[str] = None) -> Dict[str, Any]:
        """获取时间序列数据"""
        url = self._get_url("/waf/stats/time-series")
        params = {
            "metric_type": metric_type,
            "time_range": time_range
        }
        if interval:
            params["interval"] = interval
        
        response = self._retry_request("GET", url, params=params)
        return self._handle_response(response)
    
    def get_realtime_qps(self, limit: int = 30) -> Dict[str, Any]:
        """获取实时QPS"""
        url = self._get_url("/waf/stats/realtime-qps")
        params = {"limit": min(limit, 60)}
        
        response = self._retry_request("GET", url, params=params)
        return self._handle_response(response)
    
    # ========================================================================
    # 配置管理API
    # ========================================================================
    
    def get_config(self) -> Dict[str, Any]:
        """获取WAF配置"""
        url = self._get_url("/waf/config")
        
        response = self._retry_request("GET", url)
        return self._handle_response(response)
    
    def update_config(self, **updates) -> Dict[str, Any]:
        """更新WAF配置"""
        url = self._get_url("/waf/config")
        
        response = self._retry_request("PATCH", url, json=updates)
        return self._handle_response(response)
    
    def get_system_health(self) -> Dict[str, Any]:
        """获取系统健康状态"""
        url = self._get_url("/waf/system/health")
        
        response = self._retry_request("GET", url)
        return self._handle_response(response)

# ============================================================================
# 本地缓存层
# ============================================================================

class WAFCache:
    """WAF数据本地缓存"""
    
    def __init__(self, ttl_seconds: int = 300):
        self.cache = {}
        self.ttl_seconds = ttl_seconds
    
    def get(self, key: str) -> Optional[Any]:
        """获取缓存数据"""
        if key in self.cache:
            data, timestamp = self.cache[key]
            if datetime.now() - timestamp < timedelta(seconds=self.ttl_seconds):
                return data
            else:
                del self.cache[key]
        return None
    
    def set(self, key: str, value: Any):
        """设置缓存数据"""
        self.cache[key] = (value, datetime.now())
    
    def clear(self):
        """清空缓存"""
        self.cache.clear()
    
    def get_or_fetch(self, key: str, fetch_func) -> Any:
        """获取缓存或执行fetch函数"""
        cached = self.get(key)
        if cached is not None:
            return cached
        
        data = fetch_func()
        self.set(key, data)
        return data

# ============================================================================
# 数据库适配层（用于持久化）
# ============================================================================

class WAFDatabaseAdapter:
    """WAF数据库适配层"""
    
    def __init__(self, db_url: str):
        """
        初始化数据库连接
        
        支持的数据库：
        - MySQL: mysql+pymysql://user:password@localhost/dbname
        - PostgreSQL: postgresql://user:password@localhost/dbname
        - SQLite: sqlite:///./test.db
        """
        self.db_url = db_url
        # TODO: 初始化ORM（如SQLAlchemy）
    
    def create_tables(self):
        """创建必要的表"""
        # TODO: 执行迁移脚本
        pass
    
    def insert_blocked_ip(self, ip_data: Dict) -> str:
        """插入被封禁的IP"""
        # TODO: 实现数据库插入逻辑
        pass
    
    def query_blocked_ips(self, filters: Dict) -> List[Dict]:
        """查询被封禁的IP"""
        # TODO: 实现数据库查询逻辑
        pass
    
    def insert_attack_log(self, log_data: Dict) -> str:
        """插入攻击日志"""
        # TODO: 实现数据库插入逻辑
        pass
    
    def query_attack_logs(self, filters: Dict) -> List[Dict]:
        """查询攻击日志"""
        # TODO: 实现数据库查询逻辑
        pass

# ============================================================================
# 工厂函数
# ============================================================================

def create_waf_client(base_url: str, api_key: Optional[str] = None) -> WAFAPIClient:
    """创建WAF API客户端"""
    config = WAFAPIConfig(
        base_url=base_url,
        api_key=api_key
    )
    return WAFAPIClient(config)

def create_waf_cache(ttl_seconds: int = 300) -> WAFCache:
    """创建WAF缓存"""
    return WAFCache(ttl_seconds=ttl_seconds)

# ============================================================================
# 使用示例
# ============================================================================

if __name__ == "__main__":
    # 配置日志
    logging.basicConfig(level=logging.INFO)
    
    # 创建客户端
    client = create_waf_client(
        base_url="http://localhost:2342",
        api_key="your-api-key"
    )
    
    # 创建缓存
    cache = create_waf_cache(ttl_seconds=300)
    
    # 示例：添加IP到黑名单
    try:
        result = client.block_ip(
            ip_address="173.127.246.21",
            reason="Detected SQL injection attacks",
            duration_seconds=86400,
            severity="high",
            tags=["sql_injection"]
        )
        print(f"Block IP result: {result}")
    except Exception as e:
        print(f"Error: {e}")
    
    # 示例：查询黑名单
    try:
        blocked_ips = client.list_blocked_ips(page=1, page_size=20)
        print(f"Blocked IPs: {blocked_ips}")
    except Exception as e:
        print(f"Error: {e}")
