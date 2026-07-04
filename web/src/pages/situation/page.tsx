import { useState, useEffect } from 'react';
import SituationDashboard from '@/feature/situation/components/SituationDashboard';
import AttackChainTimeline from '@/feature/situation/components/AttackChainTimeline';
import AttackerRankingChart from '@/feature/situation/components/AttackerRankingChart';
import AttackerDrawer from '@/feature/situation/components/AttackerDrawer';
import { useSituationWebSocket } from '@/feature/situation/hooks/useSituationWebSocket';
import { useTrends } from '@/feature/situation/hooks/useSituationData';
import { useQueryClient } from '@tanstack/react-query';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { PageTransition } from '@/components/ui/animation/page-transition';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Wifi, WifiOff, TrendingUp, Activity, Zap } from 'lucide-react';

export default function SituationPage() {
  const [selectedIp, setSelectedIp] = useState<string | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [wsEvents, setWsEvents] = useState<{ type: string; time: string }[]>([]);
  const { connected, subscribe } = useSituationWebSocket();
  const queryClient = useQueryClient();
  const { data: trendsData, isLoading: trendsLoading } = useTrends('24h');

  // Subscribe to WebSocket events and invalidate queries on updates
  useEffect(() => {
    const unsub1 = subscribe('situation:alert', () => {
      setWsEvents((prev) => [...prev.slice(-19), { type: 'alert', time: new Date().toLocaleTimeString() }]);
      queryClient.invalidateQueries({ queryKey: ['situation', 'overview'] });
      queryClient.invalidateQueries({ queryKey: ['situation', 'chains'] });
    });
    const unsub2 = subscribe('situation:update', () => {
      setWsEvents((prev) => [...prev.slice(-19), { type: 'update', time: new Date().toLocaleTimeString() }]);
      queryClient.invalidateQueries({ queryKey: ['situation', 'overview'] });
    });
    const unsub3 = subscribe('situation:attack', () => {
      setWsEvents((prev) => [...prev.slice(-19), { type: 'attack', time: new Date().toLocaleTimeString() }]);
      queryClient.invalidateQueries({ queryKey: ['situation', 'chains'] });
      queryClient.invalidateQueries({ queryKey: ['situation', 'attackers'] });
    });

    return () => { unsub1(); unsub2(); unsub3(); };
  }, [subscribe, queryClient]);

  function handleSelectAttacker(ip: string) {
    setSelectedIp(ip);
    setDrawerOpen(true);
  }

  const eventTypeColor = (type: string) => {
    switch (type) {
      case 'alert': return 'var(--destructive)';
      case 'attack': return 'var(--warning)';
      case 'update': return 'var(--color-primary-5)';
      default: return 'var(--text-muted)';
    }
  };

  return (
    <PageTransition className="flex flex-col gap-6 p-4 md:p-6 lg:p-8 max-w-[1600px] mx-auto w-full">
      {/* Page header + status bar */}
      <div className="flex items-center justify-between flex-wrap gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight" style={{ color: 'var(--text-primary)' }}>Situation Awareness</h1>
          <p className="text-sm mt-1" style={{ color: 'var(--text-muted)' }}>
            Real-time attack surface monitoring and threat intelligence
          </p>
        </div>
        <div className="flex items-center gap-3">
          {/* WebSocket connection indicator */}
          <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium"
            style={{
              background: connected ? 'rgba(34,197,94,0.1)' : 'rgba(239,68,68,0.1)',
              color: connected ? '#22c55e' : '#ef4444',
              border: `1px solid ${connected ? 'rgba(34,197,94,0.2)' : 'rgba(239,68,68,0.2)'}`,
            }}
          >
            {connected ? <Wifi className="h-3.5 w-3.5" /> : <WifiOff className="h-3.5 w-3.5" />}
            {connected ? 'Live' : 'Disconnected'}
          </div>

          {/* Recent events feed */}
          {wsEvents.length > 0 && (
            <div className="flex items-center gap-1">
              {wsEvents.slice(-5).map((e, i) => (
                <div
                  key={i}
                  className="w-2 h-2 rounded-full"
                  style={{ background: eventTypeColor(e.type) }}
                  title={`${e.type} at ${e.time}`}
                />
              ))}
              <span className="text-xs ml-1" style={{ color: 'var(--text-muted)' }}>
                {wsEvents.length} events
              </span>
            </div>
          )}
        </div>
      </div>

      {/* Dashboard overview stats */}
      <SituationDashboard />

      {/* Trends Chart Panel */}
      {trendsLoading ? (
        <div className="surface-card p-5">
          <Skeleton className="h-4 w-32 mb-4 bg-white/10" />
          <div className="grid grid-cols-3 gap-4 mb-4">
            {[0, 1, 2].map((i) => (
              <Skeleton key={i} className="h-8 w-20 bg-white/10" />
            ))}
          </div>
          <Skeleton className="h-[120px] w-full bg-white/10" />
        </div>
      ) : trendsData ? (
        <Card className="surface-card">
          <CardHeader className="pb-2">
            <CardTitle className="flex items-center gap-2 text-base">
              <TrendingUp className="h-4 w-4" style={{ color: 'var(--color-primary-5)' }} />
              Attack Trends (24h)
            </CardTitle>
          </CardHeader>
          <CardContent>
            {/* Summary row */}
            <div className="grid grid-cols-3 gap-4 mb-4">
              <div className="text-center">
                <p className="text-xs mb-1" style={{ color: 'var(--text-dim)' }}>Active Attackers</p>
                <p className="text-2xl font-bold font-mono" style={{ color: 'var(--text-primary)' }}>
                  {trendsData.active_attackers}
                </p>
              </div>
              <div className="text-center">
                <p className="text-xs mb-1" style={{ color: 'var(--text-dim)' }}>New Chains (24h)</p>
                <p className="text-2xl font-bold font-mono" style={{ color: 'var(--text-primary)' }}>
                  {trendsData.new_chains_24h}
                </p>
              </div>
              <div className="text-center">
                <p className="text-xs mb-1" style={{ color: 'var(--text-dim)' }}>Frequent Types</p>
                <div className="flex flex-wrap justify-center gap-1">
                  {trendsData.frequent_types?.slice(0, 3).map((ft) => (
                    <Badge key={ft.label} variant="secondary" className="text-xs">{ft.label}</Badge>
                  ))}
                </div>
              </div>
            </div>

            {/* Timeline mini bar chart */}
            {trendsData.timeline && trendsData.timeline.length > 0 && (
              <div className="flex items-end gap-[2px] h-[80px] w-full">
                {trendsData.timeline.map((point, idx) => {
                  const maxEvents = Math.max(...trendsData.timeline.map((p) => p.total_events), 1);
                  const blockedPct = (point.blocked_count / maxEvents) * 100;
                  const detectPct = ((point.total_events - point.blocked_count) / maxEvents) * 100;
                  return (
                    <div
                      key={idx}
                      className="flex-1 flex flex-col justify-end"
                      title={`${new Date(point.timestamp * 1000).toLocaleTimeString()}: ${point.total_events} events, ${point.blocked_count} blocked`}
                    >
                      <div
                        className="w-full rounded-t-sm"
                        style={{ height: `${Math.max(detectPct, 2)}%`, background: 'var(--color-primary-5)', opacity: 0.6 }}
                      />
                      <div
                        className="w-full"
                        style={{ height: `${Math.max(blockedPct, 0)}%`, background: 'var(--destructive)', opacity: 0.5, marginTop: '-1px' }}
                      />
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>
      ) : null}

      {/* Tabs for timeline and ranking */}
      <Tabs defaultValue="timeline" className="w-full">
        <TabsList>
          <TabsTrigger value="timeline">Attack Chain Timeline</TabsTrigger>
          <TabsTrigger value="ranking">Attacker Ranking</TabsTrigger>
        </TabsList>
        <TabsContent value="timeline" className="mt-4">
          <AttackChainTimeline onSelectAttacker={handleSelectAttacker} />
        </TabsContent>
        <TabsContent value="ranking" className="mt-4">
          <div className="surface-card p-6">
            <h2 className="text-lg font-semibold mb-4" style={{ color: 'var(--text-primary)' }}>Attacker Ranking</h2>
            <AttackerRankingChart />
          </div>
        </TabsContent>
      </Tabs>

      {/* Attacker drawer */}
      <AttackerDrawer
        ip={selectedIp}
        open={drawerOpen}
        onOpenChange={setDrawerOpen}
      />
    </PageTransition>
  );
}
