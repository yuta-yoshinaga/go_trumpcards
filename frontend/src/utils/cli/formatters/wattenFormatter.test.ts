import { describe, expect, it } from 'vitest';
import { makeWattenState } from '../../../test/stateFactories';
import { formatWattenState } from './wattenFormatter';

describe('formatWattenState', () => {
  it('renders the header, deal/trick/phase, schlag, critical suit, stake and team scores', () => {
    const out = formatWattenState(makeWattenState({ schlagRank: 13, criticalSuit: 3, teamScores: [4, 2] }));
    expect(out).toContain('Watten');
    expect(out).toContain('deal: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('phase: Play');
    expect(out).toContain('schlag: K');
    expect(out).toContain('critical: heart');
    expect(out).toContain('stake: 2');
    expect(out).toContain('T0=4');
    expect(out).toContain('T1=2');
  });

  it('renders the schlag rank as A when it is an ace', () => {
    const out = formatWattenState(makeWattenState({ schlagRank: 1 }));
    expect(out).toContain('schlag: A');
  });

  it('renders critical as dash when unset', () => {
    const out = formatWattenState(makeWattenState({ criticalSuit: -1 }));
    expect(out).toContain('critical: -');
  });

  it('renders the pending stake during a raise', () => {
    const out = formatWattenState(makeWattenState({ stake: 2, pendingStake: 3 }));
    expect(out).toContain('stake: 2 (pending 3)');
  });

  it('renders per-team trick counts', () => {
    const out = formatWattenState(makeWattenState({ teamTricks: [2, 1] }));
    expect(out).toContain('tricks: T0=2  T1=1');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatWattenState(makeWattenState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatWattenState(
      makeWattenState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 13 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders a hint with the action, card index and reason', () => {
    const out = formatWattenState(
      makeWattenState({
        hint: { action: 'play', cardIndex: 2, reason: 'lead_trump' },
        messageCode: 'watten.hintRequested',
      }),
    );
    expect(out).toContain('HINT: play [2]');
    expect(out).toContain('lead_trump');
  });

  it('renders a declare hint without a card index', () => {
    const out = formatWattenState(
      makeWattenState({ hint: { action: 'declare', reason: 'declare_strong' }, messageCode: 'watten.hintRequested' }),
    );
    expect(out).toContain('HINT: declare');
    expect(out).toContain('declare_strong');
  });

  it('renders the game-over banner with the winning team', () => {
    const out = formatWattenState(makeWattenState({ phase: 5, gameEndFlag: true, winnerTeam: 0 }));
    expect(out).toContain('Game Over! Winner: Team 0');
  });

  it('renders an explicit message when present', () => {
    const out = formatWattenState(makeWattenState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { action: 'play', cardIndex: 2, reason: 'lead_trump' };
    expect(formatWattenState(makeWattenState({ hint, messageCode: 'watten.hintRequested' }))).toContain('HINT');
    expect(formatWattenState(makeWattenState({ hint, messageCode: 'watten.playing' }))).not.toContain('HINT');
  });
});
