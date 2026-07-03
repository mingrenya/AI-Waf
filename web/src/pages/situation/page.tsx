import { useState } from 'react';
import SituationDashboard from '@/feature/situation/components/SituationDashboard';
import AttackChainTimeline from '@/feature/situation/components/AttackChainTimeline';
import AttackerRankingChart from '@/feature/situation/components/AttackerRankingChart';
import AttackerDrawer from '@/feature/situation/components/AttackerDrawer';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { PageTransition } from '@/components/ui/animation/page-transition';

export default function SituationPage() {
  const [selectedIp, setSelectedIp] = useState<string | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);

  function handleSelectAttacker(ip: string) {
    setSelectedIp(ip);
    setDrawerOpen(true);
  }

  return (
    <PageTransition className="flex flex-col gap-6 p-4 md:p-6 lg:p-8 max-w-[1600px] mx-auto w-full">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-white">Situation Awareness</h1>
        <p className="text-sm text-white/50 mt-1">
          Real-time attack surface monitoring and threat intelligence
        </p>
      </div>

      {/* Dashboard overview stats */}
      <SituationDashboard />

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
          <div className="glass-card p-6">
            <h2 className="text-lg font-semibold text-white mb-4">Attacker Ranking</h2>
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
