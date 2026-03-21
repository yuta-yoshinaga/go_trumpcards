import { describe, expect, it } from 'bun:test';
import { buildHumanActionState, buildReplayStates } from './replayBuilder';

interface TestResponse {
  counts: number[];
  extra: string;
}

interface TestAction {
  idx: number;
  delta: number;
}

interface TestCtx {
  counts: number[];
}

const initContext = (fs: TestResponse): TestCtx => ({ counts: [...fs.counts] });
const reverseAction = (ctx: TestCtx, a: TestAction) => {
  ctx.counts[a.idx] += a.delta;
};
const applyAction = (ctx: TestCtx, a: TestAction) => {
  ctx.counts[a.idx] -= a.delta;
};
const buildState = (
  fs: TestResponse,
  ctx: TestCtx,
  _a: TestAction,
  processed: TestAction[],
  isLast: boolean,
): TestResponse => ({
  counts: [...ctx.counts],
  extra: isLast ? fs.extra : `step-${processed.length}`,
});

describe('buildReplayStates', () => {
  it('returns empty array when actions is empty', () => {
    const result = buildReplayStates({
      actions: [],
      finalState: { counts: [3, 5], extra: 'done' },
      initContext,
      reverseAction,
      applyAction,
      buildState,
    });
    expect(result).toEqual([]);
  });

  it('builds intermediate states for a single action', () => {
    const finalState: TestResponse = { counts: [2, 5], extra: 'done' };
    const actions: TestAction[] = [{ idx: 0, delta: 1 }];

    const result = buildReplayStates({
      actions,
      finalState,
      initContext,
      reverseAction,
      applyAction,
      buildState,
    });

    expect(result).toHaveLength(1);
    // reverse: counts[0] = 2 + 1 = 3; apply: 3 - 1 = 2; isLast=true
    expect(result[0]).toEqual({ counts: [2, 5], extra: 'done' });
  });

  it('builds intermediate states for multiple actions', () => {
    // Final: [1, 3]. Two actions: idx=0 delta=2, idx=1 delta=1
    const finalState: TestResponse = { counts: [1, 3], extra: 'final' };
    const actions: TestAction[] = [
      { idx: 0, delta: 2 },
      { idx: 1, delta: 1 },
    ];

    const result = buildReplayStates({
      actions,
      finalState,
      initContext,
      reverseAction,
      applyAction,
      buildState,
    });

    expect(result).toHaveLength(2);
    // reverse pass: start [1,3], reverse action[1]: counts[1]+=1 → [1,4], reverse action[0]: counts[0]+=2 → [3,4]
    // forward: apply action[0]: counts[0]-=2 → [1,4], isLast=false → extra="step-1"
    expect(result[0]).toEqual({ counts: [1, 4], extra: 'step-1' });
    // apply action[1]: counts[1]-=1 → [1,3], isLast=true → extra="final"
    expect(result[1]).toEqual({ counts: [1, 3], extra: 'final' });
  });

  it('passes processedActions slice to buildState', () => {
    const finalState: TestResponse = { counts: [0], extra: '' };
    const actions: TestAction[] = [
      { idx: 0, delta: 1 },
      { idx: 0, delta: 1 },
      { idx: 0, delta: 1 },
    ];

    const processedLengths: number[] = [];
    buildReplayStates({
      actions,
      finalState,
      initContext,
      reverseAction,
      applyAction,
      buildState: (fs, ctx, _a, processed, isLast) => {
        processedLengths.push(processed.length);
        return buildState(fs, ctx, _a, processed, isLast);
      },
    });

    expect(processedLengths).toEqual([1, 2, 3]);
  });
});

describe('buildHumanActionState', () => {
  it('returns null when actions is empty', () => {
    const result = buildHumanActionState({
      actions: [],
      finalState: { counts: [3, 5], extra: 'done' },
      initContext,
      reverseAction,
      buildState: (fs, ctx) => ({ counts: [...ctx.counts], extra: fs.extra }),
    });
    expect(result).toBeNull();
  });

  it('builds the pre-CPU state by reversing all actions', () => {
    // Final counts: [2, 4]. Actions reverse: idx=0 delta=1, idx=1 delta=2
    const finalState: TestResponse = { counts: [2, 4], extra: 'human' };
    const actions: TestAction[] = [
      { idx: 0, delta: 1 },
      { idx: 1, delta: 2 },
    ];

    const result = buildHumanActionState({
      actions,
      finalState,
      initContext,
      reverseAction,
      buildState: (fs, ctx) => ({ counts: [...ctx.counts], extra: fs.extra }),
    });

    // reverse action[1]: counts[1] += 2 → [2,6]; reverse action[0]: counts[0] += 1 → [3,6]
    expect(result).toEqual({ counts: [3, 6], extra: 'human' });
  });

  it('builds state for single action', () => {
    const finalState: TestResponse = { counts: [5], extra: 'x' };
    const actions: TestAction[] = [{ idx: 0, delta: 3 }];

    const result = buildHumanActionState({
      actions,
      finalState,
      initContext,
      reverseAction,
      buildState: (fs, ctx) => ({ counts: [...ctx.counts], extra: fs.extra }),
    });

    // reverse: counts[0] += 3 → [8]
    expect(result).toEqual({ counts: [8], extra: 'x' });
  });
});
