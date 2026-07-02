import { useState, useEffect, useCallback, useRef } from 'react';
import { startCapture, stopCapture, getSession } from '@/api/capture';
import type { CaptureSessionResponse } from '@/types/capture';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { toast } from '@/store';

export default function ControlPage() {
  const [form, setForm] = useState({
    interface: 'eth0',
    bpf_filter: 'tcp port 80 or tcp port 443',
    max_packets: '0',
    duration_secs: '0',
    description: '',
  });
  const [running, setRunning] = useState<CaptureSessionResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const updateField = (field: string) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm(prev => ({ ...prev, [field]: e.target.value }));

  const handleStart = useCallback(async () => {
    setLoading(true);
    try {
      const res = await startCapture({
        interface: form.interface || undefined,
        bpf_filter: form.bpf_filter || undefined,
        max_packets: Number(form.max_packets) || 0,
        duration_secs: Number(form.duration_secs) || 0,
        description: form.description || undefined,
      });
      const session = res.data.data;
      setRunning(session);
      toast({ title: 'Capture Started', description: `Session ${session.id} is running`, variant: 'success' });
    } catch {
      toast({ title: 'Error', description: 'Failed to start capture', variant: 'destructive' });
    } finally {
      setLoading(false);
    }
  }, [form]);

  const handleStop = useCallback(async () => {
    if (!running) return;
    setLoading(true);
    try {
      await stopCapture(running.id);
      toast({ title: 'Capture Stopped', description: `Session ${running.id} stopped`, variant: 'success' });
      if (pollRef.current) clearInterval(pollRef.current);
      // 最终刷新
      const res = await getSession(running.id);
      setRunning(res.data.data);
    } catch {
      toast({ title: 'Error', description: 'Failed to stop capture', variant: 'destructive' });
    } finally {
      setLoading(false);
    }
  }, [running]);

  useEffect(() => {
    if (running && running.status === 'running') {
      pollRef.current = setInterval(async () => {
        try {
          const res = await getSession(running.id);
          setRunning(res.data.data);
        } catch { /* ignore poll errors */ }
      }, 3000);
    }
    return () => { if (pollRef.current) clearInterval(pollRef.current); };
  }, [running?.id]);

  function statusBadge(status: string) {
    const map: Record<string, 'default' | 'destructive' | 'secondary' | 'outline'> = {
      running: 'default',
      completed: 'secondary',
      stopped: 'outline',
      error: 'destructive',
    };
    return <Badge variant={map[status] || 'outline'}>{status}</Badge>;
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* Start Capture Form */}
      <Card>
        <CardHeader>
          <CardTitle>Start Capture</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <Label>Interface</Label>
            <Input value={form.interface} onChange={updateField('interface')} placeholder="eth0" />
          </div>
          <div>
            <Label>BPF Filter</Label>
            <Input value={form.bpf_filter} onChange={updateField('bpf_filter')} placeholder="tcp port 80 or tcp port 443" />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label>Max Packets (0=unlimited)</Label>
              <Input type="number" value={form.max_packets} onChange={updateField('max_packets')} />
            </div>
            <div>
              <Label>Duration sec (0=unlimited)</Label>
              <Input type="number" value={form.duration_secs} onChange={updateField('duration_secs')} />
            </div>
          </div>
          <div>
            <Label>Description</Label>
            <Input value={form.description} onChange={updateField('description')} placeholder="Traffic analysis..." />
          </div>
          <Button onClick={handleStart} disabled={loading || running?.status === 'running'} className="w-full">
            {loading ? 'Starting...' : 'Start Capture'}
          </Button>
        </CardContent>
      </Card>

      {/* Running Session Monitor */}
      <Card>
        <CardHeader>
          <CardTitle>Running Session</CardTitle>
        </CardHeader>
        <CardContent>
          {running ? (
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">ID</span>
                <span className="font-mono text-xs">{running.id}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Status</span>
                {statusBadge(running.status)}
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Interface</span>
                <span className="font-medium">{running.interface}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Filter</span>
                <span className="font-mono text-xs">{running.bpf_filter || '(none)'}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Packets</span>
                <span className="font-bold text-lg">{running.packet_count}</span>
              </div>
              {running.description && (
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Description</span>
                  <span>{running.description}</span>
                </div>
              )}
              {running.status === 'running' && (
                <Button onClick={handleStop} variant="destructive" disabled={loading} className="w-full mt-4">
                  {loading ? 'Stopping...' : 'Stop Capture'}
                </Button>
              )}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">No capture running. Start one from the form.</p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
