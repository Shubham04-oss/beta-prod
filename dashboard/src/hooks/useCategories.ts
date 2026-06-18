import useSWR from 'swr';
import { fetchAPI } from '@/lib/api';

export interface Category {
  id: string;
  parent_id: string | null;
  name: string;
  slug: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export function useCategories() {
  const { data, error, isLoading, mutate } = useSWR<Category[]>('/api/v1/pim/categories', fetchAPI as any);

  const createCategory = async (data: Partial<Category>) => {
    const res = await fetchAPI('/api/v1/pim/categories', {
      method: 'POST',
      body: JSON.stringify({
        name: data.name,
        slug: data.slug,
        description: data.description,
        parentId: data.parent_id,
      }),
      headers: {
        'Content-Type': 'application/json',
      },
    });
    mutate();
    return res;
  };

  return {
    categories: data || [],
    isLoading,
    error,
    createCategory,
  };
}
