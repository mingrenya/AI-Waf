import { motion } from 'motion/react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { useOverview } from '../hooks/useSituationData';
import type { SituationOverview } from '@/types/situation';

function riskLabel(score: number): { label: string; variant: 'default' | 'destructive' | 'secondary' | 'outline' } {
  if (score >= 75) return { label: 'Critical', variant: 'destructive' };
  if (score >= 50) return { label: 'High', variant: 'destructive' };
  if (score >= 25) return { label: 'Medium', variant: 'secondary' };
  return { label: 'Low', variant: 'outline' };
}

function trendIcon(trend: string) {
  if (trend === 'rising') return '↑';
  if (trend === 'falling') return '↓';
  return '→';
}

function StatCard({
  title,
  value,
  subtitle,
  index,
}: {
  title: string;
  value: string | number;
  subtitle?: string;
  index: number;
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.08, duration: 0.35, ease: 'easeOut' }}
    >
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-3xl font-bold tracking-tight">{value}</div>
          {subtitle && (
            <p className="mt-1 text-xs text-muted-foreground">{subtitle}</p>
          )}
        </CardContent>
      </Card>
    </motion.div>
  );
}

function StatCardSkeleton({ index }: { index: number }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.08, duration: 0.35, ease: 'easeOut' }}
    >
      <Card className="h-full">
        <CardHeader className="pb-2">
          <Skeleton className="h-4 w-24" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-8 w-16 mb-2" />
          <Skeleton className="h-3 w-20" />
        </CardContent>
      </Card>
    </motion.div>
  );
}

function OverviewStats({ data }: { data: SituationOverview }) {
  const risk = riskLabel(data.overall_risk_score);
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <StatCard title="Active Chains" value={data.active_chains} index={0} />
      <StatCard
        title="Total Attacks (24h)"
        value={data.total_chains_24h}
        index={1}
      />
      <StatCard
        title="Attackers"
        value={data.total_attackers_24h}
        index={2}
      />
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 3 * 0.08, duration: 0.35, ease: 'easeOut' }}
      >
        <Card className="h-full">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Overall Risk Score</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2">
              <span className="text-3xl font-bold tracking-tight">{data.overall_risk_score}</span>
              <Badge variant={risk.variant}>{risk.label}</Badge>
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              Trend: {trendIcon(data.risk_trend)} {data.risk_trend}
            </p>
          </CardContent>
        </Card>
      </motion.div>
    </div>
  );
}

function BreakdownSection({
  title,
  items,
}: {
  title: string;
  items: { label: string; count: number }[];
}) {
  if (!items || items.length === 0) return null;
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {items.slice(0, 6).map((item) => (
            <div key={item.label} className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground truncate max-w-[160px]">{item.label}</span>
              <Badge variant="secondary" className="ml-2 shrink-0">{item.count}</Badge>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

export default function SituationDashboard() {
  const { data, isLoading } = useOverview();

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {[0, 1, 2, 3].map((i) => (
            <StatCardSkeleton key={i} index={i} />
          ))}
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {[0, 1, 2].map((i) => (
            <Card key={i}>
              <CardHeader className="pb-2">
                <Skeleton className="h-4 w-20" />
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  {[0, 1, 2, 3].map((j) => (
                    <Skeleton key={j} className="h-5 w-full" />
                  ))}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  if (!data) return null;

  return (
    <div className="space-y-6">
      <OverviewStats data={data} />

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <BreakdownSection title="Top Attack Types" items={data.top_attack_types} />
        <BreakdownSection title="Top Attacker IPs" items={data.top_attacker_ips} />
        <BreakdownSection title="By Country" items={data.by_country} />
      </div>
    </div>
  );
}
