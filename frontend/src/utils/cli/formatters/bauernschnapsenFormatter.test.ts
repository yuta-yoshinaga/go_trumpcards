import { describe, expect, it } from 'vitest';
import type { BauernschnapsenResponse } from '../../../types/card';
import { formatBauernschnapsenState } from './bauernschnapsenFormatter';

const state = (overrides: Partial<BauernschnapsenResponse> = {}): BauernschnapsenResponse => ({
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 2,
      cards: [
        { design: 'SPADE', value: 13 },
        { design: 'HEART', value: 12 },
      ],
      team: 0,
      trickCount: 1,
    },
    { id: 1, isHuman: false, cardCount: 2, cards: [], team: 1, trickCount: 0 },
  ],
  phase: 1,
  roundNumber: 2,
  trickNumber: 1,
  currentPlayerIdx: 0,
  dealerIdx: 1,
  trumpSuit: 1,
  contract: 1,
  declarerIdx: 0,
  currentTrick: [],
  teamScores: [4, 2],
  roundPoints: [0, 0],
  roundMarriage: [0, 0],
  validPlayIndices: [0, 1],
  marriageIndices: [],
  gameEndFlag: false,
  winnerTeam: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1, targetScore: 24 },
  ...overrides,
});

describe('formatBauernschnapsenState', () => {
  it('formats contract play with indexed selectable hand and scores', () => {
    const text = formatBauernschnapsenState(state());
    expect(text).toContain('Bauernschnapsen');
    expect(text).toContain('Rufer');
    expect(text).toContain('[0]♠K');
    expect(text).toContain('[1]♥Q');
    expect(text).toContain('use p <idx>');
    expect(text).toContain('Team0=4');
  });

  it('formats trick and a requested card or non-card hint only', () => {
    const trick = [{ playerIdx: 1, card: { design: 'CLOVER' as const, value: 11 } }];
    expect(formatBauernschnapsenState(state({ currentTrick: trick }))).toContain('trick: CPU 1=♣J');
    const cardHint = { cardIndex: 1, reason: 'follow_win', isMarriage: false };
    expect(formatBauernschnapsenState(state({ hint: cardHint }))).not.toContain('HINT:');
    expect(
      formatBauernschnapsenState(state({ hint: cardHint, messageCode: 'bauernschnapsen.hintRequested' })),
    ).toContain('HINT: play card [1]');
    expect(
      formatBauernschnapsenState(
        state({ hint: { reason: 'declare_bettel', isMarriage: false }, messageCode: 'bauernschnapsen.hintRequested' }),
      ),
    ).toContain('HINT: declare_bettel');
    expect(formatBauernschnapsenState(state({ hint: cardHint, messageCode: 'bauernschnapsen.noHint' }))).not.toContain(
      'HINT:',
    );
  });

  it('formats marriage hints and game end', () => {
    const text = formatBauernschnapsenState(
      state({
        hint: { cardIndex: 0, reason: 'marriage', isMarriage: true },
        messageCode: 'bauernschnapsen.hintRequested',
        gameEndFlag: true,
        winnerTeam: 0,
      }),
    );
    expect(text).toContain('HINT: declare a marriage with card [0]');
    expect(text).toContain('Team 0 wins');
    expect(
      formatBauernschnapsenState(
        state({ hint: { reason: 'marriage', isMarriage: true }, messageCode: 'bauernschnapsen.hintRequested' }),
      ),
    ).toContain('card [?]');
    expect(formatBauernschnapsenState(state({ gameEndFlag: true, winnerTeam: -1 }))).toContain('Game Over! Draw.');
  });

  it('handles undecided trump, unknown values and empty optional sections', () => {
    const text = formatBauernschnapsenState(state({ contract: 9, trumpSuit: 0, message: '', currentTrick: [] }));
    expect(text).toContain('contract: Unknown');
    expect(text).not.toContain('trump:');
    expect(formatBauernschnapsenState(state({ trumpSuit: 9 }))).toContain('trump: ?');
  });
});
