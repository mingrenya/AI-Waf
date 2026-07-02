import { motion } from 'motion/react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useChains } from '../hooks/useSituationData';
import type { ChainSummary } from '@/types/situation';

const STAGE_COLORS: Record<string, string> = {
  reconnaissance: 'bg-blue-400 dark:bg-blue-500',
  scanning: 'bg-cyan-400 dark:bg-cyan-500',
  exploitation: 'bg-orange-400 dark:bg-orange-500',
  lateral_movement: 'bg-purple-400 dark:bg-purple-500',
  command_and_control: 'bg-red-400 dark:bg-red-500',
  exfiltration: 'bg-rose-400 dark:bg-rose-500',
  unknown: 'bg-gray-400 dark:bg-gray-500',
};

const STAGE_LABELS: Record<string, string> = {
  reconnaissance: 'Recon',
  scanning: 'Scan',
  exploitation: 'Exploit',
  lateral_movement: 'Lat. Move',
  command_and_control: 'C2',
  exfiltration: 'Exfil',
  unknown: 'Unknown',
};

function riskBadge(score: number): { label: string; variant: 'default' | 'destructive' | 'secondary' | 'outline' } {
  if (score >= 75) return { label: 'Critical', variant: 'destructive' };
  if (score >= 50) return { label: 'High', variant: 'destructive' };
  if (score >= 25) return { label: 'Medium', variant: 'secondary' };
  return { label: 'Low', variant: 'outline' };
}

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

interface AttackChainTimelineProps {
  onSelectAttacker?: (ip: string) => void;
}

export default function AttackChainTimeline({ onSelectAttacker }: AttackChainTimelineProps) {
  const { data, isLoading } = useChains({ active: true, page_size: 20 });

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Attack Chain Timeline</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[0, 1, 2, 3, 4].map((i) => (
              <div key={i} className="flex items-center gap-4">
                <Skeleton className="h-10 w-10 rounded-full" />
                <div className="flex-1 space-y-2">
                  <Skeleton className="h-4 w-32" />
                  <Skeleton className="h-3 w-48" />
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!data || !data.chains || data.chains.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Attack Chain Timeline</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground text-center py-8">No active attack chains</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Attack Chain Timeline</CardTitle>
      </CardHeader>
      <CardContent>
        <ScrollArea className="h-[400px] pr-4">
          <div className="relative">
            {/* Timeline vertical line */}
            <div className="absolute left-5 top-0 bottom-0 w-px bg-border" />

            <div className="space-y-0">
              {data.chains.map((chain: ChainSummary, idx: number) => (
                <motion.div
                  key={chain.id}
                  initial={{ opacity: 0, x: -16 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: idx * 0.04, duration: 0.3 }}
                  className="relative flex items-start gap-4 py-3 pl-12 pr-2 cursor-pointer hover:bg-muted/40 rounded-md transition-colors"
                  onClick={() => onSelectAttacker?.(chain.source_ip)}
                >
                  {/* Timeline dot */}
                  <div
                    className={`absolute left-[14px] top-[18px] h-3 w-3 rounded-full border-2 border-background ${
                      chain.active ? 'bg-green-500' : 'bg-gray-400'
                    }`}
                  />

                  {/* IP & Country */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="font-mono text-sm font-medium">{chain.source_ip}</span>
                      {chain.geo_country && (
                        <Badge variant="outline" className="text-xs">
                          {chain.geo_country}
                        </Badge>
                      )}
                      <Badge
                        variant={riskBadge(chain.risk_score).variant}
                        className="text-xs ml-auto shrink-0"
                      >
                        {chain.risk_score}
                      </Badge>
                    </div>

                    {/* Stage progression dots */}
                    <div className="flex items-center gap-1.5 mt-1.5">
                      {chain.stages.map((stage, si) => (
                        <div key={`${stage}-${si}`} className="flex items-center gap-1.5">
                          {si > 0 && (
                            <div className="w-3 h-px bg-border" />
                          )}
                          <div
                            className={`h-2.5 w-2.5 rounded-full ${STAGE_COLORS[stage] ?? 'bg-gray-400'}`}
                            title={STAGE_LABELS[stage] ?? stage}
                          />
                        </div>
                      ))}
                      <span className="text-xs text-muted-foreground ml-2">
                        {chain.stages.length} stage{chain.stages.length > 1 ? 's' : ''}
                      </span>
                    </div>

                    {/* Time */}
                    <p className="text-xs text-muted-foreground mt-1">
                      Last seen: {timeAgo(chain.last_seen)}
                    </p>
                  </div>
                </motion.div>
              ))}
            </div>
          </div>
        </ScrollArea>
      </CardContent>
    </Card>
  );
}
