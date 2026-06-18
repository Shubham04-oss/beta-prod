import useSWR from 'swr';
import { fetchAPI } from '@/lib/api';

export interface Brand {
  id: string;
  name: string;
  description: string;
  logo_url: string;
  created_at: string;
  updated_at: string;
}

export function useBrands() {
  const { data, error, isLoading, mutate } = useSWR<Brand[]>('/api/v1/pim/brands', fetchAPI as any);

  const createBrand = async (data: Partial<Brand>) => {
    const payload = {
      name: data.name,
      description: data.description,
      logoUrl: data.logo_url
    };
    const res = await fetchAPI('/api/v1/pim/brands', {
      method: 'POST',
      body: JSON.stringify(payload),
      headers: {
        'Content-Type': 'application/json',
      },
    });
    mutate();
    return res;
  };

  return {
    brands: data || [],
    isLoading,
    error,
    createBrand,
  };
}
