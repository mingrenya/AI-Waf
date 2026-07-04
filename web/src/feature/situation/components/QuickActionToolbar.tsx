import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { quickAction } from '@/api/situation';
import { toast } from '@/store';
import { Shield, Lock, Ban } from 'lucide-react';

interface QuickActionToolbarProps {
  sourceIp: string;
}

const ACTIONS: { label: string; durationHours: number; variant: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link'; action: 'block' | 'blacklist' | 'both' }[] = [
  { label: 'Block 1h', durationHours: 1, variant: 'outline', action: 'block' },
  { label: 'Block 24h', durationHours: 24, variant: 'outline', action: 'block' },
  { label: 'Block 7d', durationHours: 168, variant: 'outline', action: 'block' },
  { label: 'Block Permanent', durationHours: 87600, variant: 'destructive', action: 'block' },
  { label: 'Add to Blacklist', durationHours: 0, variant: 'destructive', action: 'blacklist' },
];

export default function QuickActionToolbar({ sourceIp }: QuickActionToolbarProps) {
  const [loading, setLoading] = useState<string | null>(null);

  async function handleAction(label: string, action: 'block' | 'blacklist' | 'both', durationHours: number) {
    setLoading(label);
    try {
      await quickAction({
        source_ip: sourceIp,
        action,
        duration_hours: durationHours,
        reason: `Quick action: ${label}`,
      });
      toast({
        title: label,
        description: `Action applied to ${sourceIp}`,
        variant: 'success',
      });
    } catch {
      toast({
        title: 'Action Failed',
        description: `Failed to apply "${label}" to ${sourceIp}`,
        variant: 'destructive',
      });
    } finally {
      setLoading(null);
    }
  }

  return (
    <div className="flex items-center gap-2 flex-wrap">
      <span className="text-xs flex items-center gap-1 mr-1" style={{color:'var(--text-muted)'}}>
        <Shield className="h-3.5 w-3.5" />
        Quick Actions:
      </span>
      {ACTIONS.map(({ label, durationHours, variant, action }) => (
        <Button
          key={label}
          variant={variant}
          size="sm"
          disabled={loading !== null}
          onClick={() => handleAction(label, action, durationHours)}
          className="text-xs h-7"
        >
          {loading === label ? (
            <span className="flex items-center gap-1">
              <span className="animate-spin h-3 w-3 border-2 border-current border-t-transparent rounded-full" />
              Processing...
            </span>
          ) : action === 'blacklist' ? (
            <>
              <Ban className="h-3 w-3" />
              {label}
            </>
          ) : (
            <>
              <Lock className="h-3 w-3" />
              {label}
            </>
          )}
        </Button>
      ))}
    </div>
  );
}
