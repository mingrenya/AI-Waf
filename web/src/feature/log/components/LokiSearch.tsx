import { useState, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Search, Terminal, Clock, FilterX } from 'lucide-react'
import { logApi, LokiLogEntry } from '@/api/log'


const TIME_RANGES: Array<{ start: string; labelKey: 'lokiSearch.last1h' | 'lokiSearch.last6h' | 'lokiSearch.last24h' }> = [
  { start: '1h', labelKey: 'lokiSearch.last1h' },
  { start: '6h', labelKey: 'lokiSearch.last6h' },
  { start: '24h', labelKey: 'lokiSearch.last24h' },
]

const LEVEL_COLORS: Record<string, string> = {
  error: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300',
  warn: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300',
  info: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300',
  debug: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-300',
}

interface LokiSearchProps {
  defaultQuery?: string
}

export function LokiSearch({ defaultQuery }: LokiSearchProps) {
  const { t } = useTranslation()
  const [query, setQuery] = useState(defaultQuery || '{container_name="mrya-waf"}')
  const [timeRange, setTimeRange] = useState<string>('1h')
  const [level, setLevel] = useState('all')
  const [submittedQuery, setSubmittedQuery] = useState(query)

  const buildQuery = useCallback(() => {
    let q = query
    if (level !== 'all') {
      q += ` |= "${level}"`
    }
    return q
  }, [query, level])

  const { data, isPending, isError } = useQuery({
    queryKey: ['loki-search', submittedQuery, timeRange, level],
    queryFn: () => logApi.lokiQuery(buildQuery(), 100, timeRange, 'now'),
    enabled: !!submittedQuery,
    refetchOnMount: false,
  })

  const handleSearch = () => {
    setSubmittedQuery(buildQuery())
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') handleSearch()
  }

  // Auto-scroll: only show last 50 results
  const displayResults = data?.results?.slice(-100) || []

  return (
    <Card className="flex flex-col h-full p-0 border-none shadow-none">
      {/* Search Bar */}
      <div className="p-4 border-b dark:border-muted-foreground/20 space-y-3 flex-shrink-0">
        {/* Row 1: LogQL input + Search button */}
        <div className="flex gap-2">
          <div className="relative flex-1">
            <Terminal className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={t('lokiSearch.searchPlaceholder')}
              className="pl-9 font-mono text-sm h-9 dark:bg-background dark:border-muted-foreground/30"
            />
          </div>
          <Button onClick={handleSearch} size="sm" className="h-9 gap-1">
            <Search className="h-4 w-4" />
            {t('lokiSearch.search')}
          </Button>
        </div>

        {/* Row 2: Filters */}
        <div className="flex items-center gap-3 flex-wrap">
          {/* Presets */}
          <span className="text-xs text-muted-foreground">{t('lokiSearch.preset')}:</span>
          <Badge
            variant={query === '{container_name="mrya-waf"} |= "error"' ? 'default' : 'outline'}
            className="cursor-pointer text-xs"
            onClick={() => { setQuery('{container_name="mrya-waf"} |= "error"'); setLevel('all') }}
          >
            {t('lokiSearch.presetErrors')}
          </Badge>
          <Badge
            variant={query === '{container_name="mrya-waf"} |= "traffic-analyzer"' ? 'default' : 'outline'}
            className="cursor-pointer text-xs"
            onClick={() => { setQuery('{container_name="mrya-waf"} |= "traffic-analyzer"'); setLevel('all') }}
          >
            {t('lokiSearch.presetTraffic')}
          </Badge>
          <Badge
            variant={query === '{container_name="mrya-waf"} |= "blocked" |= "interrupt"' ? 'default' : 'outline'}
            className="cursor-pointer text-xs"
            onClick={() => { setQuery('{container_name="mrya-waf"} |= "blocked" |= "interrupt"'); setLevel('all') }}
          >
            {t('lokiSearch.presetBlocks')}
          </Badge>
          <Badge
            variant={query === '{container_name="mrya-waf"}' ? 'default' : 'outline'}
            className="cursor-pointer text-xs"
            onClick={() => { setQuery('{container_name="mrya-waf"}'); setLevel('all') }}
          >
            {t('lokiSearch.presetAll')}
          </Badge>

          <div className="w-px h-5 bg-border" />

          {/* Level filter */}
          <Select value={level} onValueChange={setLevel}>
            <SelectTrigger className="w-[110px] h-7 text-xs">
              <SelectValue placeholder={t('lokiSearch.level')} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">{t('lokiSearch.allLevels')}</SelectItem>
              <SelectItem value="error">{t('lokiSearch.error')}</SelectItem>
              <SelectItem value="warn">{t('lokiSearch.warn')}</SelectItem>
              <SelectItem value="info">{t('lokiSearch.info')}</SelectItem>
              <SelectItem value="debug">{t('lokiSearch.debug')}</SelectItem>
            </SelectContent>
          </Select>

          {/* Time range */}
          <Select value={timeRange} onValueChange={setTimeRange}>
            <SelectTrigger className="w-[120px] h-7 text-xs">
              <Clock className="h-3 w-3 mr-1" />
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {TIME_RANGES.map((tr) => (
                <SelectItem key={tr.start} value={tr.start}>{t(tr.labelKey)}</SelectItem>
              ))}
            </SelectContent>
          </Select>

          {/* Reset */}
          <Button
            variant="ghost"
            size="sm"
            className="h-7 text-xs"
            onClick={() => { setQuery('{container_name="mrya-waf"}'); setLevel('all'); setTimeRange('1h') }}
          >
            <FilterX className="h-3 w-3 mr-1" />
            {t('reset')}
          </Button>

          {/* Result count */}
          {data && (
            <span className="text-xs text-muted-foreground ml-auto">
              {t('lokiSearch.totalHits')}: {data.totalHits}
            </span>
          )}
        </div>
      </div>

      {/* Results */}
      <div className="flex-1 overflow-auto font-mono text-sm">
        {isPending && (
          <div className="flex items-center justify-center h-32 text-muted-foreground">
            <Terminal className="h-5 w-5 animate-pulse mr-2" />
            Searching...
          </div>
        )}
        {isError && (
          <div className="flex items-center justify-center h-32 text-red-500">
            Query failed. Check Loki connection.
          </div>
        )}
        {!isPending && !isError && displayResults.length === 0 && (
          <div className="flex items-center justify-center h-32 text-muted-foreground">
            {t('lokiSearch.noResults')}
          </div>
        )}
        {displayResults.map((entry, i) => (
          <LogLine key={`${entry.timestamp}-${i}`} entry={entry} />
        ))}
      </div>
    </Card>
  )
}

function LogLine({ entry }: { entry: LokiLogEntry }) {
  const level = entry.level?.toLowerCase() || ''
  const badgeColor = LEVEL_COLORS[level] || ''
  const timestamp = entry.timestamp.length > 19
    ? new Date(parseInt(entry.timestamp.slice(0, 13))).toLocaleString()
    : new Date(entry.timestamp).toLocaleString()

  return (
    <div className="px-4 py-1.5 border-b dark:border-muted-foreground/10 hover:bg-muted/50 transition-colors flex gap-3 items-start">
      <span className="text-xs text-muted-foreground whitespace-nowrap shrink-0 pt-0.5">
        {timestamp}
      </span>
      {level && (
        <Badge variant="outline" className={`text-[10px] px-1 py-0 h-4 shrink-0 ${badgeColor}`}>
          {level.toUpperCase()}
        </Badge>
      )}
      {entry.component && (
        <span className="text-xs text-muted-foreground shrink-0 pt-0.5">
          [{entry.component}]
        </span>
      )}
      <span className="text-xs break-all leading-relaxed dark:text-gray-200">
        {entry.message}
      </span>
    </div>
  )
}
