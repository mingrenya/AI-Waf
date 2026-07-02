import { useQuery } from '@tanstack/react-query';
import { getOverview, listChains, listAttackers, getTrends, getAttackerProfile } from '@/api/situation';

export function useOverview() {
  return useQuery({
    queryKey: ['situation', 'overview'],
    queryFn: () => getOverview().then(r => r.data.data),
    refetchInterval: 30000,
  });
}

export function useChains(params: Record<string, unknown> = {}) {
  return useQuery({
    queryKey: ['situation', 'chains', params],
    queryFn: () =>
      listChains(params as Record<string, unknown> & { page?: number; page_size?: number }).then(
        r => r.data.data,
      ),
  });
}

export function useAttackers(params: Record<string, unknown> = {}) {
  return useQuery({
    queryKey: ['situation', 'attackers', params],
    queryFn: () =>
      listAttackers(params as Record<string, unknown> & { page?: number; page_size?: number }).then(
        r => r.data.data,
      ),
  });
}

export function useTrends(duration = '24h') {
  return useQuery({
    queryKey: ['situation', 'trends', duration],
    queryFn: () => getTrends(duration).then(r => r.data.data),
    refetchInterval: 60000,
  });
}

export function useAttackerProfile(ip: string) {
  return useQuery({
    queryKey: ['situation', 'attacker', ip],
    queryFn: () => getAttackerProfile(ip).then(r => r.data.data),
    enabled: !!ip,
  });
}
