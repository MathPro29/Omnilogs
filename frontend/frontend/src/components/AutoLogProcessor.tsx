import { useEffect, useRef } from 'react';
import { productAdminService } from '@/services';
import { useAutoLogStore } from '@/store/auto-log.store';
import type { ProjectFeature } from '@/types';

const AUTO_LOG_LEVELS = [
  'DEBUG',
  'INFO',
  'INFO',
  'WARN',
  'ERROR',
] as const;

const AUTO_LOG_MESSAGES = {
  DEBUG: [
    'Hierarchy route validation in progress',
    'Child feature metrics updated from auto sender',
    'Auto sender inspected node metadata before dispatch',
  ],
  INFO: [
    'Auto log dispatched to hierarchy node',
    'Hierarchy node accepted the background event',
    'Auto sender completed the scheduled hierarchy dispatch',
  ],
  WARN: [
    'Hierarchy node reported a warning state',
    'Auto sender detected a degraded but recoverable node state',
    'Hierarchy node exceeded the soft threshold and raised a warning',
  ],
  ERROR: [
    'Hierarchy node raised an error event',
    'Auto sender captured a failed hierarchy operation',
    'Hierarchy node returned an exception-like failure state',
  ],
} satisfies Record<(typeof AUTO_LOG_LEVELS)[number], string[]>;

function pickLogLevel(index: number) {
  return AUTO_LOG_LEVELS[index % AUTO_LOG_LEVELS.length];
}

function pickLogMessage(level: keyof typeof AUTO_LOG_MESSAGES, index: number) {
  const messages = AUTO_LOG_MESSAGES[level];
  return messages[index % messages.length];
}

function isFeatureInScope(feature: ProjectFeature, rootCategoryId?: number) {
  if (!rootCategoryId) {
    return true;
  }

  const pathIds = feature.pathIds?.split(',').map((value) => value.trim()) || [];
  return feature.categoryId === rootCategoryId || pathIds.includes(String(rootCategoryId));
}

function getHierarchyTargets(features: ProjectFeature[] | undefined, rootCategoryId?: number) {
  return (features || [])
    .filter((feature) => feature.isActive)
    .filter((feature) => isFeatureInScope(feature, rootCategoryId))
    .sort((left, right) => {
      if (left.level !== right.level) {
        return left.level - right.level;
      }
      return (left.fullPath || left.categoryName).localeCompare(right.fullPath || right.categoryName);
    });
}

export function AutoLogProcessor() {
  const isRunning = useAutoLogStore((state) => state.isRunning);
  const config = useAutoLogStore((state) => state.config);
  const sendingRef = useRef(false);
  const sequenceIndexRef = useRef(0);

  useEffect(() => {
    if (!isRunning || !config) {
      sequenceIndexRef.current = 0;
      return;
    }

    const hierarchyTargets = getHierarchyTargets(config.features, config.categoryId);

    const sendLog = async () => {
      if (sendingRef.current) return;
      sendingRef.current = true;

      const currentIndex = sequenceIndexRef.current;
      const currentLevel = pickLogLevel(currentIndex);
      const currentMessage = pickLogMessage(currentLevel, currentIndex);
      const targetFeature =
        hierarchyTargets.length > 0
          ? hierarchyTargets[currentIndex % hierarchyTargets.length]
          : null;

      const dynamicCategoryId = targetFeature?.categoryId ?? config.categoryId;
      const dynamicFeatureFullPath = targetFeature?.fullPath ?? config.featureFullPath;
      const dynamicFeaturePathIds = targetFeature?.pathIds ?? config.featurePathIds;
      const targetLabel = targetFeature?.fullPath || targetFeature?.categoryName || 'product-scope';

      let finalCategoryId = dynamicCategoryId;
      let finalProjectId = config.projectId;
      let finalMessage = `${currentMessage} [level: ${currentLevel}] [target: ${targetLabel}] (seq: ${currentIndex + 1}, time: ${new Date().toLocaleTimeString()})`;
      let finalLogLevel = currentLevel;

      const mode = config.simulationMode || 'SUCCESS';
      let isFailure = false;
      if (mode === 'ALL_FAILURES') {
        isFailure = true;
      } else if (mode === 'MIXED_FAILURES') {
        isFailure = Math.random() < 0.3; // 30% chance of failure
      }

      if (isFailure) {
        finalLogLevel = 'ERROR';
        const rand = Math.floor(Math.random() * 3);
        if (rand === 0) {
          // Case 3: Orphan Category (Category with no Project)
          finalProjectId = undefined;
          finalCategoryId = 666666;
          finalMessage = `[Auto-Sender Simulated Failure] Orphan Category Error (seq: ${currentIndex + 1}, time: ${new Date().toLocaleTimeString()})`;
        } else if (rand === 1) {
          // Case 4: Project Mismatch
          finalProjectId = 555555;
          finalMessage = `[Auto-Sender Simulated Failure] Project Mismatch Error - Project 555555 doesn't exist (seq: ${currentIndex + 1}, time: ${new Date().toLocaleTimeString()})`;
        } else {
          // Case 5: Category Mismatch
          finalProjectId = config.projectId || 7777;
          finalCategoryId = 111111;
          finalMessage = `[Auto-Sender Simulated Failure] Category Mismatch Error - Category 111111 doesn't exist/belong (seq: ${currentIndex + 1}, time: ${new Date().toLocaleTimeString()})`;
        }
      }

      try {
        await productAdminService.importLogs({
          ...config,
          projectId: finalProjectId,
          categoryId: finalCategoryId,
          logLevel: finalLogLevel,
          message: finalMessage,
        });

        sequenceIndexRef.current += 1;
      } catch (error) {
        console.error('Auto log sending failed:', error);
      } finally {
        sendingRef.current = false;
      }
    };

    const interval = window.setInterval(sendLog, 3000);
    return () => window.clearInterval(interval);
  }, [isRunning, config]);

  return null;
}
