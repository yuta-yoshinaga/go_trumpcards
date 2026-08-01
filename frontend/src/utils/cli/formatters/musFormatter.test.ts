import { describe, expect, it } from 'vitest';
import { makeMusState } from '../../../test/stateFactories';
import { formatMusState } from './musFormatter';

describe('formatMusState', () => {
  it('renders header, round, phase, and amarrakos', () => {
    const out = formatMusState(makeMusState());
    expect(out).toContain('Mus');
    expect(out).toContain('round: 1');
    expect(out).toContain('phase: Grande');
    expect(out).toContain('amarrakos: team0=5  team1=3');
  });

  it('renders the human hand with indices', () => {
    const out = formatMusState(makeMusState());
    expect(out).toContain('[0]');
    expect(out).toContain('[3]');
    expect(out).toContain('cards=4');
  });

  it('shows a numeric pending bet', () => {
    const out = formatMusState(makeMusState({ pendingStake: 4 }));
    expect(out).toContain('pending bet: 4');
  });

  it('shows an ordago pending bet for -1', () => {
    const out = formatMusState(makeMusState({ pendingStake: -1 }));
    expect(out).toContain('pending bet: órdago');
  });

  it('omits the pending bet line when there is no pending stake', () => {
    const out = formatMusState(makeMusState({ pendingStake: 0 }));
    expect(out).not.toContain('pending bet');
  });

  it('falls back to the raw phase number for an unknown phase', () => {
    const out = formatMusState(makeMusState({ phase: 99 as unknown as 0 }));
    expect(out).toContain('phase: 99');
  });

  it('renders round results, skipping unresolved (team < 0) entries', () => {
    const out = formatMusState(
      makeMusState({
        results: [
          { kind: 0, stake: 2, team: 0 },
          { kind: 1, stake: 1, team: 1 },
          { kind: 2, stake: 0, team: -1 },
          { kind: 3, stake: 3, team: 0 },
        ],
      }),
    );
    expect(out).toContain('results:');
    expect(out).toContain('Grande: team0 +2');
    expect(out).toContain('Chica: team1 +1');
    expect(out).toContain('Juego: team0 +3');
    expect(out).not.toContain('Pares:');
  });

  it('omits the results line when no round has resolved', () => {
    const out = formatMusState(makeMusState());
    expect(out).not.toContain('results:');
  });

  it('renders a hint with indices and reason', () => {
    const out = formatMusState(
      makeMusState({
        messageCode: 'mus.hintRequested',
        hint: { mus: false, action: 1, amount: 2, indices: [0, 1], reason: 'strong_hand' },
      }),
    );
    expect(out).toContain('HINT: action=1');
    expect(out).toContain('indices [0, 1]');
    expect(out).toContain('strong_hand');
  });

  it('renders a hint without indices', () => {
    const out = formatMusState(
      makeMusState({
        messageCode: 'mus.hintRequested',
        hint: { mus: true, action: 0, amount: 0, indices: [], reason: 'weak_hand' },
      }),
    );
    expect(out).toContain('HINT: action=0');
    expect(out).not.toContain('indices [');
    expect(out).toContain('weak_hand');
  });

  it('renders the message line', () => {
    const out = formatMusState(makeMusState({ message: 'You won Grande' }));
    expect(out).toContain('You won Grande');
  });

  it('renders the game over banner when ended', () => {
    const out = formatMusState(makeMusState({ gameEndFlag: true, winnerTeam: 1 }));
    expect(out).toContain('Game Over! Winner: team1');
  });

  it('omits the game over banner when there is no winner', () => {
    const out = formatMusState(makeMusState({ gameEndFlag: true, winnerTeam: -1 }));
    expect(out).not.toContain('Game Over');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { mus: false, action: 1, amount: 2, indices: [0, 1], reason: 'strong_hand' };
    expect(formatMusState(makeMusState({ hint, messageCode: 'mus.hintRequested' }))).toContain('HINT');
    expect(formatMusState(makeMusState({ hint, messageCode: 'mus.playing' }))).not.toContain('HINT');
  });
});
