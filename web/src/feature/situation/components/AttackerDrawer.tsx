import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useAttackerProfile } from '../hooks/useSituationData';
import QuickActionToolbar from './QuickActionToolbar';
import type { LogEventItem } from '@/types/situation';

interface AttackerDrawerProps {
  ip: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

function riskBadge(score: number, label: string): { variant: 'default' | 'destructive' | 'secondary' | 'outline' } {
  if (label === 'critical' || score >= 75) return { variant: 'destructive' };
  if (label === 'high' || score >= 50) return { variant: 'destructive' };
  if (label === 'medium' || score >= 25) return { variant: 'secondary' };
  return { variant: 'outline' };
}

function severityVariant(severity: string): 'default' | 'destructive' | 'secondary' | 'outline' {
  switch (severity?.toLowerCase()) {
    case 'critical':
      return 'destructive';
    case 'high':
      return 'destructive';
    case 'medium':
      return 'secondary';
    default:
      return 'outline';
  }
}

export default function AttackerDrawer({ ip, open, onOpenChange }: AttackerDrawerProps) {
  const { data, isLoading } = useAttackerProfile(ip ?? '');

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="surface-modal max-w-lg max-h-[90vh] p-0 gap-0 sm:rounded-lg" style={{color:'var(--text-primary)'}}>
        <ScrollArea className="max-h-[calc(90vh-64px)]" scrollbarVariant="neon">
          {isLoading ? (
            <div className="p-6 space-y-4" style={{color:'var(--text-primary)'}}>
              <Skeleton className="h-6 w-48 bg-white/10" />
              <Skeleton className="h-4 w-32 bg-white/10" />
              <div className="grid grid-cols-2 gap-4">
                {[0, 1, 2, 3, 4, 5].map((i) => (
                  <div key={i} className="space-y-2">
                    <Skeleton className="h-4 w-20 bg-white/10" />
                    <Skeleton className="h-5 w-28 bg-white/10" />
                  </div>
                ))}
              </div>
              <div className="space-y-2">
                {[0, 1, 2].map((i) => (
                  <Skeleton key={i} className="h-12 w-full bg-white/10" />
                ))}
              </div>
            </div>
          ) : data ? (
            <div className="p-6 space-y-6" style={{color:'var(--text-primary)'}}>
              {/* IP and risk */}
              <DialogHeader>
                <DialogTitle className="flex items-center gap-2 flex-wrap" style={{color:'var(--text-primary)'}}>
                  <span className="font-mono text-lg">{data.source_ip}</span>
                  <Badge variant={riskBadge(data.risk_score, data.risk_label).variant}>
                    {data.risk_label} ({data.risk_score})
                  </Badge>
                </DialogTitle>
              </DialogHeader>

              {/* Stats grid */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-xs" style={{color:'var(--text-dim)'}}>Country</p>
                  <p className="text-sm font-medium" style={{color:'var(--text-secondary)'}}>
                    {data.geo_country}{data.geo_city ? `, ${data.geo_city}` : ''}
                  </p>
                </div>
                <div>
                  <p className="text-xs" style={{color:'var(--text-dim)'}}>Total Attacks</p>
                  <p className="text-sm font-medium" style={{color:'var(--text-secondary)'}}>{data.total_attacks}</p>
                </div>
                <div>
                  <p className="text-xs" style={{color:'var(--text-dim)'}}>Top Attack Type</p>
                  <p className="text-sm font-medium" style={{color:'var(--text-secondary)'}}>{data.top_attack_type}</p>
                </div>
                <div>
                  <p className="text-xs" style={{color:'var(--text-dim)'}}>Unique Targets</p>
                  <p className="text-sm font-medium" style={{color:'var(--text-secondary)'}}>{data.unique_target_sites}</p>
                </div>
                <div>
                  <p className="text-xs" style={{color:'var(--text-dim)'}}>Attack Phase</p>
                  <Badge variant="secondary" className="capitalize">
                    {data.attack_phase?.replace(/_/g, ' ') ?? '-'}
                  </Badge>
                </div>
                <div>
                  <p className="text-xs" style={{color:'var(--text-dim)'}}>Tools</p>
                  <p className="text-sm font-medium truncate max-w-[160px]" style={{color:'var(--text-secondary)'}}>
                    {data.tools_identified ?? '-'}
                  </p>
                </div>
                <div>
                  <p className="text-xs" style={{color:'var(--text-dim)'}}>First Seen</p>
                  <p className="text-sm font-medium" style={{color:'var(--text-secondary)'}}>
                    {data.first_seen ? new Date(data.first_seen).toLocaleString() : '-'}
                  </p>
                </div>
                <div>
                  <p className="text-xs" style={{color:'var(--text-dim)'}}>Last Seen</p>
                  <p className="text-sm font-medium" style={{color:'var(--text-secondary)'}}>
                    {data.last_seen ? new Date(data.last_seen).toLocaleString() : '-'}
                  </p>
                </div>
              </div>

              {/* Flags */}
              <div className="flex items-center gap-2 flex-wrap">
                {data.is_automated && (
                  <Badge variant="secondary">Automated</Badge>
                )}
                {data.is_persistent && (
                  <Badge variant="destructive">Persistent</Badge>
                )}
                {data.unique_attack_types > 1 && (
                  <Badge variant="outline">{data.unique_attack_types} attack types</Badge>
                )}
              </div>

              {/* Active hours */}
              {data.active_hours && data.active_hours.length > 0 && (
                <div>
                  <p className="text-xs mb-2" style={{color:'var(--text-dim)'}}>Active Hours (UTC)</p>
                  <div className="flex flex-wrap gap-1">
                    {data.active_hours.map((h) => (
                      <Badge key={h} variant="outline" className="text-xs">
                        {String(h).padStart(2, '0')}:00
                      </Badge>
                    ))}
                  </div>
                </div>
              )}

              {/* Recent events */}
              {data.recent_events && data.recent_events.length > 0 && (
                <div>
                  <p className="text-sm font-medium mb-3" style={{color:'var(--text-secondary)'}}>Recent Events</p>
                  <div className="space-y-2">
                    {data.recent_events.map((event: LogEventItem) => (
                      <div
                        key={event.id}
                        className="flex items-center justify-between rounded-md border border-black/10 dark:border-white/10 px-3 py-2 text-sm bg-black/5 dark:bg-white/5"
                      >
                        <div className="flex items-center gap-2 min-w-0">
                          <Badge variant={severityVariant(event.severity)} className="text-xs shrink-0">
                            {event.severity}
                          </Badge>
                          <span className="truncate" style={{color:'var(--text-secondary)'}}>{event.attack_type}</span>
                        </div>
                        <div className="flex items-center gap-2 shrink-0 ml-2">
                          <span className="text-xs" style={{color:'var(--text-dim)'}}>{event.action}</span>
                          <span className="text-xs" style={{color:'var(--text-dim)'}}>
                            {event.timestamp ? new Date(event.timestamp).toLocaleTimeString() : ''}
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          ) : null}
        </ScrollArea>

        {/* QuickActionToolbar pinned at bottom */}
        {ip && (
          <div className="border-t border-black/10 dark:border-white/10 px-6 py-3 rounded-b-lg" style={{ background: 'rgba(255,255,255,0.06)' }}>
            <QuickActionToolbar sourceIp={ip} />
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
