import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { retentionPolicyService } from "@/features/retention-policies/services/retentionPolicy.service";
import type {
  CreateRetentionPolicyPayload,
  UpdateRetentionPolicyPayload,
  RetentionMode,
  RetentionUnit,
} from "@/features/retention-policies/types/retentionPolicy.types";
import { message } from "antd";

export const useRetentionPolicies = (
  productId: number | string | undefined,
  environmentId: number | undefined,
) => {
  const queryClient = useQueryClient();

  const policiesQuery = useQuery({
    queryKey: ["retention-policies", productId, environmentId],
    queryFn: () => retentionPolicyService.getPolicies(productId!, environmentId!),
    enabled: Boolean(productId && environmentId),
  });

  const statsQuery = useQuery({
    queryKey: ["retention-policy-stats", productId, environmentId],
    queryFn: () => retentionPolicyService.getStats(productId!, environmentId!),
    enabled: Boolean(productId && environmentId),
  });

  const createMutation = useMutation({
    mutationFn: (payload: CreateRetentionPolicyPayload) =>
      retentionPolicyService.createPolicy(productId!, payload),
    onSuccess: () => {
      message.success("สร้าง Retention Policy เรียบร้อยแล้ว");
      queryClient.invalidateQueries({ queryKey: ["retention-policies", productId, environmentId] });
      queryClient.invalidateQueries({ queryKey: ["retention-policy-stats", productId, environmentId] });
    },
    onError: (err: Error) => {
      message.error(`สร้าง Retention Policy ไม่สำเร็จ: ${err.message}`);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ policyId, payload }: { policyId: number | string; payload: UpdateRetentionPolicyPayload }) =>
      retentionPolicyService.updatePolicy(productId!, policyId, payload),
    onSuccess: () => {
      message.success("อัปเดต Retention Policy เรียบร้อยแล้ว");
      queryClient.invalidateQueries({ queryKey: ["retention-policies", productId, environmentId] });
      queryClient.invalidateQueries({ queryKey: ["retention-policy-stats", productId, environmentId] });
    },
    onError: (err: Error) => {
      message.error(`อัปเดต Retention Policy ไม่สำเร็จ: ${err.message}`);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (policyId: number | string) => retentionPolicyService.deletePolicy(productId!, policyId),
    onSuccess: () => {
      message.success("ลบ Retention Policy เรียบร้อยแล้ว");
      queryClient.invalidateQueries({ queryKey: ["retention-policies", productId, environmentId] });
      queryClient.invalidateQueries({ queryKey: ["retention-policy-stats", productId, environmentId] });
    },
    onError: (err: Error) => {
      message.error(`ลบ Retention Policy ไม่สำเร็จ: ${err.message}`);
    },
  });

  const toggleMutation = useMutation({
    mutationFn: ({ policyId, isActive }: { policyId: number | string; isActive: boolean }) =>
      retentionPolicyService.togglePolicy(productId!, policyId, isActive),
    onSuccess: (_, variables) => {
      message.success(`เปลี่ยนสถานะเป็น ${variables.isActive ? "เปิดใช้งาน" : "ปิดใช้งาน"} แล้ว`);
      queryClient.invalidateQueries({ queryKey: ["retention-policies", productId, environmentId] });
      queryClient.invalidateQueries({ queryKey: ["retention-policy-stats", productId, environmentId] });
    },
    onError: (err: Error) => {
      message.error(`เปลี่ยนสถานะไม่สำเร็จ: ${err.message}`);
    },
  });

  const triggerNowMutation = useMutation({
    mutationFn: (policyId: number | string) =>
      retentionPolicyService.triggerPolicyNow(productId!, policyId),
    onSuccess: () => {
      message.success("ส่งคำสั่งทดสอบไปยัง Retention Worker แล้ว");
      queryClient.invalidateQueries({ queryKey: ["retention-policies", productId, environmentId] });
    },
    onError: (err: Error) => {
      message.error(`ส่งคำสั่งทดสอบไม่สำเร็จ: ${err.message}`);
    },
  });

  const simulateMutation = useMutation({
    mutationFn: (payload: {
      retention_mode: RetentionMode;
      retention_unit: RetentionUnit;
      retention_value: number;
      folder_structure?: string;
    }) => retentionPolicyService.simulatePolicy(productId!, payload),
  });

  return {
    policies: policiesQuery.data || [],
    isLoading: policiesQuery.isLoading,
    isError: policiesQuery.isError,
    error: policiesQuery.error,
    refetch: policiesQuery.refetch,

    stats: statsQuery.data,
    isStatsLoading: statsQuery.isLoading,

    createPolicy: createMutation.mutateAsync,
    isCreating: createMutation.isPending,

    updatePolicy: updateMutation.mutateAsync,
    isUpdating: updateMutation.isPending,

    deletePolicy: deleteMutation.mutateAsync,
    isDeleting: deleteMutation.isPending,

    togglePolicy: toggleMutation.mutateAsync,
    isToggling: toggleMutation.isPending,

    triggerPolicyNow: triggerNowMutation.mutateAsync,
    isTriggering: triggerNowMutation.isPending,

    simulatePolicy: simulateMutation.mutateAsync,
    isSimulating: simulateMutation.isPending,
    simulationData: simulateMutation.data,
  };
};
