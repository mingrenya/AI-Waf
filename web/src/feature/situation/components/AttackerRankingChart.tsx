import { useMemo } from 'react';
import {
  useReactTable,
  getCoreRowModel,
  type ColumnDef,
} from '@tanstack/react-table';
import { DataTable } from '@/components/table/data-table';
import { Badge } from '@/components/ui/badge';
import { useAttackers } from '../hooks/useSituationData';
import type { AttackerSummary } from '@/types/situation';

function riskBadge(score: number, label: string): { variant: 'default' | 'destructive' | 'secondary' | 'outline' } {
  if (label === 'critical' || score >= 75) return { variant: 'destructive' };
  if (label === 'high' || score >= 50) return { variant: 'destructive' };
  if (label === 'medium' || score >= 25) return { variant: 'secondary' };
  return { variant: 'outline' };
}

function timeAgo(dateStr: string): string {
  if (!dateStr) return '-';
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

interface AttackerRankingChartProps {
  // onSelectAttacker is reserved for future use (click-to-filter)
  // onSelectAttacker?: (ip: string) => void;
}

export default function AttackerRankingChart(_props: AttackerRankingChartProps) {
  const { data, isLoading } = useAttackers({ page_size: 50, sort_by: 'risk_score' });

  const columns: ColumnDef<AttackerSummary>[] = useMemo(
    () => [
      {
        accessorKey: 'source_ip',
        header: 'IP',
        cell: ({ row }) => (
          <span className="font-mono text-sm">{row.original.source_ip}</span>
        ),
      },
      {
        accessorKey: 'geo_country',
        header: 'Country',
        cell: ({ row }) => (
          <Badge variant="outline">{row.original.geo_country ?? '-'}</Badge>
        ),
      },
      {
        accessorKey: 'total_attacks',
        header: 'Total Attacks',
        cell: ({ row }) => (
          <span className="font-medium">{row.original.total_attacks}</span>
        ),
      },
      {
        accessorKey: 'top_attack_type',
        header: 'Attack Type',
        cell: ({ row }) => (
          <span className="text-sm">{row.original.top_attack_type}</span>
        ),
      },
      {
        accessorKey: 'attack_phase',
        header: 'Phase',
        cell: ({ row }) => (
          <Badge variant="secondary" className="capitalize">
            {row.original.attack_phase?.replace(/_/g, ' ') ?? '-'}
          </Badge>
        ),
      },
      {
        accessorKey: 'risk_score',
        header: 'Risk Score',
        cell: ({ row }) => {
          const badge = riskBadge(row.original.risk_score, row.original.risk_label);
          return (
            <Badge variant={badge.variant}>
              {row.original.risk_score} - {row.original.risk_label ?? '-'}
            </Badge>
          );
        },
      },
      {
        accessorKey: 'last_seen',
        header: 'Last Seen',
        cell: ({ row }) => (
          <span className="text-xs text-muted-foreground">
            {timeAgo(row.original.last_seen)}
          </span>
        ),
      },
    ],
    [],
  );

  const table = useReactTable({
    data: data?.attackers ?? [],
    columns,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
  });

  return (
    <DataTable
      table={table}
      columns={columns}
      style="border"
      isLoading={isLoading}
      loadingRows={5}
      loadingStyle="skeleton"
    />
  );
}
