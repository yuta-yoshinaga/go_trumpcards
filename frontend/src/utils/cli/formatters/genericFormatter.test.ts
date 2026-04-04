import { describe, expect, it } from 'vitest';
import { formatGenericState } from './genericFormatter';

describe('formatGenericState', () => {
  it('renders title header', () => {
    const result = formatGenericState({ title: 'TestGame' });
    expect(result).toContain('TestGame');
  });

  it('renders phase as string', () => {
    const result = formatGenericState({ title: 'Game', phase: 'betting' });
    expect(result).toContain('phase: betting');
  });

  it('renders phase as number with phase names', () => {
    const result = formatGenericState({
      title: 'Game',
      phase: 1,
      phaseNames: { 1: 'Dealing' },
    });
    expect(result).toContain('phase: Dealing');
  });

  it('renders phase as number without phase names', () => {
    const result = formatGenericState({ title: 'Game', phase: 3 });
    expect(result).toContain('phase: 3');
  });

  it('renders pot', () => {
    const result = formatGenericState({ title: 'Game', pot: 500 });
    expect(result).toContain('pot: 500');
  });

  it('renders round and trick numbers', () => {
    const result = formatGenericState({ title: 'Game', roundNumber: 3, trickNumber: 2 });
    expect(result).toContain('round: 3');
    expect(result).toContain('trick: 2');
  });

  it('renders round without trick', () => {
    const result = formatGenericState({ title: 'Game', roundNumber: 1 });
    expect(result).toContain('round: 1');
    expect(result).not.toContain('trick:');
  });

  it('renders players with chips score mode', () => {
    const result = formatGenericState({
      title: 'Game',
      scoreMode: 'chips',
      players: [{ id: 0, isHuman: true, chips: 1000, cards: [] }],
    });
    expect(result).toContain('chips=1000');
  });

  it('renders players with score mode', () => {
    const result = formatGenericState({
      title: 'Game',
      scoreMode: 'score',
      players: [{ id: 0, isHuman: false, cumulativeScore: 42 }],
    });
    expect(result).toContain('total=42');
  });

  it('renders players with trick mode', () => {
    const result = formatGenericState({
      title: 'Game',
      scoreMode: 'trick',
      players: [{ id: 0, isHuman: false, trickCount: 5 }],
    });
    expect(result).toContain('tricks=5');
  });

  it('renders player statuses', () => {
    const result = formatGenericState({
      title: 'Game',
      players: [{ id: 0, isHuman: false, folded: true, allIn: false }],
    });
    expect(result).toContain('[Folded]');
  });

  it('renders allIn status', () => {
    const result = formatGenericState({
      title: 'Game',
      players: [{ id: 0, isHuman: false, allIn: true }],
    });
    expect(result).toContain('[All-in]');
  });

  it('renders isFinished status', () => {
    const result = formatGenericState({
      title: 'Game',
      players: [{ id: 0, isHuman: false, isFinished: true }],
    });
    expect(result).toContain('[Finished]');
  });

  it('renders handName', () => {
    const result = formatGenericState({
      title: 'Game',
      players: [{ id: 0, isHuman: false, handName: 'Flush' }],
    });
    expect(result).toContain('[Flush]');
  });

  it('renders bid', () => {
    const result = formatGenericState({
      title: 'Game',
      players: [{ id: 0, isHuman: false, bid: 3 }],
    });
    expect(result).toContain('bid=3');
  });

  it('renders cardCount', () => {
    const result = formatGenericState({
      title: 'Game',
      players: [{ id: 0, isHuman: false, cardCount: 7 }],
    });
    expect(result).toContain('7 cards');
  });

  it('renders roundScore', () => {
    const result = formatGenericState({
      title: 'Game',
      players: [{ id: 0, isHuman: false, roundScore: 15 }],
    });
    expect(result).toContain('round=15');
  });

  it('renders human player cards as indexed', () => {
    const result = formatGenericState({
      title: 'Game',
      players: [
        {
          id: 0,
          isHuman: true,
          cards: [{ design: 'SPADE', value: 1 }],
        },
      ],
    });
    expect(result).toContain('[0]');
  });

  it('renders cpu cards on game end', () => {
    const result = formatGenericState({
      title: 'Game',
      gameEndFlag: true,
      players: [
        {
          id: 1,
          isHuman: false,
          cards: [{ design: 'HEART', value: 13 }],
        },
      ],
    });
    expect(result).toContain('\u2665K');
  });

  it('renders community cards', () => {
    const result = formatGenericState({
      title: 'Game',
      communityCards: [{ design: 'SPADE', value: 5 }],
    });
    expect(result).toContain('board:');
  });

  it('renders table cards', () => {
    const result = formatGenericState({
      title: 'Game',
      tableCards: [{ design: 'DIAMOND', value: 7 }],
    });
    expect(result).toContain('table:');
  });

  it('renders current trick', () => {
    const result = formatGenericState({
      title: 'Game',
      players: [{ id: 0, isHuman: true }],
      currentTrick: [{ playerIdx: 0, card: { design: 'CLOVER', value: 10 } }],
    });
    expect(result).toContain('trick:');
  });

  it('renders custom lines', () => {
    const result = formatGenericState({
      title: 'Game',
      customLines: ['custom info here'],
    });
    expect(result).toContain('custom info here');
  });

  it('renders turn info when not game end', () => {
    const result = formatGenericState({
      title: 'Game',
      currentTurn: 0,
      players: [{ id: 0, isHuman: true }],
    });
    expect(result).toContain('turn:');
  });

  it('does not render turn on game end', () => {
    const result = formatGenericState({
      title: 'Game',
      currentTurn: 0,
      gameEndFlag: true,
      players: [{ id: 0, isHuman: true }],
    });
    expect(result).not.toContain('turn:');
  });

  it('renders message', () => {
    const result = formatGenericState({ title: 'Game', message: 'You win!' });
    expect(result).toContain('You win!');
  });

  it('renders Game Over on game end', () => {
    const result = formatGenericState({ title: 'Game', gameEndFlag: true });
    expect(result).toContain('Game Over');
  });
});
