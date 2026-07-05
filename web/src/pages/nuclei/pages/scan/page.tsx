import { useState, useEffect, useCallback } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { startScan, getTaskDetail, listTasks, cancelTask } from '@/api/nuclei';
import type { ScanTaskResponse, ScanTaskDetail, NucleiFinding, TemplateInfo } from '@/types/nuclei';
import { listTemplates } from '@/api/nuclei';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { X, AlertTriangle, Bug, Info, Terminal, Loader2, Eye } from 'lucide-react';
import { formatDateTime } from '@/lib/utils';

const SEVERITY_MAP: Record<string, { color: string; bg: string; icon: typeof Bug }> = {
  critical: { color: 'text-red-400', bg: 'bg-red-500/10 border-red-500/20', icon: AlertTriangle },
  high: { color: 'text-orange-400', bg: 'bg-orange-500/10 border-orange-500/20', icon: AlertTriangle },
  medium: { color: 'text-yellow-400', bg: 'bg-yellow-500/10 border-yellow-500/20', icon: Bug },
  low: { color: 'text-blue-400', bg: 'bg-blue-500/10 border-blue-500/20', icon: Info },
  info: { color: 'text-slate-400', bg: 'bg-slate-500/10 border-slate-500/20', icon: Info },
};

function unwrapData<T>(value: unknown): T | unknown {
  if (value && typeof value === 'object' && 'data' in value) {
    return (value as { data?: unknown }).data;
  }

  return value;
}

function normalizeArray<T>(value: unknown): T[] {
  const unwrapped = unwrapData<T[]>(value);
  return Array.isArray(unwrapped) ? unwrapped : [];
}

function normalizeTaskDetail(value: unknown): ScanTaskDetail | null {
  const unwrapped = unwrapData<ScanTaskDetail>(value);
  if (!unwrapped || typeof unwrapped !== 'object') return null;

  const detail = unwrapped as ScanTaskDetail;
  return {
    ...detail,
    findings: Array.isArray(detail.findings) ? detail.findings : [],
  };
}

