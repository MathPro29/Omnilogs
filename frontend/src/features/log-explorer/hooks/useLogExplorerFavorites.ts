import { useCallback, useEffect, useMemo } from "react";
import { message } from "antd";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  customFieldFavoriteService,
  type FavoriteField,
} from "@/features/log-explorer/services/custom-field-favorite.service";
import { useAuthStore } from "@/store";
import {
  favoriteLogColumnKey,
  getUserFavoriteIds,
  isFavoriteLogColumn,
  saveUserFavoriteIds,
  type LogColumn,
} from "@/features/log-explorer/store/useLogExplorerStore";

export function useLogExplorerFavorites(
  productId: number | undefined,
  columns: LogColumn[],
  setColumns: (columns: LogColumn[]) => void,
) {
  const currentUser = useAuthStore((state) => state.currentUser);
  const userId = currentUser?.id;
  const queryClient = useQueryClient();
  const query = useQuery({
    queryKey: ["custom-fields", "favorites", productId],
    queryFn: () => customFieldFavoriteService.list(productId!),
    enabled: Boolean(productId),
  });
  const allFavoriteFields = useMemo(() => query.data ?? [], [query.data]);

  const favoriteFields = useMemo(() => {
    if (!userId || !productId || !query.isSuccess) return [];
    const userFavIds = getUserFavoriteIds(userId, productId);
    if (userFavIds === null) {
      // Initialize with empty array for new users so other users' favorites don't spill over
      saveUserFavoriteIds(userId, productId, []);
      return [];
    }
    const favSet = new Set(userFavIds);
    return allFavoriteFields.filter((field) => favSet.has(field.field_definition_id));
  }, [allFavoriteFields, userId, productId, query.isSuccess]);

  useEffect(() => {
    if (!query.isSuccess) return;
    const validColumns = new Set(
      favoriteFields.map((field) =>
        favoriteLogColumnKey(field.field_definition_id),
      ),
    );
    const nextColumns = columns.filter(
      (column) => !isFavoriteLogColumn(column) || validColumns.has(column),
    );
    if (nextColumns.length !== columns.length) setColumns(nextColumns);
  }, [columns, favoriteFields, query.isSuccess, setColumns]);

  const reorderFavoriteFields = useCallback(
    (fieldDefinitionIds: number[]) => {
      if (!productId || !fieldDefinitionIds.length) return;
      const queryKey = ["custom-fields", "favorites", productId] as const;
      const current = queryClient.getQueryData<FavoriteField[]>(queryKey);
      const order = new Map(
        fieldDefinitionIds.map((id, index) => [id, index]),
      );
      if (current) {
        queryClient.setQueryData<FavoriteField[]>(
          queryKey,
          [...current]
            .sort(
              (left, right) =>
                (order.get(left.field_definition_id) ??
                  Number.MAX_SAFE_INTEGER) -
                (order.get(right.field_definition_id) ??
                  Number.MAX_SAFE_INTEGER),
            )
            .map((field, index) => ({ ...field, display_order: index })),
        );
      }
      void customFieldFavoriteService
        .reorder({
          product_id: productId,
          items: fieldDefinitionIds.map(
            (field_definition_id, display_order) => ({
              field_definition_id,
              display_order,
            }),
          ),
        })
        .catch(() => {
          void queryClient.invalidateQueries({ queryKey });
          message.error("Could not save custom field order");
        });
    },
    [productId, queryClient],
  );

  return { favoriteFields, reorderFavoriteFields };
}
