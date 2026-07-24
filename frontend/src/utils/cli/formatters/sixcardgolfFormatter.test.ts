import { describe, expect, it } from 'vitest';
import type { SixCardGolfResponse } from '../../../types/card';
import { formatSixCardGolfState } from './sixcardgolfFormatter';

describe('formatSixCardGolfState', () => {
  it('renders the round and phase summary', () => {
    const out = formatSixCardGolfState({ roundNumber: 2, totalRounds: 9, phase: 1 } as SixCardGolfResponse);
    expect(out).toBe('Round 2/9 | Phase 1');
  });
});
