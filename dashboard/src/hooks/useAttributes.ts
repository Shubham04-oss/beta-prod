import useSWR from 'swr';
import { fetchAPI } from '@/lib/api';

export interface Attribute {
  id: string;
  name: string;
  slug: string;
  type: string;
  created_at: string;
  updated_at: string;
}

export function useAttributes() {
  const { data, error, isLoading, mutate } = useSWR<Attribute[]>('/api/v1/pim/attributes', fetchAPI as any);

  const createAttribute = async (data: Partial<Attribute>) => {
    const res = await fetchAPI('/api/v1/pim/attributes', {
      method: 'POST',
      body: JSON.stringify({
        name: data.name,
        slug: data.slug,
        type: data.type || "TEXT",
      }),
      headers: {
        'Content-Type': 'application/json',
      },
    });
    mutate();
    return res;
  };

  return {
    attributes: data || [],
    isLoading,
    error,
    createAttribute,
  };
}
