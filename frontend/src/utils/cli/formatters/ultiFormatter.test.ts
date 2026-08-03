import { describe, expect, it } from 'vitest';
import { makeUltiState } from '../../../test/stateFactories';
import { formatUltiState } from './ultiFormatter';

describe('formatUltiState', () => {
  it('renders the header, deal/trick, contract, trump and per-player coins', () => {
    const out = formatUltiState(makeUltiState({ playerCoins: [3, 1, 2], contract: 1, trumpSuit: 1 }));
    expect(out).toContain('Ulti');
    expect(out).toContain('deal: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('contract: Party');
    expect(out).toContain('trump: spade');
    expect(out).toContain('P0=3');
    expect(out).toContain('P1=1');
    expect(out).toContain('P2=2');
  });

  it('marks the Declarer role for the human (seat 0)', () => {
    const out = formatUltiState(makeUltiState());
    expect(out).toContain('Declarer');
    expect(out).toContain('Coalition');
  });

  it('renders a betli contract label', () => {
    const out = formatUltiState(makeUltiState({ contract: 2 }));
    expect(out).toContain('contract: Betli');
  });

  it('renders trump as dash when unset', () => {
    const out = formatUltiState(makeUltiState({ trumpSuit: -1 }));
    expect(out).toContain('trump: -');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatUltiState(makeUltiState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatUltiState(
      makeUltiState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders the outcome during RoundEnd', () => {
    const out = formatUltiState(makeUltiState({ phase: 4, outcome: 2 }));
    expect(out).toContain('deal result: Failed (coalition wins)');
  });

  it('renders a hint with card indices', () => {
    const out = formatUltiState(
      makeUltiState({ hint: { cardIndices: [1, 2], reason: 'follow_win' }, messageCode: 'ulti.hintRequested' }),
    );
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('follow_win');
  });

  it('renders the game-over banner with the winning player', () => {
    const out = formatUltiState(makeUltiState({ phase: 5, gameEndFlag: true, winnerPlayer: 1 }));
    expect(out).toContain('Game Over! Winner: Player 1');
  });

  it('renders an explicit message when present', () => {
    const out = formatUltiState(makeUltiState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1, 2], reason: 'follow_win' };
    expect(formatUltiState(makeUltiState({ hint, messageCode: 'ulti.hintRequested' }))).toContain('HINT');
    expect(formatUltiState(makeUltiState({ hint, messageCode: 'ulti.playing' }))).not.toContain('HINT');
  });
});
