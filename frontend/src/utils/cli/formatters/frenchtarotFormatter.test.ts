import { describe, expect, it } from 'vitest';
import { makeFrenchTarotState } from '../../../test/stateFactories';
import type { FrenchTarotPhaseValue } from '../../../types/card';
import { formatFrenchTarotState } from './frenchtarotFormatter';

describe('formatFrenchTarotState', () => {
  it('formats a play-phase state with header, players and hand', () => {
    const out = formatFrenchTarotState(makeFrenchTarotState());
    expect(out).toContain('French Tarot');
    expect(out).toContain('phase: Play');
    expect(out).toContain('contract: Petite');
    expect(out).toContain('Declarer');
    // The human's indexed hand is shown.
    expect(out).toContain('[0]');
  });

  it('renders the current trick when cards are present', () => {
    const out = formatFrenchTarotState(
      makeFrenchTarotState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 13, glyph: '♥', label: 'D', color: 'red', deck: 'tarot' } },
          { playerIdx: 1, card: { design: 'CLOVER', value: 1, glyph: '♣', label: '1', color: 'black', deck: 'tarot' } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders the revealed chien', () => {
    const out = formatFrenchTarotState(
      makeFrenchTarotState({
        phase: 1,
        chienRevealed: true,
        chien: [{ design: 'SPADE', value: 5, glyph: '♠', label: '5', color: 'black', deck: 'tarot' }],
      }),
    );
    expect(out).toContain('chien:');
  });

  it('renders the deal result and game-over line', () => {
    const out = formatFrenchTarotState(
      makeFrenchTarotState({ phase: 5, outcome: 1, gameEndFlag: true, winnerPlayer: 0 }),
    );
    expect(out).toContain('deal result: Made (declarer wins)');
    expect(out).toContain('Game Over! Winner: Player 0');
  });

  it('renders a hint line and message', () => {
    const out = formatFrenchTarotState(
      makeFrenchTarotState({
        messageCode: 'frenchtarot.hintRequested',
        hint: { cardIndices: [1, 2], reason: 'discard_weak' },
        message: 'hello',
      }),
    );
    expect(out).toContain('HINT: card indices [1, 2] (discard_weak)');
    expect(out).toContain('hello');
  });

  it('falls back to raw numbers for out-of-range phase / contract / bid', () => {
    const out = formatFrenchTarotState(
      makeFrenchTarotState({ phase: 99 as FrenchTarotPhaseValue, contract: 99, highestBid: 99 }),
    );
    // Unknown enum indices fall through the `?? state.x` fallbacks and print the raw number.
    expect(out).toContain('phase: 99');
    expect(out).toContain('contract: 99');
    expect(out).toContain('highestBid: 99');
  });

  it('falls back to the raw outcome number when out of range', () => {
    const out = formatFrenchTarotState(makeFrenchTarotState({ phase: 5, outcome: 99 }));
    expect(out).toContain('deal result: 99');
  });

  it('renders a trick card whose playerIdx is not a seated player', () => {
    const out = formatFrenchTarotState(
      makeFrenchTarotState({
        currentTrick: [
          { playerIdx: 9, card: { design: 'SPADE', value: 5, glyph: '♠', label: '5', color: 'black', deck: 'tarot' } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1, 2], reason: 'discard_weak' };
    expect(formatFrenchTarotState(makeFrenchTarotState({ hint, messageCode: 'frenchtarot.hintRequested' }))).toContain(
      'HINT',
    );
    expect(formatFrenchTarotState(makeFrenchTarotState({ hint, messageCode: 'frenchtarot.playing' }))).not.toContain(
      'HINT',
    );
  });
});
