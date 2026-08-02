import { describe, expect, it } from 'vitest';
import { makeMichiganState } from '../../../test/stateFactories';
import { formatMichiganState } from './michiganFormatter';

describe('formatMichiganState', () => {
  it('includes the header, round, phase and ante', () => {
    const out = formatMichiganState(makeMichiganState());
    expect(out).toContain('Michigan');
    expect(out).toContain('round: 1');
    expect(out).toContain('phase: Bet');
    expect(out).toContain('ante: 8');
  });

  it('renders the four boodles with chips and claim status', () => {
    const out = formatMichiganState(makeMichiganState());
    expect(out).toContain('boodles:');
    expect(out).toContain('chips=2');
    expect(out).toContain('[open]');
  });

  it('marks a claimed boodle', () => {
    const state = makeMichiganState();
    state.boodles[0].claimedBy = 1;
    const out = formatMichiganState(state);
    expect(out).toContain('claimed by P1');
  });

  it('shows a new-sequence prompt when needNewSequence', () => {
    const out = formatMichiganState(makeMichiganState({ needNewSequence: true }));
    expect(out).toContain('start a new run');
  });

  it('shows the active sequence when a run is in progress', () => {
    const out = formatMichiganState(
      makeMichiganState({ needNewSequence: false, seqSuit: 1, seqSuitName: 'Hearts', seqHighValue: 7 }),
    );
    expect(out).toContain('Hearts up to 7');
  });

  it('renders each player with chips, bet, hand count and status', () => {
    const out = formatMichiganState(makeMichiganState());
    expect(out).toContain('chips=192');
    expect(out).toContain('bet=8');
    expect(out).toContain('hand=5');
    expect(out).toContain('[to play]');
  });

  it('marks a player that went out', () => {
    const state = makeMichiganState();
    state.players[0].isWinner = true;
    state.players[0].isCurrent = false;
    const out = formatMichiganState(state);
    expect(out).toContain('[WENT OUT]');
  });

  it('renders the playable indices and the backend hint line', () => {
    const out = formatMichiganState(
      makeMichiganState({ messageCode: 'michigan.hintRequested', hint: { cardIndex: 2, reason: 'claim_boodle' } }),
    );
    expect(out).toContain('playable:');
    expect(out).toContain('HINT: play 2 (claim_boodle)');
  });

  it('renders the game-over line with the match winner', () => {
    const out = formatMichiganState(makeMichiganState({ gameEndFlag: true, matchWinnerIdx: 0, phase: 2 }));
    expect(out).toContain('Game Over!');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndex: 2, reason: 'claim_boodle' };
    expect(formatMichiganState(makeMichiganState({ hint, messageCode: 'michigan.hintRequested' }))).toContain('HINT');
    expect(formatMichiganState(makeMichiganState({ hint, messageCode: 'michigan.playing' }))).not.toContain('HINT');
  });
});
