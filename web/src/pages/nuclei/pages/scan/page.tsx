import { useState, useEffect, useCallback } from 'react';
import { startScan, listTasks, cancelTask } from '@/api/nuclei';
import type { ScanTaskResponse, TemplateInfo } from '@/types/nuclei';
import { listTemplates } from '@/api/nuclei';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';

export default function ScanPage() {
  const [siteId, setSiteId] = useState('');
  const [targetUrl, setTargetUrl] = useState('');
  const [selectedTemplates, setSelectedTemplates] = useState<string[]>([]);
  const [severity, setSeverity] = useState('critical,high,medium');
  const [tasks, setTasks] = useState<ScanTaskResponse[]>([]);
  const [templates, setTemplates] = useState<TemplateInfo[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchTasks = useCallback(async () => {
    try {
      const data = await listTasks() as unknown as ScanTaskResponse[];
      setTasks(data || []);
    } catch {
      // ignore fetch errors
    }
  }, []);

  const fetchTemplates = useCallback(async () => {
    try {
      const data = await listTemplates() as unknown as TemplateInfo[];
      setTemplates(data || []);
    } catch {
      // ignore
    }
  }, []);

  useEffect(() => {
    fetchTasks();
    fetchTemplates();
  }, [fetchTasks, fetchTemplates]);

  const handleStartScan = async () => {
    if (!siteId || !targetUrl) return;
    setLoading(true);
    try {
      await startScan({
        site_id: siteId,
        target_url: targetUrl,
        templates: selectedTemplates.length > 0 ? selectedTemplates : undefined,
        severity,
      });
      fetchTasks();
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  };

  const handleCancelTask = async (id: string) => {
    try {
      await cancelTask(id);
      fetchTasks();
    } catch {
      // ignore
    }
  };

  const toggleTemplate = (path: string) => {
    setSelectedTemplates((prev) =>
      prev.includes(path) ? prev.filter((p) => p !== path) : [...prev, path]
    );
  };

  const statusVariant = (status: string): 'default' | 'secondary' | 'destructive' | 'outline' => {
    switch (status) {
      case 'running': return 'default';
      case 'completed': return 'secondary';
      case 'failed': return 'destructive';
      default: return 'outline';
    }
  };

  return (
    <div className="flex flex-col gap-4 p-4">
      <Card>
        <CardHeader>
          <CardTitle>Start Scan</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="siteId">Site ID</Label>
              <Input
                id="siteId"
                value={siteId}
                onChange={(e) => setSiteId(e.target.value)}
                placeholder="Enter site ID"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="targetUrl">Target URL</Label>
              <Input
                id="targetUrl"
                value={targetUrl}
                onChange={(e) => setTargetUrl(e.target.value)}
                placeholder="https://example.com"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="severity">Severity</Label>
              <Select value={severity} onValueChange={setSeverity}>
                <SelectTrigger id="severity">
                  <SelectValue placeholder="Select severity" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="critical,high,medium,low,info">All</SelectItem>
                  <SelectItem value="critical,high,medium">Critical, High, Medium</SelectItem>
                  <SelectItem value="critical">Critical Only</SelectItem>
                  <SelectItem value="critical,high">Critical, High</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2 flex items-end">
              <Button onClick={handleStartScan} disabled={loading}>
                {loading ? 'Starting...' : 'Start Scan'}
              </Button>
            </div>
          </div>
          {templates.length > 0 && (
            <div className="space-y-2">
              <Label>Templates ({selectedTemplates.length} selected)</Label>
              <div className="flex flex-wrap gap-2 max-h-32 overflow-y-auto border rounded-md p-2">
                {templates.map((t) => (
                  <Badge
                    key={t.path}
                    variant={selectedTemplates.includes(t.path) ? 'default' : 'outline'}
                    className="cursor-pointer"
                    onClick={() => toggleTemplate(t.path)}
                  >
                    {t.name}
                  </Badge>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Scan Tasks</CardTitle>
        </CardHeader>
        <CardContent>
          {tasks.length === 0 ? (
            <p className="text-muted-foreground text-sm">No scan tasks yet.</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>Site ID</TableHead>
                  <TableHead>Target URL</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Findings / Total</TableHead>
                  <TableHead>Created At</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {tasks.map((task) => (
                  <TableRow key={task.id}>
                    <TableCell className="font-mono text-xs">{task.id}</TableCell>
                    <TableCell>{task.site_id}</TableCell>
                    <TableCell className="max-w-[200px] truncate">{task.target_url}</TableCell>
                    <TableCell>
                      <Badge variant={statusVariant(task.status)}>{task.status}</Badge>
                    </TableCell>
                    <TableCell>
                      {task.findings} / {task.total}
                    </TableCell>
                    <TableCell className="text-xs">{new Date(task.created_at).toLocaleString()}</TableCell>
                    <TableCell>
                      {task.status === 'running' && (
                        <Button size="sm" variant="outline" onClick={() => handleCancelTask(task.id)}>
                          Cancel
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
