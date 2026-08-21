import { useCallback, useEffect, useMemo, useState } from "react";
import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useSearchParams } from "react-router";
import { logSearchService } from "@/features/log-explorer/services/log-search.service";
import type { SearchModel, SearchRule } from "@/features/log-explorer/services/log-search.service";
import { productService } from "@/services/product.service";
import { getRecordId } from "@/features/log-explorer/utils/log-utils";
import { useAuthStore } from "@/store";
import { getSavedUserColumns, useLogExplorerStore } from "@/features/log-explorer/store/useLogExplorerStore";
import { useLogExplorerFavorites } from "./useLogExplorerFavorites";

function comparableRules(values: SearchRule[]): string {
  return JSON.stringify(
    values.map(({ field, type, operator, value, enabled }) => ({
      field,
      type,
      operator,
      value,
      enabled,
    })),
  );
}

export function useLogExplorerController() {
  const [searchParams] = useSearchParams();
  const currentUser = useAuthStore((state) => state.currentUser);
  const userId = currentUser?.id;
  const routeEnvironmentId = Number(searchParams.get("environment_id")) || undefined;
  const archiveId = searchParams.get("archive_id") ?? undefined;
  const routeTimeRange = searchParams.get("time_range") || "24h";
  const [query, setQueryState] = useState("");
  const [submittedQuery, setSubmittedQuery] = useState("");
  const [productId, setProductId] = useState<number>();
  const [projectId, setProjectIdState] = useState<number>();
  const [environmentId, setEnvironmentIdState] = useState<number | undefined>(routeEnvironmentId);
  const [categoryId, setCategoryIdState] = useState<number>();
  const [subFeatureId, setSubFeatureIdState] = useState<number>();
  const [timeRange, setTimeRangeState] = useState(routeTimeRange);
  const [rules, setRulesState] = useState<SearchRule[]>([]);
  const [appliedProjectId, setAppliedProjectId] = useState<number>();
  const [appliedEnvironmentId, setAppliedEnvironmentId] = useState<number | undefined>(routeEnvironmentId);
  const [appliedCategoryId, setAppliedCategoryId] = useState<number>();
  const [appliedSubFeatureId, setAppliedSubFeatureId] = useState<number>();
  const [appliedTimeRange, setAppliedTimeRange] = useState(routeTimeRange);
  const [appliedRules, setAppliedRules] = useState<SearchRule[]>([]);
  const [searchRevision, setSearchRevision] = useState(0);
  const [autoRefreshInterval, setAutoRefreshInterval] = useState(5000);
  const explorer = useLogExplorerStore();

  useEffect(() => {
    const userColumns = getSavedUserColumns(userId);
    const currentColumns = useLogExplorerStore.getState().columns;
    if (JSON.stringify(currentColumns) !== JSON.stringify(userColumns)) {
      explorer.setColumns(userColumns);
    }
  }, [userId, explorer]);

  const { select, selectedId, setFiltersOpen } = explorer;
  const productsQuery = useQuery({
    queryKey: ["products", "options"],
    queryFn: productService.listOptions,
  });
  const products = useMemo(() => [...(productsQuery.data ?? [])], [productsQuery.data]);
  const routeProductId = useMemo(() => {
    const value =
      searchParams.get("product") ??
      searchParams.get("product_id") ??
      searchParams.get("productId");
    if (!value) return undefined;
    return products.find(
      (product) =>
        String(product.id) === value ||
        product.product_code?.toLowerCase() === value.toLowerCase() ||
        product.name.toLowerCase() === value.toLowerCase(),
    )?.id;
  }, [products, searchParams]);
  const selectedProductId = productId ?? routeProductId ?? (productsQuery.data?.length ? productsQuery.data[0].id : products[0]?.id);
  const { favoriteFields, reorderFavoriteFields } = useLogExplorerFavorites(
    selectedProductId,
    explorer.columns,
    explorer.setColumns,
  );
  const projectsQuery = useQuery({
    queryKey: ["products", selectedProductId, "projects"],
    queryFn: () => productService.listProjects(selectedProductId!),
    enabled: Boolean(selectedProductId),
    staleTime: 30_000,
  });
  const environmentsQuery = useQuery({
    queryKey: ["products", selectedProductId, "environments"],
    queryFn: () => productService.listEnvironments(selectedProductId!),
    enabled: Boolean(selectedProductId),
    staleTime: 30_000,
  });
  const featuresQuery = useQuery({
    queryKey: ["products", selectedProductId, "features", projectId],
    queryFn: () => productService.listFeatures(selectedProductId!, projectId),
    enabled: Boolean(selectedProductId),
    staleTime: 30_000,
  });
  const projects = useMemo(() => projectsQuery.data?.length ? projectsQuery.data : [], [projectsQuery.data]);
  const environments = useMemo(() => environmentsQuery.data?.length ? environmentsQuery.data : [], [environmentsQuery.data]);
  const features = useMemo(() => featuresQuery.data?.length ? featuresQuery.data : [], [featuresQuery.data]);

  const categories = useMemo(() => {
    return features.filter((f) => !f.parent_id || f.level === 1);
  }, [features]);

  const subFeatures = useMemo(() => {
    if (categoryId) {
      const catIdStr = String(categoryId);
      return features.filter((f) => {
        if (f.category_id === categoryId) return false;
        if (f.parent_id === categoryId) return true;
        if (f.path_ids) {
          const ids = f.path_ids.split(",");
          return ids.includes(catIdStr);
        }
        return false;
      });
    }
    return features.filter((f) => f.parent_id != null && f.level > 1);
  }, [features, categoryId]);

  const setProjectId = useCallback((val?: number) => {
    setProjectIdState(val);
    setAppliedProjectId(val);
    setCategoryIdState(undefined);
    setSubFeatureIdState(undefined);
    setAppliedCategoryId(undefined);
    setAppliedSubFeatureId(undefined);
    setSearchRevision((current) => current + 1);
  }, []);

  const setEnvironmentId = useCallback((val?: number) => {
    setEnvironmentIdState(val);
    setAppliedEnvironmentId(val);
    setSearchRevision((current) => current + 1);
  }, []);

  const setCategoryId = useCallback(
    (val?: number) => {
      setCategoryIdState(val);
      setAppliedCategoryId(val);
      if (val && subFeatureId) {
        const match = features.find((f) => f.category_id === subFeatureId);
        if (match && match.parent_id !== val) {
          setSubFeatureIdState(undefined);
          setAppliedSubFeatureId(undefined);
        }
      }
      setSearchRevision((current) => current + 1);
    },
    [subFeatureId, features],
  );

  const setSubFeatureId = useCallback(
    (val?: number) => {
      setSubFeatureIdState(val);
      setAppliedSubFeatureId(val);
      if (val) {
        const match = features.find((f) => f.category_id === val);
        if (match && match.parent_id) {
          setCategoryIdState(match.parent_id);
          setAppliedCategoryId(match.parent_id);
        }
      }
      setSearchRevision((current) => current + 1);
    },
    [features],
  );

  const setTimeRange = useCallback((val: string) => {
    setTimeRangeState(val);
    setAppliedTimeRange(val);
    setSearchRevision((current) => current + 1);
  }, []);

  const setRules = useCallback(
    (newRules: SearchRule[] | ((prev: SearchRule[]) => SearchRule[])) => {
      setRulesState((prev) => {
        const next = typeof newRules === "function" ? newRules(prev) : newRules;
        setAppliedRules(next.map((rule) => ({ ...rule })));
        return next;
      });
      setSearchRevision((current) => current + 1);
    },
    [],
  );

  const setQuery = useCallback((val: string) => {
    setQueryState(val);
  }, []);

  const effectiveRules = useMemo(() => {
    const rest = appliedRules.filter((rule) => rule.id !== "natural-search");
    return submittedQuery
      ? [
          ...rest,
          {
            id: "natural-search",
            field: "global_search",
            label: "Search",
            type: "text" as const,
            operator: "contains" as const,
            value: submittedQuery,
            enabled: true,
          },
        ]
      : rest;
  }, [appliedRules, submittedQuery]);
  const model = useMemo<SearchModel>(
    () => ({
      version: 1,
      scope: {
        product_id: selectedProductId ?? 0,
        project_id: appliedProjectId,
        environment_id: appliedEnvironmentId,
        category_id: appliedSubFeatureId ?? appliedCategoryId,
        archive_id: archiveId,
      },
      time_range: {
        field: "@timestamp",
        type: "relative",
        value: appliedTimeRange,
      },
      root: {
        type: "group",
        operator: "AND",
        children: effectiveRules.map(
          ({ field, type, operator, value, enabled }) => ({
            type: "condition",
            field,
            data_type: type,
            operator,
            value,
            enabled,
          }),
        ),
      },
      page_size: 100,
    }),
    [
      appliedCategoryId,
      appliedEnvironmentId,
      appliedProjectId,
      appliedSubFeatureId,
      appliedTimeRange,
      effectiveRules,
      selectedProductId,
      archiveId,
    ],
  );
  const logsQuery = useInfiniteQuery({
    queryKey: [
      "logs",
      "explorer",
      selectedProductId,
      appliedProjectId,
      appliedEnvironmentId,
      appliedCategoryId,
      appliedSubFeatureId,
      appliedTimeRange,
      archiveId,
      effectiveRules,
      searchRevision,
    ],
    queryFn: ({ pageParam, signal }) => logSearchService.execute(model, pageParam, 100, signal),
    initialPageParam: 1,
    getNextPageParam: (last, pages) =>
      pages.flatMap((page) => page.records).length < last.total
        ? pages.length + 1
        : undefined,
    enabled: selectedProductId !== undefined && selectedProductId !== -1,
    staleTime: autoRefreshInterval > 0 ? 0 : 30_000,
    // Pause auto-refresh while the user has an active search query
    refetchInterval: submittedQuery
      ? false
      : autoRefreshInterval > 0
        ? autoRefreshInterval
        : false,
  });
  const records = useMemo(() => {
    const apiRecords = logsQuery.data?.pages.flatMap((page) => page.records) ?? [];
    const productNames = new Map(products.map((product) => [product.id, product.name]));
    return apiRecords.map((record) => ({
      ...record,
      product_name: record.product_name ?? productNames.get(Number(record.product_id)) ?? productNames.get(selectedProductId ?? 0),
    }));
  }, [logsQuery.data, products, selectedProductId]);
  const total = logsQuery.data?.pages[0]?.total ?? records.length;

  useEffect(() => {
    if (!selectedId && records[0]) {
      select(getRecordId(records[0], 0));
    }
  }, [records, select, selectedId]);
  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        document
          .querySelector<HTMLInputElement>(
            '.log-search-panel input[aria-label="Search logs"]',
          )
          ?.focus();
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  const changeProduct = useCallback((nextProductId: number) => {
    setProductId(nextProductId);
    setProjectIdState(undefined);
    setEnvironmentIdState(undefined);
    setCategoryIdState(undefined);
    setSubFeatureIdState(undefined);
    setAppliedProjectId(undefined);
    setAppliedEnvironmentId(undefined);
    setAppliedCategoryId(undefined);
    setAppliedSubFeatureId(undefined);
    setRulesState([]);
    setAppliedRules([]);
    setQueryState("");
    setSubmittedQuery("");
    setSearchRevision((current) => current + 1);
  }, []);
  const filtersDirty =
    query.trim() !== submittedQuery ||
    projectId !== appliedProjectId ||
    environmentId !== appliedEnvironmentId ||
    categoryId !== appliedCategoryId ||
    subFeatureId !== appliedSubFeatureId ||
    timeRange !== appliedTimeRange ||
    comparableRules(rules) !== comparableRules(appliedRules);
  const applyFilters = useCallback(() => {
    setSubmittedQuery(query.trim());
    setAppliedProjectId(projectId);
    setAppliedEnvironmentId(environmentId);
    setAppliedCategoryId(categoryId);
    setAppliedSubFeatureId(subFeatureId);
    setAppliedTimeRange(timeRange);
    setAppliedRules(rules.map((rule) => ({ ...rule })));
    setSearchRevision((current) => current + 1);
  }, [categoryId, environmentId, projectId, query, rules, subFeatureId, timeRange]);
  const clearFilters = useCallback(() => {
    setQueryState("");
    setSubmittedQuery("");
    setProjectIdState(undefined);
    setEnvironmentIdState(undefined);
    setCategoryIdState(undefined);
    setSubFeatureIdState(undefined);
    setAppliedProjectId(undefined);
    setAppliedEnvironmentId(undefined);
    setAppliedCategoryId(undefined);
    setAppliedSubFeatureId(undefined);
    setTimeRangeState("all");
    setAppliedTimeRange("all");
    setRulesState([]);
    setAppliedRules([]);
    setSearchRevision((current) => current + 1);
  }, []);
  const selectedRecord =
    records.find(
      (record, index) => getRecordId(record, index) === selectedId,
    ) ?? null;
  const addRule = useCallback(
    (rule: SearchRule) => {
      setRules((current) => [
        ...current.filter((item) => item.field !== rule.field),
        rule,
      ]);
      setFiltersOpen(true);
    },
    [setFiltersOpen, setRules],
  );
  const loadMore = useCallback(() => {
    if (logsQuery.hasNextPage && !logsQuery.isFetchingNextPage) {
      void logsQuery.fetchNextPage();
    }
  }, [logsQuery]);

  return {
    query,
    submittedQuery,
    selectedProductId,
    projectId,
    environmentId,
    categoryId,
    subFeatureId,
    timeRange,
    rules,
    products,
    projects,
    environments,
    categories,
    subFeatures,
    favoriteFields,
    records,
    total,
    selectedRecord,
    filtersDirty,
    autoRefreshInterval,
    explorer,
    isLoading: logsQuery.isLoading,
    isFetching: logsQuery.isFetching,
    isFetchingNextPage: logsQuery.isFetchingNextPage,
    hasNextPage: Boolean(logsQuery.hasNextPage),
    isError: logsQuery.isError,
    setQuery,
    setProjectId,
    setEnvironmentId,
    setCategoryId,
    setSubFeatureId,
    setTimeRange,
    setRules,
    setAutoRefreshInterval,
    changeProduct,
    applyFilters,
    clearFilters,
    reorderFavoriteFields,
    addRule,
    loadMore,
    refresh: () => void logsQuery.refetch(),
  };
}

export type LogExplorerController = ReturnType<
  typeof useLogExplorerController
>;
