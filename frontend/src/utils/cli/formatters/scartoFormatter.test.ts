import { describe, expect, it } from 'vitest';
import { makeScartoState } from '../../../test/stateFactories';
import { formatScartoState } from './scartoFormatter';

describe('formatScartoState', () => {
  it('renders the header, deal/trick/phase line, and scores', () => {
    const out = formatScartoState(makeScartoState());
    expect(out).toContain('Scarto');
    expect(out).toContain('deal: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('phase: Play');
    expect(out).toContain('scores:');
  });

  it('shows the human hand with indexed cards and the dealer role', () => {
    const out = formatScartoState(makeScartoState());
    // Player 2 is the dealer in the base state.
    expect(out).toContain('(Dealer)');
    expect(out).toContain('(Player)');
    // Human hand is rendered with indices.
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards are on the table', () => {
    const out = formatScartoState(
      makeScartoState({
        currentTrick: [
          { playerIdx: 1, card: { design: 'HEART', value: 5, glyph: '♥', label: '5', color: 'red', deck: 'tarot' } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders the deal settlement at round end', () => {
    const out = formatScartoState(makeScartoState({ phase: 3, outcome: 1, dealScores: [4, -1, -3] }));
    expect(out).toContain('deal result: Above average');
    expect(out).toContain('deal settlement:');
    expect(out).toContain('+4');
  });

  it('renders a hint line and the win banner', () => {
    const out = formatScartoState(
      makeScartoState({
        messageCode: 'scarto.hintRequested',
        hint: { cardIndices: [1, 2], reason: 'lead_low' },
        gameEndFlag: true,
        winnerPlayer: 0,
      }),
    );
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('Winner: Player 0');
  });

  it('renders a draw banner when there is no winner', () => {
    const out = formatScartoState(makeScartoState({ gameEndFlag: true, winnerPlayer: -1 }));
    expect(out).toContain('Draw!');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1, 2], reason: 'lead_low' };
    expect(formatScartoState(makeScartoState({ hint, messageCode: 'scarto.hintRequested' }))).toContain('HINT');
    expect(formatScartoState(makeScartoState({ hint, messageCode: 'scarto.playing' }))).not.toContain('HINT');
  });
});
