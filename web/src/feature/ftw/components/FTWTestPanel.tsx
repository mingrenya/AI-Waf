import { useState } from 'react'
import { Play, Loader2, Check, X, Shield, AlertTriangle, FileText } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { ftwApi, FTWResult, FTWReport } from '@/api/ftw'

export function FTWTestPanel() {
  const queryClient = useQueryClient()
  const [targetUrl, setTargetUrl] = useState('http://localhost:8080')

  const { data: files } = useQuery({
    queryKey: ['ftw-files'],
    queryFn: () => ftwApi.files(),
  })

  const { data: history, isPending: histLoading } = useQuery({
    queryKey: ['ftw-reports'],
    queryFn: () => ftwApi.reports(10),
  })

  const runMutation = useMutation({
    mutationFn: () => ftwApi.run(targetUrl),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ftw-reports'] })
    },
  })

  const [selectedReport, setSelectedReport] = useState<FTWReport | null>(null)

  return (
    <Card className="flex flex-col h-full p-0 border-none shadow-none">
      {/* Header */}
      <div className="p-4 border-b dark:border-muted-foreground/20 space-y-3 flex-shrink-0">
        <div className="flex gap-2 items-end">
          <div className="flex-1">
            <label className="text-xs text-muted-foreground">Target URL</label>
            <Input
              value={targetUrl}
              onChange={(e) => setTargetUrl(e.target.value)}
              className="h-8 text-sm font-mono dark:bg-background"
              placeholder="http://localhost:8080"
            />
          </div>
          <Button
            onClick={() => runMutation.mutate()}
            disabled={runMutation.isPending}
            size="sm"
            className="h-8 gap-1"
          >
            {runMutation.isPending ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Play className="h-4 w-4" />
            )}
            Run FTW Tests
          </Button>
        </div>

        {/* Test files */}
        {files && (
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <FileText className="h-3 w-3" />
            Test files: {files.join(', ')}
          </div>
        )}

        {/* Latest run result */}
        {runMutation.data && (
          <RunSummary report={runMutation.data} />
        )}
      </div>

      {/* Body: History + Detail */}
      <div className="flex-1 overflow-auto">
        {selectedReport ? (
          <ReportDetail report={selectedReport} onBack={() => setSelectedReport(null)} />
        ) : (
          <div className="p-4">
            <h3 className="text-sm font-medium mb-3 text-muted-foreground">Test History</h3>
            {histLoading ? (
              <div className="text-center py-8 text-muted-foreground">Loading...</div>
            ) : !history || history.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                No test reports yet. Run your first test above.
              </div>
            ) : (
              <div className="space-y-2">
                {history.map((r) => (
                  <Card
                    key={r.id}
                    className="p-3 cursor-pointer hover:bg-muted/50 transition-colors"
                    onClick={() => setSelectedReport(r)}
                  >
                    <div className="flex items-center justify-between">
                      <div>
                        <span className="text-sm font-medium">{r.targetUrl}</span>
                        <span className="text-xs text-muted-foreground ml-2">
                          {new Date(r.createdAt).toLocaleString()}
                        </span>
                      </div>
                      <div className="flex items-center gap-2">
                        <Badge variant={r.blockRate >= 90 ? 'default' : r.blockRate >= 50 ? 'secondary' : 'destructive'}>
                          {r.blockRate.toFixed(0)}% block
                        </Badge>
                        <span className="text-xs text-success">{r.passed} passed</span>
                        <span className="text-xs text-red-600">{r.failed} failed</span>
                      </div>
                    </div>
                  </Card>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </Card>
  )
}

function RunSummary({ report }: { report: FTWReport }) {
  return (
    <div className="flex items-center gap-3 text-xs bg-muted/30 rounded-lg p-2">
      <span className="font-medium">Latest:</span>
      <Badge variant="outline">{report.totalTests} tests</Badge>
      <span className="text-success flex items-center gap-0.5"><Check className="h-3 w-3" />{report.passed}</span>
      <span className="text-red-600 flex items-center gap-0.5"><X className="h-3 w-3" />{report.failed}</span>
      <span className="flex items-center gap-0.5"><Shield className="h-3 w-3" />{report.blockRate.toFixed(0)}%</span>
      {report.falseNegs > 0 && (
        <span className="text-orange-600 flex items-center gap-0.5">
          <AlertTriangle className="h-3 w-3" />{report.falseNegs} FN
        </span>
      )}
      <span className="text-muted-foreground ml-auto">{report.durationSec.toFixed(1)}s</span>
    </div>
  )
}

function ReportDetail({ report, onBack }: { report: FTWReport; onBack: () => void }) {
  const results = report.results || []

  return (
    <div className="p-4">
      <Button variant="ghost" size="sm" onClick={onBack} className="mb-3">&larr; Back to history</Button>

      <div className="grid grid-cols-6 gap-3 mb-4">
        <StatBox label="Total" value={report.totalTests} />
        <StatBox label="Passed" value={report.passed} color="text-success" />
        <StatBox label="Failed" value={report.failed} color="text-red-600" />
        <StatBox label="Block Rate" value={`${report.blockRate.toFixed(0)}%`} />
        <StatBox label="False Neg" value={report.falseNegs} color="text-orange-600" />
        <StatBox label="Duration" value={`${report.durationSec.toFixed(1)}s`} />
      </div>

      <div className="space-y-1">
        {results.map((r: FTWResult, i: number) => (
          <div key={i} className={`flex items-center gap-3 px-3 py-2 rounded text-xs ${r.passed ? 'bg-success/5 dark:bg-success/10' : 'bg-red-50 dark:bg-red-900/10'}`}>
            <span>{r.passed ? <Check className="h-3.5 w-3.5 text-success" /> : <X className="h-3.5 w-3.5 text-red-600" />}</span>
            <span className="w-14 text-muted-foreground font-mono">#{r.testId}</span>
            <span className="flex-1 truncate">{r.title}</span>
            <Badge variant="outline" className="text-[10px]">{r.statusCode}</Badge>
            {r.wafHit ? (
              <Badge variant="default" className="text-[10px] bg-success">WAF HIT</Badge>
            ) : (
              <Badge variant="outline" className="text-[10px] text-red-600">MISS</Badge>
            )}
            {r.error && <span className="text-red-500 truncate max-w-[120px]">{r.error}</span>}
            <span className="text-muted-foreground">{r.durationMs.toFixed(0)}ms</span>
          </div>
        ))}
      </div>
    </div>
  )
}

function StatBox({ label, value, color }: { label: string; value: string | number; color?: string }) {
  return (
    <Card className="p-3 text-center">
      <div className={`text-lg font-bold ${color || ''}`}>{value}</div>
      <div className="text-[10px] text-muted-foreground">{label}</div>
    </Card>
  )
}
