import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { LokiSearch } from '@/feature/log/components/LokiSearch'
import { LokiStatsPanel } from '@/feature/log/components/LokiStatsPanel'
import { BarChart3, Search } from 'lucide-react'

export default function EngineLogPage() {
  const { t } = useTranslation()
  const [view, setView] = useState<'search' | 'stats'>('search')

  return (
    <div className="flex flex-col h-full">
      {/* Toggle */}
      <div className="flex items-center gap-2 px-4 py-2 border-b dark:border-muted-foreground/20">
        <Button
          variant={view === 'search' ? 'default' : 'ghost'}
          size="sm"
          className="h-7 text-xs gap-1"
          onClick={() => setView('search')}
        >
          <Search className="h-3 w-3" />
          {t('lokiSearch.search')}
        </Button>
        <Button
          variant={view === 'stats' ? 'default' : 'ghost'}
          size="sm"
          className="h-7 text-xs gap-1"
          onClick={() => setView('stats')}
        >
          <BarChart3 className="h-3 w-3" />
          Stats
        </Button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto">
        {view === 'search' ? <LokiSearch /> : <LokiStatsPanel />}
      </div>
    </div>
  )
}
