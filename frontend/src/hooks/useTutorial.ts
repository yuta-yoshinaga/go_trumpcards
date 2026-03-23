import { useCallback, useMemo, useState } from 'react';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';

/** Storage key prefix for tutorial completion flags. */
const STORAGE_PREFIX = 'tutorial_completed_';

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
  /** Start or restart the tutorial. */
  start: () => void;
  /** Advance to the next step, completing the tutorial if on the last step. */
  next: () => void;
  /** Skip/dismiss the tutorial without marking it complete. */
  skip: () => void;
}

/** Manages tutorial state: step progression, completion persistence, and lifecycle callbacks. */
export function useTutorial(config: TutorialConfig): UseTutorialReturn {
  const storageKey = `${STORAGE_PREFIX}${config.gameName}`;

  const [isActive, setIsActive] = useState(false);
  const [currentStepIndex, setCurrentStepIndex] = useState(0);
  const [isCompleted, setIsCompleted] = useState(() => localStorage.getItem(storageKey) === 'true');

  const totalSteps = config.steps.length;

  const currentStep = isActive ? (config.steps[currentStepIndex] ?? null) : null;

  /** Starts or restarts the tutorial. Does not clear localStorage completion flag — once completed, isCompleted remains true until the tutorial is finished again. */
  const start = useCallback(() => {
    if (totalSteps === 0) {
      setIsCompleted(true);
      localStorage.setItem(storageKey, 'true');
      return;
    }
    setCurrentStepIndex(0);
    setIsActive(true);
    config.steps[0]?.onEnter?.();
  }, [config.steps, storageKey, totalSteps]);

  const next = useCallback(() => {
    if (!isActive) return;
    const nextIndex = currentStepIndex + 1;
    if (nextIndex >= totalSteps) {
      setIsActive(false);
      setIsCompleted(true);
      localStorage.setItem(storageKey, 'true');
    } else {
      setCurrentStepIndex(nextIndex);
      config.steps[nextIndex]?.onEnter?.();
    }
  }, [isActive, currentStepIndex, totalSteps, storageKey, config.steps]);

  const skip = useCallback(() => {
    if (!isActive) return;
    setIsActive(false);
    setCurrentStepIndex(0);
  }, [isActive]);

  return useMemo(
    () => ({ isActive, currentStepIndex, currentStep, totalSteps, isCompleted, start, next, skip }),
    [isActive, currentStepIndex, currentStep, totalSteps, isCompleted, start, next, skip],
  );
}
