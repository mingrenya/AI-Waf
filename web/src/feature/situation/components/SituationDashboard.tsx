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
}: {
  title: string;
  value: string | number;
  subtitle?: string;
  index: number;
}) {
  return (
    <div className="surface-card p-5 h-full">
      <p className="text-xs font-medium mb-2" style={{color:'var(--text-muted)'}}>{title}</p>
      <div className="text-3xl font-bold font-mono" style={{color:'var(--text-primary)'}}>{value}</div>
      {subtitle && (
        <p className="mt-1 text-xs" style={{color:'var(--text-dim)'}}>{subtitle}</p>
      )}
    </div>
  );
}

function StatCardSkeleton() {
  return (
    <div className="surface-card p-5 h-full">
      <Skeleton className="h-4 w-24 mb-3 bg-white/10" />
      <Skeleton className="h-8 w-16 mb-2 bg-white/10" />
      <Skeleton className="h-3 w-20 bg-white/10" />
    </div>
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
      <div className="surface-card p-5 h-full">
        <p className="text-xs font-medium mb-2" style={{color:'var(--text-muted)'}}>Overall Risk Score</p>
        <div className="flex items-center gap-2">
          <span className="text-3xl font-bold font-mono" style={{color:'var(--text-primary)'}}>{data.overall_risk_score}</span>
          <Badge variant={risk.variant}>{risk.label}</Badge>
        </div>
        <p className="mt-1 text-xs" style={{color:'var(--text-dim)'}}>
          Trend: {trendIcon(data.risk_trend)} {data.risk_trend}
        </p>
      </div>
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
    <div className="surface-card p-5">
      <h3 className="text-sm font-medium mb-3" style={{color:'var(--text-primary)'}}>{title}</h3>
      <div className="space-y-2">
        {items.slice(0, 6).map((item) => (
          <div key={item.label} className="flex items-center justify-between text-sm">
            <span className="truncate max-w-[160px]" style={{color:'var(--text-secondary)'}}>{item.label}</span>
            <Badge variant="secondary" className="ml-2 shrink-0">{item.count}</Badge>
          </div>
        ))}
      </div>
    </div>
  );
}

export default function SituationDashboard() {
  const { data, isLoading } = useOverview();

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {[0, 1, 2, 3].map((i) => (
            <StatCardSkeleton key={i} />
          ))}
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {[0, 1, 2].map((i) => (
            <div key={i} className="surface-card p-5">
              <Skeleton className="h-4 w-20 mb-4 bg-white/10" />
              <div className="space-y-2">
                {[0, 1, 2, 3].map((j) => (
                  <Skeleton key={j} className="h-5 w-full bg-white/10" />
                ))}
              </div>
            </div>
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
