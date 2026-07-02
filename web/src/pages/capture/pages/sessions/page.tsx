import { useState, useEffect, useCallback, useRef } from 'react';
import { listSessions, getDownloadUrl } from '@/api/capture';
import type { CaptureSessionResponse } from '@/types/capture';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
}

function statusBadge(status: string) {
  const map: Record<string, 'default' | 'destructive' | 'secondary' | 'outline'> = {
    running: 'default',
    completed: 'secondary',
    stopped: 'outline',
    error: 'destructive',
  };
  return <Badge variant={map[status] || 'outline'}>{status}</Badge>;
}

export default function SessionsPage() {
  const [data, setData] = useState<{ sessions: CaptureSessionResponse[]; total: number }>({ sessions: [], total: 0 });
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const fetchSessions = useCallback(async () => {
    try {
      const res = await listSessions();
      setData(res.data.data);
    } catch { /* ignore */ }
  }, []);

  useEffect(() => {
    fetchSessions();
    intervalRef.current = setInterval(fetchSessions, 5000);
    return () => { if (intervalRef.current) clearInterval(intervalRef.current); };
  }, [fetchSessions]);

  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between">
        <CardTitle>Capture Sessions</CardTitle>
        <span className="text-sm text-muted-foreground">{data.total} total</span>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Created</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Interface</TableHead>
              <TableHead>Filter</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Packets</TableHead>
              <TableHead>Size</TableHead>
              <TableHead>Download</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.sessions.length === 0 && (
              <TableRow>
                <TableCell colSpan={8} className="text-center text-muted-foreground">No capture sessions yet</TableCell>
              </TableRow>
            )}
            {data.sessions.map(s => (
              <TableRow key={s.id}>
                <TableCell className="text-xs">{new Date(s.created_at).toLocaleString()}</TableCell>
                <TableCell>{s.description || '-'}</TableCell>
                <TableCell className="font-mono text-xs">{s.interface}</TableCell>
                <TableCell className="font-mono text-xs max-w-[150px] truncate">{s.bpf_filter || '-'}</TableCell>
                <TableCell>{statusBadge(s.status)}</TableCell>
                <TableCell>{s.packet_count}</TableCell>
                <TableCell>{formatBytes(s.file_size)}</TableCell>
                <TableCell>
                  {(s.status === 'completed' || s.status === 'stopped') && (
                    <Button asChild variant="outline" size="sm">
                      <a href={getDownloadUrl(s.id)} download>PCAP</a>
                    </Button>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
