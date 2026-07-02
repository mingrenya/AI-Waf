import { useState } from 'react';
import SituationDashboard from '@/feature/situation/components/SituationDashboard';
import AttackChainTimeline from '@/feature/situation/components/AttackChainTimeline';
import AttackerRankingChart from '@/feature/situation/components/AttackerRankingChart';
import AttackerDrawer from '@/feature/situation/components/AttackerDrawer';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export default function SituationPage() {
  const [selectedIp, setSelectedIp] = useState<string | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);

  function handleSelectAttacker(ip: string) {
    setSelectedIp(ip);
    setDrawerOpen(true);
  }

  return (
    <div className="flex flex-col gap-6 p-4 md:p-6 lg:p-8 max-w-[1600px] mx-auto w-full">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Situation Awareness</h1>
        <p className="text-sm text-muted-foreground mt-1">
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
          <Card>
            <CardHeader>
              <CardTitle>Attacker Ranking</CardTitle>
            </CardHeader>
            <CardContent>
              <AttackerRankingChart />
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Attacker drawer */}
      <AttackerDrawer
        ip={selectedIp}
        open={drawerOpen}
        onOpenChange={setDrawerOpen}
      />
    </div>
  );
}
