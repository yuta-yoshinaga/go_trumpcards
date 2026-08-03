import { describe, expect, it } from 'vitest';
import { makeTuteState } from '../../../test/stateFactories';
import { formatTuteState } from './tuteFormatter';

describe('formatTuteState', () => {
  it('renders the header, round/trick, trump and team scores', () => {
    const out = formatTuteState(makeTuteState({ teamScores: [12, 7] }));
    expect(out).toContain('Tute');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('trump: ♦');
    expect(out).toContain('team scores: A=12  B=7');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatTuteState(makeTuteState());
    // human hand line begins after the player name with indexed cards
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatTuteState(
      makeTuteState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders the round result during RoundEnd', () => {
    const out = formatTuteState(makeTuteState({ phase: 2, roundTeamPoints: [70, 60] }));
    expect(out).toContain('round result: A points=70 B points=60');
  });

  it('renders a hint with marriage suit', () => {
    const out = formatTuteState(
      makeTuteState({
        hint: { cardIndices: [1, 2], marriage: 3, reason: 'declare_marriage' },
        messageCode: 'tute.hintRequested',
      }),
    );
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('marriage=♥');
    expect(out).toContain('declare_marriage');
  });

  it('renders a hint without marriage when marriage is 0', () => {
    const out = formatTuteState(
      makeTuteState({ hint: { cardIndices: [0], marriage: 0, reason: 'lead_low' }, messageCode: 'tute.hintRequested' }),
    );
    expect(out).toContain('HINT: card indices [0]');
    expect(out).not.toContain('marriage=');
  });

  it('renders the game-over banner with the winning team', () => {
    const out = formatTuteState(makeTuteState({ phase: 3, gameEndFlag: true, winnerTeam: 1 }));
    expect(out).toContain('Game Over! Winner: Team B');
  });

  it('renders an explicit message when present', () => {
    const out = formatTuteState(makeTuteState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [0], marriage: 0, reason: 'lead_low' };
    expect(formatTuteState(makeTuteState({ hint, messageCode: 'tute.hintRequested' }))).toContain('HINT');
    expect(formatTuteState(makeTuteState({ hint, messageCode: 'tute.playing' }))).not.toContain('HINT');
  });
});