export default function ScanPage() {
  const [siteId, setSiteId] = useState('');
  const [targetUrl, setTargetUrl] = useState('');
  const [selectedTemplates, setSelectedTemplates] = useState<string[]>([]);
  const [severity, setSeverity] = useState('critical,high,medium');
  const [loading, setLoading] = useState(false);
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const { data: tasks = [], refetch: refetchTasks, isLoading: tasksLoading } = useQuery({
    queryKey: ['nuclei-tasks'],
    queryFn: () => listTasks().then((res) => normalizeArray<ScanTaskResponse>(res)),
    refetchInterval: 5000,
  });

  const { data: templates = [] } = useQuery({
    queryKey: ['nuclei-templates'],
    queryFn: () => listTemplates().then((res) => normalizeArray<TemplateInfo>(res)),
  });

  const { data: taskDetail, isLoading: detailLoading } = useQuery({
    queryKey: ['nuclei-task-detail', selectedTaskId],
    queryFn: () => getTaskDetail(selectedTaskId!).then((res) => normalizeTaskDetail(res)),
    enabled: !!selectedTaskId,
  });

  const fetchTasks = useCallback(async () => {
    try {
      const data = await listTasks();
      return normalizeArray<ScanTaskResponse>(data);
    } catch {
      return [];
    }
  }, []);

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
      refetchTasks();
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  };

  const handleCancelTask = async (id: string) => {
    try {
      await cancelTask(id);
      refetchTasks();
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
      case 'cron': return 'outline';
      default: return 'outline';
    }
  };

  return (
    <div className="flex flex-col gap-4 p-4">
      {/* ===== 发起扫描表单 ===== */}
      <Card className="surface-card">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Terminal className="h-5 w-5" style={{ color: 'var(--color-primary-5)' }} />
            Start Scan
          </CardTitle>
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

      {/* ===== 任务列表 ===== */}
      <Card className="surface-card">
        <CardHeader>
          <CardTitle>Scan Tasks</CardTitle>
        </CardHeader>
        <CardContent>
          {tasksLoading ? (
            <div className="flex items-center gap-2 text-sm" style={{ color: 'var(--text-muted)' }}>
              <Loader2 className="h-4 w-4 animate-spin" />
              Loading tasks...
            </div>
          ) : tasks.length === 0 ? (
            <p className="text-sm" style={{ color: 'var(--text-muted)' }}>No scan tasks yet.</p>
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
                  <TableRow
                    key={task.id}
                    className={selectedTaskId === task.id ? 'bg-primary/5' : ''}
                  >
                    <TableCell className="font-mono text-xs">{task.id.slice(0, 8)}</TableCell>
                    <TableCell>{task.site_id}</TableCell>
                    <TableCell className="max-w-[200px] truncate">{task.target_url}</TableCell>
                    <TableCell>
                      <Badge variant={statusVariant(task.status)}>{task.status}</Badge>
                    </TableCell>
                    <TableCell>
                      {task.findings} / {task.total}
                    </TableCell>
                    <TableCell className="text-xs">{formatDateTime(task.created_at)}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => setSelectedTaskId(selectedTaskId === task.id ? null : task.id)}
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                        {task.status === 'running' && (
                          <Button size="sm" variant="outline" onClick={() => handleCancelTask(task.id)}>
                            Cancel
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* ===== Findings 详情面板 ===== */}
      {selectedTaskId && (
        <Card className="surface-card">
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <span>Scan Findings</span>
              <Button variant="ghost" size="icon" onClick={() => setSelectedTaskId(null)}>
                <X className="h-4 w-4" />
              </Button>
            </CardTitle>
          </CardHeader>
          <CardContent>
            {detailLoading ? (
              <div className="flex items-center gap-2 text-sm" style={{ color: 'var(--text-muted)' }}>
                <Loader2 className="h-4 w-4 animate-spin" />
                Loading details...
              </div>
            ) : !taskDetail ? (
              <p className="text-sm" style={{ color: 'var(--text-muted)' }}>No detail available.</p>
            ) : taskDetail.findings.length === 0 ? (
              <p className="text-sm" style={{ color: 'var(--text-muted)' }}>No findings from this scan.</p>
            ) : (
              <div className="space-y-3">
                {taskDetail.findings.map((f, idx) => (
                  <FindingCard key={`${f.template_id}-${idx}`} finding={f} index={idx} />
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function FindingCard({ finding, index }: { finding: NucleiFinding; index: number }) {
  const severity = SEVERITY_MAP[finding.severity] ?? SEVERITY_MAP.info;
  const SeverityIcon = severity.icon;

  return (
    <div className="p-4 rounded-lg border" style={{ borderColor: 'var(--surface-root-border)', background: 'var(--surface-card-bg)' }}>
      <div className="flex items-start justify-between mb-2">
        <div className="flex items-center gap-2">
          <span className="text-xs font-mono" style={{ color: 'var(--text-dim)' }}>#{index + 1}</span>
          <SeverityIcon className={`h-4 w-4 ${severity.color}`} />
          <span className="font-medium text-sm" style={{ color: 'var(--text-primary)' }}>
            {finding.name}
          </span>
          <Badge variant="outline" className={`text-xs ${severity.color} ${severity.bg}`}>
            {finding.severity}
          </Badge>
        </div>
        <span className="text-xs font-mono" style={{ color: 'var(--text-muted)' }}>
          {finding.template_id}
        </span>
      </div>

      {finding.matched_at && (
        <div className="mb-2 text-xs" style={{ color: 'var(--text-secondary)' }}>
          <span className="font-medium">Matched at:</span> <code className="font-mono bg-black/5 dark:bg-white/5 px-1 py-0.5 rounded">{finding.matched_at}</code>
        </div>
      )}

      {(finding.extracted_results ?? []).length > 0 && (
        <div className="mb-2">
          <span className="text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>Extracted Results:</span>
          <div className="flex flex-wrap gap-1 mt-1">
            {(finding.extracted_results ?? []).map((r, i) => (
              <code key={i} className="text-xs font-mono bg-black/5 dark:bg-white/5 px-2 py-0.5 rounded" style={{ color: 'var(--text-secondary)' }}>
                {r}
              </code>
            ))}
          </div>
        </div>
      )}

      {finding.curl_command && (
        <details className="mt-2">
          <summary className="text-xs cursor-pointer flex items-center gap-1" style={{ color: 'var(--text-muted)' }}>
            <Terminal className="h-3 w-3" />
            cURL reproduction command
          </summary>
          <pre className="mt-1 p-2 rounded text-xs font-mono overflow-x-auto bg-black/5 dark:bg-white/5"
            style={{ color: 'var(--text-secondary)' }}>
            {finding.curl_command}
          </pre>
        </details>
      )}
    </div>
  );
}
