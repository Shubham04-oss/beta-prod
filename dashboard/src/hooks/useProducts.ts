import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api";
import { Product } from "@/store/pimStore";

export function useGetProducts() {
  return useQuery<Product[]>({
    queryKey: ["products"],
    queryFn: async () => {
      const data = await apiClient.get("api/v1/pim/products").json<Product[]>();
      return data || [];
    },
  });
}

export function useCreateProduct() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (newProduct: { title: string; category?: string; description?: string }) => {
      return await apiClient.post("api/v1/pim/products", { json: newProduct }).json<Product>();
    },
    onSuccess: () => {
      // Invalidate and refetch
      queryClient.invalidateQueries({ queryKey: ["products"] });
    },
  });
}

export function useGetProduct(id: string) {
  return useQuery<Product>({
    queryKey: ["product", id],
    queryFn: async () => {
      return await apiClient.get(`api/v1/pim/products/${id}`).json<Product>();
    },
    enabled: !!id,
  });
}

export function useUpdateProduct() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, ...data }: { id: string; title: string; category?: string; description?: string }) => {
      return await apiClient.put(`api/v1/pim/products/${id}`, { json: data }).json<Product>();
    },
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({ queryKey: ["products"] });
      queryClient.invalidateQueries({ queryKey: ["product", variables.id] });
    },
  });
}

export function useDeleteProduct() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      return await apiClient.delete(`api/v1/pim/products/${id}`).json();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["products"] });
      queryClient.invalidateQueries({ queryKey: ["pimStats"] });
    },
  });
}

export interface TopLowStockProduct {
  product_id: string;
  product_title: string;
  stock_left: number;
}

export interface PIMStats {
  totalProducts: number;
  lowStockVariants: number;
  outOfStockVariants: number;
  totalInventoryValue: number;
  topLowStock: TopLowStockProduct[];
}

export function useGetPIMStats() {
  return useQuery<PIMStats>({
    queryKey: ["pimStats"],
    queryFn: async () => {
      const data = await apiClient.get("api/v1/pim/stats").json<PIMStats>();
      return data || {
        totalProducts: 0,
        lowStockVariants: 0,
        outOfStockVariants: 0,
        totalInventoryValue: 0,
        topLowStock: []
      };
    },
  });
}
