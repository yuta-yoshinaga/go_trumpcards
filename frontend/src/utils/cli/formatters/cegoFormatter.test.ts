import { describe, expect, it } from 'vitest';
import { makeCegoState } from '../../../test/stateFactories';
import type { CegoPhaseValue } from '../../../types/card';
import { formatCegoState } from './cegoFormatter';

describe('formatCegoState', () => {
  it('renders the header, deal/phase line and the human hand', () => {
    const out = formatCegoState(makeCegoState());
    expect(out).toContain('Cego');
    expect(out).toContain('phase: Play');
    expect(out).toContain('contract: Cego');
    expect(out).toContain('blind: 10');
    expect(out).toContain('scores: P0=0');
    // Human hand is rendered with indexed cards; a CPU (Defender) row is present.
    expect(out).toContain('Declarer');
    expect(out).toContain('Defender');
  });

  it('renders the current trick line when a trick is in progress', () => {
    const out = formatCegoState(
      makeCegoState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 8, glyph: '♥', label: 'K', color: 'red', deck: 'tarot' } },
          { playerIdx: 1, card: { design: 'SPADE', value: 5, glyph: '♠', label: 'J', color: 'black', deck: 'tarot' } },
        ],
      }),
    );
    expect(out).toMatch(/^trick:/m);
  });

  it('omits the trick line when no cards have been played', () => {
    expect(formatCegoState(makeCegoState({ currentTrick: [] }))).not.toMatch(/^trick:/m);
  });

  it('shows the deal result at round end when the contract was made', () => {
    const out = formatCegoState(makeCegoState({ phase: 5, outcome: 1 }));
    expect(out).toContain('deal result: Made (declarer wins)');
  });

  it('shows the deal result at round end when the contract failed', () => {
    const out = formatCegoState(makeCegoState({ phase: 5, outcome: 2 }));
    expect(out).toContain('deal result: Failed (defenders win)');
  });

  it('does not show a deal result outside round end', () => {
    expect(formatCegoState(makeCegoState({ phase: 3, outcome: 1 }))).not.toContain('deal result:');
  });

  it('renders a hint line with its card indices and reason', () => {
    const out = formatCegoState(
      makeCegoState({ messageCode: 'cego.hintRequested', hint: { cardIndices: [1, 3], reason: 'lead_high' } }),
    );
    expect(out).toContain('HINT: card indices [1, 3] (lead_high)');
  });

  it('renders a hint line even when cardIndices is missing', () => {
    const out = formatCegoState(
      makeCegoState({
        hint: { reason: 'bid_take', cardIndices: undefined as unknown as number[] },
        messageCode: 'cego.hintRequested',
      }),
    );
    expect(out).toContain('HINT: card indices [] (bid_take)');
  });

  it('appends the server message when present', () => {
    expect(formatCegoState(makeCegoState({ message: 'hello world' }))).toContain('hello world');
  });

  it('announces the winner at game end', () => {
    const out = formatCegoState(makeCegoState({ gameEndFlag: true, winnerPlayer: 2 }));
    expect(out).toContain('Game Over! Winner: Player 2');
  });

  it('announces a draw at game end when there is no winner', () => {
    const out = formatCegoState(makeCegoState({ gameEndFlag: true, winnerPlayer: -1 }));
    expect(out).toContain('Game Over! Draw!');
  });

  it('falls back to raw values for out-of-range phase and contract type', () => {
    const out = formatCegoState(makeCegoState({ phase: 99 as CegoPhaseValue, contractType: 9 }));
    expect(out).toContain('phase: 99');
    expect(out).toContain('contract: 9');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1, 3], reason: 'lead_high' };
    expect(formatCegoState(makeCegoState({ hint, messageCode: 'cego.hintRequested' }))).toContain('HINT');
    expect(formatCegoState(makeCegoState({ hint, messageCode: 'cego.playing' }))).not.toContain('HINT');
  });
});
