import { describe, expect, it } from 'vitest';
import { makeGanjifaState } from '../../../test/stateFactories';
import { formatGanjifaState } from './ganjifaFormatter';

describe('formatGanjifaState', () => {
  it('renders the header, round and trump', () => {
    const text = formatGanjifaState(makeGanjifaState({ trumpSuit: 2 }));
    expect(text).toContain('Ganjifa');
    expect(text).toContain('round: 1');
    expect(text).toContain('trick: 1');
    expect(text).toContain('Shamsher');
  });

  // The rank direction is the game's one non-obvious rule, so the terminal view
  // has to state it — and state the correct one.
  it('says higher wins for a strong trump', () => {
    const text = formatGanjifaState(makeGanjifaState({ trumpSuit: 4 }));
    expect(text).toContain('higher numbers are stronger');
    expect(text).not.toContain('lower numbers are stronger');
  });

  it('says lower wins for a weak trump', () => {
    const text = formatGanjifaState(makeGanjifaState({ trumpSuit: 5 }));
    expect(text).toContain('lower numbers are stronger');
    expect(text).not.toContain('higher numbers are stronger');
  });

  it('lists the phase name', () => {
    expect(formatGanjifaState(makeGanjifaState({ phase: 1 }))).toContain('phase: TrickEnd');
    expect(formatGanjifaState(makeGanjifaState({ phase: 2 }))).toContain('phase: RoundEnd');
  });

  it('shows the human hand with indices and hides the CPU hands', () => {
    const text = formatGanjifaState(makeGanjifaState());
    expect(text).toContain('[0]');
    expect(text).toContain('cards=32');
  });

  it('renders the current trick', () => {
    const text = formatGanjifaState(
      makeGanjifaState({
        currentTrick: [
          { playerIdx: 1, card: { design: 'JOKER', value: 3, glyph: '♪', label: '3', color: 'blue', deck: 'ganjifa' } },
        ],
      }),
    );
    expect(text).toContain('trick:');
    expect(text).toContain('♪3');
  });

  it('shows the round tally at round end', () => {
    const text = formatGanjifaState(makeGanjifaState({ phase: 2, roundTricks: [14, 10, 8] }));
    expect(text).toContain('round result: tricks P0=14 P1=10 P2=8');
  });

  it('omits the round tally mid-play', () => {
    expect(formatGanjifaState(makeGanjifaState())).not.toContain('round result');
  });

  it('shows the hint only once it was requested', () => {
    const hint = { cardIndices: [2], reason: 'lead_high' };
    expect(formatGanjifaState(makeGanjifaState({ hint }))).not.toContain('HINT:');
    const requested = formatGanjifaState(makeGanjifaState({ hint, messageCode: 'ganjifa.hintRequested' }));
    expect(requested).toContain('HINT: card indices [2] (lead_high)');
  });

  it('names the winner at game end', () => {
    const text = formatGanjifaState(makeGanjifaState({ phase: 3, gameEndFlag: true, winnerPlayer: 1 }));
    expect(text).toContain('Game Over! Winner: Player 1');
  });

  // A tie ends the match with winnerPlayer at -1; "Player -1" would be a lie.
  it('reports a draw rather than player -1', () => {
    const text = formatGanjifaState(makeGanjifaState({ phase: 3, gameEndFlag: true, winnerPlayer: -1 }));
    expect(text).toContain('Game Over! Draw.');
    expect(text).not.toContain('-1');
  });
});
