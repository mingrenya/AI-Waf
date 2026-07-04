import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { getTaskDetail, cancelTask } from '@/api/nuclei';
import { listTasks } from '@/api/nuclei';
import type { ScanTaskResponse, ScanTaskDetail, NucleiFinding } from '@/types/nuclei';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { X, AlertTriangle, Bug, Shield, Info, Terminal, Loader2, Eye } from 'lucide-react';

const SEVERITY_MAP: Record<string, { color: string; bg: string; icon: typeof Bug }> = {
  critical: { color: 'text-red-400', bg: 'bg-red-500/10 border-red-500/20', icon: AlertTriangle },
  high: { color: 'text-orange-400', bg: 'bg-orange-500/10 border-orange-500/20', icon: AlertTriangle },
  medium: { color: 'text-yellow-400', bg: 'bg-yellow-500/10 border-yellow-500/20', icon: Bug },
  low: { color: 'text-blue-400', bg: 'bg-blue-500/10 border-blue-500/20', icon: Info },
  info: { color: 'text-slate-400', bg: 'bg-slate-500/10 border-slate-500/20', icon: Info },
};

export default function ScanPage() {
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);

  const { data: tasks = [], refetch: refetchTasks, isLoading } = useQuery({
    queryKey: ['nuclei-tasks'],
    queryFn: () => listTasks().then((res) => (res as unknown as { data?: ScanTaskResponse[] })?.data ?? []),
    refetchInterval: 5000,
  });

  const { data: taskDetail, isLoading: detailLoading } = useQuery({
    queryKey: ['nuclei-task-detail', selectedTaskId],
    queryFn: () => getTaskDetail(selectedTaskId!).then((r) => (r as unknown as { data: ScanTaskDetail }).data),
    enabled: !!selectedTaskId,
  });

  const handleCancel = async (id: string) => {
    await cancelTask(id);
    refetchTasks();
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
    <div className="flex flex-col gap-4 p-4 h-full">
      <Card className="surface-card">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Terminal className="h-5 w-5" style={{ color: 'var(--color-primary-5)' }} />
            Scan Tasks
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center gap-2 text-sm" style={{ color: 'var(--text-muted)' }}>
              <Loader2 className="h-4 w-4 animate-spin" />
              Loading tasks...
            </div>
          ) : tasks.length === 0 ? (
            <p className="text-sm" style={{ color: 'var(--text-muted)' }}>No scan tasks yet.</p>
          ) : (
            <div className="space-y-2">
              {tasks.map((task) => (
                <div
                  key={task.id}
                  className={`flex items-center gap-3 p-3 rounded-lg border cursor-pointer transition-colors ${
                    selectedTaskId === task.id
                      ? 'border-primary bg-primary/5'
                      : 'border-transparent hover:border-border'
                  }`}
                  style={{
                    borderColor: selectedTaskId === task.id ? 'var(--color-primary-5)' : 'var(--surface-root-border)',
                    background: selectedTaskId === task.id ? 'var(--surface-card-bg)' : undefined,
                  }}
                  onClick={() => setSelectedTaskId(task.id)}
                >
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium truncate" style={{ color: 'var(--text-primary)' }}>
                        {task.target_url}
                      </span>
                      <Badge variant={statusVariant(task.status)} className="text-xs">
                        {task.status}
                      </Badge>
                    </div>
                    <div className="flex items-center gap-3 mt-0.5 text-xs" style={{ color: 'var(--text-muted)' }}>
                      <span className="font-mono">{task.id.slice(0, 8)}</span>
                      <span>findings: {task.findings}/{task.total}</span>
                      {task.created_at && <span>{new Date(task.created_at).toLocaleString()}</span>}
                    </div>
                  </div>
                  <div className="flex items-center gap-1">
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={(e) => { e.stopPropagation(); setSelectedTaskId(task.id); }}
                    >
                      <Eye className="h-4 w-4" />
                    </Button>
                    {task.status === 'running' && (
                      <Button size="sm" variant="outline" onClick={(e) => { e.stopPropagation(); handleCancel(task.id); }}>
                        <X className="h-4 w-4" />
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* 任务详情 — Findings 列表 */}
      {selectedTaskId && (
        <Card className="surface-card">
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <span className="flex items-center gap-2">
                <Search className="h-5 w-5" style={{ color: 'var(--color-primary-5)' }} />
                Scan Findings
              </span>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setSelectedTaskId(null)}
              >
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
            ) : taskDetail.findings?.length === 0 ? (
              <p className="text-sm" style={{ color: 'var(--text-muted)' }}>No findings from this scan.</p>
            ) : (
              <div className="space-y-3">
                {taskDetail.findings.map((f, idx) => (
                  <FindingCard key={idx} finding={f} index={idx} />
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

      {finding.extracted_results?.length > 0 && (
        <div className="mb-2">
          <span className="text-xs font-medium" style={{ color: 'var(--text-secondary)' }}>Extracted Results:</span>
          <div className="flex flex-wrap gap-1 mt-1">
            {finding.extracted_results.map((r, i) => (
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

// 搜索图标组件
function Search({ className }: { className?: string }) {
  return (
    <svg className={className} width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="11" cy="11" r="8"/>
      <path d="m21 21-4.3-4.3"/>
    </svg>
  );
}
