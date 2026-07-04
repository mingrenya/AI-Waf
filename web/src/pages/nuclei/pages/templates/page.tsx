import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { listTemplates } from '@/api/nuclei';
import type { TemplateInfo } from '@/types/nuclei';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { FileCode, Search, Loader2, Folder } from 'lucide-react';

export default function TemplatesPage() {
  const [search, setSearch] = useState('');

  const { data: templates = [], isLoading } = useQuery({
    queryKey: ['nuclei-templates'],
    queryFn: () => listTemplates().then((res) => (res as unknown as { data?: TemplateInfo[] })?.data ?? []),
  });

  const filtered = templates.filter((t) =>
    !search || t.name.toLowerCase().includes(search.toLowerCase()) || t.path.toLowerCase().includes(search.toLowerCase())
  );

  // 按目录分组
  const grouped = groupByDir(filtered);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64 text-sm" style={{ color: 'var(--text-muted)' }}>
        <Loader2 className="h-4 w-4 animate-spin mr-2" />
        Loading templates...
      </div>
    );
  }

  return (
    <div className="p-4">
      <Card className="surface-card">
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <span className="flex items-center gap-2">
              <FileCode className="h-5 w-5" style={{ color: 'var(--color-primary-5)' }} />
              Nuclei Templates
              <Badge variant="outline" className="ml-2 text-xs">{filtered.length} templates</Badge>
            </span>
          </CardTitle>
          {/* 搜索框 */}
          <div className="relative mt-2">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4" style={{ color: 'var(--text-muted)' }} />
            <input
              type="text"
              placeholder="Search templates..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-10 pr-4 py-2 rounded-lg border text-sm bg-transparent focus:outline-none focus:ring-1"
              style={{ borderColor: 'var(--border)', color: 'var(--text-primary)', ringColor: 'var(--color-primary-5)' }}
            />
          </div>
        </CardHeader>
        <CardContent>
          {filtered.length === 0 ? (
            <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
              {search ? 'No templates match your search.' : 'No templates available. Run nuclei -ut to download templates.'}
            </p>
          ) : (
            <div className="space-y-4 max-h-[600px] overflow-y-auto scrollbar-custom">
              {Object.entries(grouped).map(([dir, items]) => (
                <div key={dir}>
                  <div className="flex items-center gap-2 mb-2 text-xs" style={{ color: 'var(--text-muted)' }}>
                    <Folder className="h-3.5 w-3.5" />
                    <span className="font-mono">{dir}</span>
                    <span className="ml-auto">{items.length} files</span>
                  </div>
                  <div className="flex flex-wrap gap-1.5">
                    {items.map((t) => (
                      <Badge
                        key={t.path}
                        variant="outline"
                        className="text-xs cursor-default font-mono"
                        title={t.path}
                      >
                        {t.name}
                      </Badge>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function groupByDir(templates: TemplateInfo[]): Record<string, TemplateInfo[]> {
  const groups: Record<string, TemplateInfo[]> = {};
  for (const t of templates) {
    const parts = t.path.split('/');
    const dir = parts.length > 1 ? parts.slice(0, -1).join('/') : '(root)';
    if (!groups[dir]) groups[dir] = [];
    groups[dir].push(t);
  }
  return groups;
}
