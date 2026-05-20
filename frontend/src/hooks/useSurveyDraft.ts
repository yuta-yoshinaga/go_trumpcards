/**
 * Persistent survey draft for `/discover` resume-after-reload.
 *
 * Stores the in-progress answers under a single localStorage key. The
 * stored blob is version-tagged; reading a draft from an incompatible
 * schema version drops it and starts fresh (no silent migration).
 *
 * `setItem` is wrapped in try/catch so Private Browsing / quota errors
 * never crash the survey — the draft simply isn't persisted in that
 * session, which is the agreed graceful behavior.
 */

import { useCallback, useEffect, useState } from 'react';
import { AXIS_KEYS, AXIS_QUESTION_OPTION_COUNTS, type AxisKey } from '../constants/discoverAxes';

/** localStorage key for the discover draft (versioned blob). */
export const DISCOVER_DRAFT_KEY = 'trumpcards-discover-draft';

/**
 * Bumping this invalidates every previously stored draft. v2 changed
 * answer indices from "master option list" to "per-sub-question options",
 * so old drafts must not be silently re-interpreted.
 */
export const DRAFT_SCHEMA_VERSION = 2;

/** One axis's answers (length 2). */
export type DraftAxisAnswers = readonly [number | null, number | null];

/** Full draft payload (versioned). */
export interface SurveyDraft {
  readonly v: number;
  readonly axes: Readonly<Record<AxisKey, DraftAxisAnswers>>;
}

const EMPTY_ANSWERS: Readonly<Record<AxisKey, DraftAxisAnswers>> = {
  mood: [null, null],
  skill: [null, null],
  social: [null, null],
  theme: [null, null],
};

/** Read the persisted draft, or `null` if missing/invalid/version-mismatched. */
export function readDraft(): SurveyDraft | null {
  try {
    const raw = localStorage.getItem(DISCOVER_DRAFT_KEY);
    if (!raw) return null;
    const parsed: unknown = JSON.parse(raw);
    if (!parsed || typeof parsed !== 'object') return null;
    const obj = parsed as Partial<SurveyDraft>;
    if (obj.v !== DRAFT_SCHEMA_VERSION) {
      // Schema mismatch — drop and start fresh.
      localStorage.removeItem(DISCOVER_DRAFT_KEY);
      return null;
    }
    if (!obj.axes || typeof obj.axes !== 'object') return null;
    const axes: Record<AxisKey, DraftAxisAnswers> = { ...EMPTY_ANSWERS };
    for (const key of AXIS_KEYS) {
      const v = (obj.axes as Record<string, unknown>)[key];
      if (Array.isArray(v) && v.length === 2) {
        const limits = AXIS_QUESTION_OPTION_COUNTS[key];
        const sanitized: (number | null)[] = v.map((entry, qIdx) => {
          if (entry === null) return null;
          if (typeof entry !== 'number' || !Number.isInteger(entry)) return null;
          const max = limits[qIdx as 0 | 1];
          if (entry < 0 || entry >= max) return null;
          return entry;
        });
        axes[key] = [sanitized[0] ?? null, sanitized[1] ?? null] as const;
      }
    }
    return { v: DRAFT_SCHEMA_VERSION, axes };
  } catch {
    return null;
  }
}

/** Persist a draft, swallowing storage errors (private mode / quota). */
function writeDraft(draft: SurveyDraft): void {
  try {
    localStorage.setItem(DISCOVER_DRAFT_KEY, JSON.stringify(draft));
  } catch {
    if (import.meta.env.DEV) {
      console.warn('[useSurveyDraft] failed to persist draft (storage unavailable)');
    }
  }
}

/** Clear the persisted draft (no-op if missing). */
export function clearDraft(): void {
  try {
    localStorage.removeItem(DISCOVER_DRAFT_KEY);
  } catch {
    /* ignore */
  }
}

/**
 * Hook that returns the current axes answers plus a setter and a clear
 * function. Reads existing draft on mount; persists every update.
 */
export function useSurveyDraft() {
  const [axes, setAxes] = useState<Readonly<Record<AxisKey, DraftAxisAnswers>>>(
    () => readDraft()?.axes ?? EMPTY_ANSWERS,
  );

  useEffect(() => {
    // Don't persist a fully-empty draft — it would re-create the entry
    // immediately after a reset() and defeat the clear.
    const hasAny = AXIS_KEYS.some((k) => axes[k][0] !== null || axes[k][1] !== null);
    if (hasAny) writeDraft({ v: DRAFT_SCHEMA_VERSION, axes });
    else clearDraft();
  }, [axes]);

  const setAnswer = useCallback((axis: AxisKey, qIdx: 0 | 1, value: number | null) => {
    setAxes((prev) => {
      const current = prev[axis];
      const next: DraftAxisAnswers = qIdx === 0 ? [value, current[1]] : [current[0], value];
      return { ...prev, [axis]: next };
    });
  }, []);

  const reset = useCallback(() => {
    setAxes(EMPTY_ANSWERS);
    clearDraft();
  }, []);

  return { axes, setAnswer, reset };
}
