import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Shield, Plus, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { getGeoIPConfig, updateGeoIPConfig } from '@/api/geoip';
import type { GeoIPConfig } from '@/types/geoip';
import { toast } from '@/hooks/use-toast';

export function GeoIPFilter() {
  const { t } = useTranslation();
  const [config, setConfig] = useState<GeoIPConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [newCountry, setNewCountry] = useState('');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      const res = await getGeoIPConfig();
      setConfig(res.data.data);
    } catch {
      // 静默失败
    } finally {
      setLoading(false);
    }
  };

  const addCountry = () => {
    const code = newCountry.trim().toUpperCase();
    if (!code || code.length !== 2) return;
    if (!config) return;
    if (config.block_countries.includes(code)) return;
    setConfig({ ...config, block_countries: [...config.block_countries, code] });
    setNewCountry('');
  };

  const removeCountry = (code: string) => {
    if (!config) return;
    setConfig({
      ...config,
      block_countries: config.block_countries.filter((c) => c !== code),
    });
  };

  const toggleMode = () => {
    if (!config) return;
    setConfig({ ...config, allow_mode: !config.allow_mode });
  };

  const saveConfig = async () => {
    if (!config) return;
    setSaving(true);
    try {
      await updateGeoIPConfig({
        block_countries: config.block_countries,
        allow_mode: config.allow_mode,
      });
      toast({ title: t('common.saved') });
    } catch {
      toast({ title: t('common.saveFailed'), variant: 'destructive' });
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return <div className="animate-pulse h-24 bg-muted rounded-lg" />;
  }

  return (
    <div className="bg-background rounded-xl border p-6 animate-fade-in-up">
      <div className="space-y-3 mb-8">
        <div className="flex items-center gap-2 pb-2 border-b border-border">
          <Shield className="w-5 h-5 text-iconStroke dark:text-primary" />
          <h3 className="text-lg font-medium text-foreground">
            {t('globalSetting.geoip.title', 'GeoIP 国家/地区过滤')}
          </h3>
        </div>
        <div className="pl-7">
          <p className="text-sm text-muted-foreground">
            {t('globalSetting.geoip.description', '根据 IP 地理位置阻止或允许特定国家/地区的访问')}
          </p>
        </div>
      </div>

      <div className="pl-7 space-y-6">
        {/* 模式切换 */}
        <div className="flex items-center justify-between">
          <div>
            <Label className="text-sm font-medium">
              {t('globalSetting.geoip.allowMode', '白名单模式')}
            </Label>
            <p className="text-xs text-muted-foreground mt-1">
              {config?.allow_mode
                ? '仅允许列表中国家/地区访问'
                : '阻止列表中国家/地区访问'}
            </p>
          </div>
          <Switch checked={config?.allow_mode ?? false} onCheckedChange={toggleMode} />
        </div>

        {/* 国家列表 */}
        <div>
          <Label className="text-sm font-medium mb-2 block">
            {config?.allow_mode
              ? '允许的国家/地区 (ISO 两位码)'
              : '封锁的国家/地区 (ISO 两位码)'}
          </Label>

          <div className="flex gap-2 mb-3">
            <Input
              placeholder="CN, US, RU..."
              value={newCountry}
              onChange={(e) => setNewCountry(e.target.value.toUpperCase())}
              onKeyDown={(e) => e.key === 'Enter' && addCountry()}
              maxLength={2}
              className="w-32"
            />
            <Button variant="outline" size="sm" onClick={addCountry}>
              <Plus className="w-4 h-4 mr-1" />
              {t('common.add')}
            </Button>
          </div>

          <div className="flex flex-wrap gap-2 min-h-[36px]">
            {config?.block_countries.length === 0 && (
              <span className="text-sm text-muted-foreground italic">
                {t('globalSetting.geoip.empty', '未配置任何国家/地区')}
              </span>
            )}
            {config?.block_countries.map((code) => (
              <Badge key={code} variant="secondary" className="gap-1 pr-1">
                {code}
                <button
                  onClick={() => removeCountry(code)}
                  className="ml-1 hover:text-destructive"
                >
                  <X className="w-3 h-3" />
                </button>
              </Badge>
            ))}
          </div>
        </div>

        {/* 保存按钮 */}
        <Button onClick={saveConfig} disabled={saving} className="w-full sm:w-auto">
          {saving ? t('common.saving') : t('common.save')}
        </Button>
      </div>
    </div>
  );
}
