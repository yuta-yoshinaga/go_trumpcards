import { useCallback, useMemo, useState } from 'react';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';

/** Storage key prefix for tutorial completion flags. */
const STORAGE_PREFIX = 'tutorial_completed_';

/** Storage key prefix for tutorial progress (step index on skip). */
const PROGRESS_PREFIX = 'tutorial_progress_';

/** Reads a valid progress index from localStorage, or returns -1 if absent/invalid. */
function readProgress(key: string, totalSteps: number): number {
  const stored = localStorage.getItem(key);
  if (stored === null) return -1;
  const idx = Number.parseInt(stored, 10);
  if (Number.isNaN(idx) || idx < 0 || idx >= totalSteps) return -1;
  return idx;
}

/** Return type of the useTutorial hook. */
export interface UseTutorialReturn {
  /** Whether the tutorial is currently running. */
  isActive: boolean;
  /** Zero-based index of the current step. */
  currentStepIndex: number;
  /** The current step definition, or null if not active. */
  currentStep: TutorialStep | null;
  /** Total number of steps in the tutorial. */
  totalSteps: number;
  /** Whether the tutorial has been completed (persisted in localStorage). */
  isCompleted: boolean;
  /** Whether a previous session can be resumed from a saved step. */
  canResume: boolean;
  /** Start the tutorial, resuming from saved progress if available. */
  start: () => void;
  /** Restart the tutorial from step 0, clearing any saved progress. */
  restart: () => void;
  /** Advance to the next step, completing the tutorial if on the last step. */
  next: () => void;
  /** Skip/dismiss the tutorial without marking it complete. Saves progress if past step 0. */
  skip: () => void;
}

/** Manages tutorial state: step progression, completion persistence, resume, and lifecycle callbacks. */
export function useTutorial(config: TutorialConfig): UseTutorialReturn {
  const storageKey = `${STORAGE_PREFIX}${config.gameName}`;
  const progressKey = `${PROGRESS_PREFIX}${config.gameName}`;

  const [isActive, setIsActive] = useState(false);
  const [currentStepIndex, setCurrentStepIndex] = useState(0);
  const [isCompleted, setIsCompleted] = useState(() => localStorage.getItem(storageKey) === 'true');
  const [canResume, setCanResume] = useState(() => readProgress(progressKey, config.steps.length) >= 0);

  const totalSteps = config.steps.length;

  const currentStep = isActive ? (config.steps[currentStepIndex] ?? null) : null;

  /** Starts the tutorial, resuming from saved progress if available. */
  const start = useCallback(() => {
    if (totalSteps === 0) {
      setIsCompleted(true);
      localStorage.setItem(storageKey, 'true');
      return;
    }
    const savedIdx = readProgress(progressKey, totalSteps);
    const startIdx = savedIdx >= 0 ? savedIdx : 0;
    setCurrentStepIndex(startIdx);
    setIsActive(true);
    config.steps[startIdx]?.onEnter?.();
  }, [config.steps, storageKey, progressKey, totalSteps]);

  /** Restarts the tutorial from step 0, clearing any saved progress. */
  const restart = useCallback(() => {
    if (totalSteps === 0) {
      setIsCompleted(true);
      localStorage.setItem(storageKey, 'true');
      return;
    }
    localStorage.removeItem(progressKey);
    setCanResume(false);
    setCurrentStepIndex(0);
    setIsActive(true);
    config.steps[0]?.onEnter?.();
  }, [config.steps, storageKey, progressKey, totalSteps]);

  const next = useCallback(() => {
    if (!isActive) return;
    const nextIndex = currentStepIndex + 1;
    if (nextIndex >= totalSteps) {
      setIsActive(false);
      setIsCompleted(true);
      localStorage.setItem(storageKey, 'true');
      localStorage.removeItem(progressKey);
      setCanResume(false);
    } else {
      setCurrentStepIndex(nextIndex);
      config.steps[nextIndex]?.onEnter?.();
    }
  }, [isActive, currentStepIndex, totalSteps, storageKey, progressKey, config.steps]);

  const skip = useCallback(() => {
    if (!isActive) return;
    setIsActive(false);
    if (currentStepIndex > 0) {
      localStorage.setItem(progressKey, String(currentStepIndex));
      setCanResume(true);
    }
    setCurrentStepIndex(0);
  }, [isActive, currentStepIndex, progressKey]);

  return useMemo(
    () => ({ isActive, currentStepIndex, currentStep, totalSteps, isCompleted, canResume, start, restart, next, skip }),
    [isActive, currentStepIndex, currentStep, totalSteps, isCompleted, canResume, start, restart, next, skip],
  );
}
