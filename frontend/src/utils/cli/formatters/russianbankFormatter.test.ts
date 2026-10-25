import { describe, expect, it } from 'vitest';
import { makeRussianBankState } from '../../../test/stateFactories';
import { RussianBankPhase } from '../../../types/phases';
import { formatRussianBankState } from './russianbankFormatter';

describe('formatRussianBankState', () => {
  it('shows phase, top cards, and available controls during play', () => {
    const s = makeRussianBankState({
      tableau: [
        [
          { design: 'SPADE', value: 8 },
          { design: 'HEART', value: 9 },
        ],
        [],
        [],
        [],
      ],
      canUndo: true,
      canCallStop: true,
    });
    const text = formatRussianBankState(s);
    expect(text).toContain('phase: Playing');
    expect(text).toContain('Turn: You');
    expect(text).toContain('reserve *available');
    expect(text).toContain('Opponent sources: reserve *available');
    expect(text).toContain('Legal moves (*):');
    expect(text).toContain('[0] ♥9 (2 cards)');
    expect(text).toContain('Use s to call stop.');
    expect(text).toContain('Use u to undo.');
  });
  it('shows requested move hints using server message codes only', () => {
    const hint = { zone: 1, fromOpponent: true, col: 0, toFoundation: false, toCol: 2 };
    const base = makeRussianBankState({ hint });
    expect(formatRussianBankState(base)).not.toContain('HINT');
    expect(formatRussianBankState({ ...base, messageCode: 'russianbank.hintRequested' })).toContain(
      'HINT: move opponent waste to tableau column 2.',
    );
    expect(formatRussianBankState({ ...base, hint: undefined })).not.toContain('HINT');
    expect(formatRussianBankState({ ...base, hint: undefined, messageCode: 'russianbank.noHint' })).toContain(
      'No hint available.',
    );
  });
  it('renders a human tableau hint to a foundation and falls back for an unknown zone', () => {
    const state = makeRussianBankState({
      hint: { zone: 2, fromOpponent: false, col: 1, toFoundation: true, toCol: 0 },
      messageCode: 'russianbank.hintRequested',
      message: 'server hint copy',
    });
    expect(formatRussianBankState(state)).toContain('HINT: move tableau column 1 to its foundation.');
    expect(formatRussianBankState(state)).toContain('server hint copy');
    expect(
      formatRussianBankState({
        ...state,
        hint: { zone: 9, fromOpponent: false, col: 0, toFoundation: false, toCol: 3 },
      }),
    ).toContain('your pile');
  });
  it('handles idle states without a human player and suppresses messages after game end', () => {
    const idle = makeRussianBankState({ phase: RussianBankPhase.IDLE, players: [] });
    expect(formatRussianBankState(idle)).toContain('phase: Idle');
    expect(formatRussianBankState(idle)).toContain('Use r to reset.');
    expect(
      formatRussianBankState(
        makeRussianBankState({ phase: RussianBankPhase.GAME_END, gameEndFlag: true, winnerIdx: 1, message: '日本語' }),
      ),
    ).not.toContain('日本語');
  });
  it('shows only reset after game end and reports the result in English', () => {
    const text = formatRussianBankState(
      makeRussianBankState({ phase: RussianBankPhase.GAME_END, gameEndFlag: true, winnerIdx: 0, message: '終局です' }),
    );
    expect(text).toContain('Game Over! You win!');
    expect(text).toContain('Use r to reset.');
    expect(text).not.toContain('終局');
    expect(
      formatRussianBankState(
        makeRussianBankState({ phase: RussianBankPhase.GAME_END, gameEndFlag: true, winnerIdx: -1 }),
      ),
    ).toContain('Draw');
    expect(
      formatRussianBankState(
        makeRussianBankState({ phase: RussianBankPhase.GAME_END, gameEndFlag: true, winnerIdx: 1 }),
      ),
    ).toContain('CPU wins!');
  });
  it('hides move commands when no source has a legal destination', () => {
    const state = makeRussianBankState({
      foundationNext: Array.from({ length: 8 }, () => ({ design: 'HEART', value: 13 })),
      tableauNext: Array.from({ length: 4 }, () => ({ any: false, black: false, value: 13 })),
    });
    const text = formatRussianBankState(state);
    expect(text).toContain('Commands: d, h, r');
    expect(text).not.toContain('Commands: pf');
    expect(text).not.toContain('Legal moves (*)');
  });
  it('keeps non-move commands available if the active response has no human player record', () => {
    const text = formatRussianBankState(makeRussianBankState({ players: [] }));
    expect(text).toContain('Turn: You');
    expect(text).toContain('Commands: d, h, r');
  });
});
