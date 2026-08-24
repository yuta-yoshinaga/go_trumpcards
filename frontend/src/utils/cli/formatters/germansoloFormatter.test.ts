import { describe, expect, it } from 'vitest';
import { makeGermanSoloState } from '../../../test/stateFactories';
import { formatGermanSoloState } from './germansoloFormatter';

describe('formatGermanSoloState', () => {
  it('renders the header, round/trick, bid, trump and per-player scores', () => {
    const out = formatGermanSoloState(makeGermanSoloState({ playerScores: [3, 1, 2, 0], winningBid: 2, trumpSuit: 1 }));
    expect(out).toContain('German Solo');
    expect(out).toContain('round: 1');
    expect(out).toContain('trick: 1');
    expect(out).toContain('bid: Frage');
    expect(out).toContain('trump: spade');
    expect(out).toContain('P0=3');
    expect(out).toContain('P1=1');
    expect(out).toContain('P2=2');
  });

  it('marks the declarer and the defenders', () => {
    const out = formatGermanSoloState(makeGermanSoloState());
    expect(out).toContain('Declarer');
    expect(out).toContain('Defender');
  });

  // **味方は呼ばれたエースが場に出るまで伏せる。** partnerIdx が返ってきて
  // 初めて Partner と呼ぶ。
  it('names the partner only once the server reveals it', () => {
    const hidden = formatGermanSoloState(makeGermanSoloState({ playsAlone: false, calledAceSuit: 3, partnerIdx: -1 }));
    expect(hidden).toContain('holder hidden');
    expect(hidden).not.toContain('Partner');
    const shown = formatGermanSoloState(makeGermanSoloState({ playsAlone: false, calledAceSuit: 3, partnerIdx: 2 }));
    expect(shown).toContain('partner is P2');
    expect(shown).toContain('Partner');
  });

  // **フェーズ名は phase 値で引く。** AceCall を並びから落とすと、以降が
  // 1 つずつずれて TrickEnd が「Play」と表示される。
  it('names every phase by its own value', () => {
    expect(formatGermanSoloState(makeGermanSoloState({ phase: 1 }))).toContain('phase: AceCall');
    expect(formatGermanSoloState(makeGermanSoloState({ phase: 2 }))).toContain('phase: Play');
    expect(formatGermanSoloState(makeGermanSoloState({ phase: 3 }))).toContain('phase: TrickEnd');
  });

  it('renders the contract target and the running trick split', () => {
    const out = formatGermanSoloState(makeGermanSoloState({ requiredTricks: 8, declarerTricks: 4, defenderTricks: 1 }));
    expect(out).toContain('need 8 tricks');
    expect(out).toContain('declarers 4 / defenders 1');
  });

  it('renders a Tout bid label', () => {
    const out = formatGermanSoloState(makeGermanSoloState({ winningBid: 4 }));
    expect(out).toContain('bid: Tout');
  });

  it('renders trump as dash when unset', () => {
    const out = formatGermanSoloState(makeGermanSoloState({ trumpSuit: -1 }));
    expect(out).toContain('trump: -');
  });

  it('renders the human hand with indices but not opponents', () => {
    const out = formatGermanSoloState(makeGermanSoloState());
    expect(out).toMatch(/\[0\]/);
  });

  it('renders the current trick when cards have been played', () => {
    const out = formatGermanSoloState(
      makeGermanSoloState({
        currentTrick: [
          { playerIdx: 0, card: { design: 'HEART', value: 12 } },
          { playerIdx: 1, card: { design: 'SPADE', value: 1 } },
        ],
      }),
    );
    expect(out).toContain('trick:');
  });

  it('renders the outcome during RoundEnd', () => {
    const out = formatGermanSoloState(makeGermanSoloState({ phase: 4, outcome: 2 }));
    expect(out).toContain('round result: failed');
  });

  it('renders a hint with card indices', () => {
    const out = formatGermanSoloState(
      makeGermanSoloState({
        hint: { cardIndices: [1, 2], reason: 'follow_win' },
        messageCode: 'germansolo.hintRequested',
      }),
    );
    expect(out).toContain('HINT: card indices [1, 2]');
    expect(out).toContain('follow_win');
  });

  it('renders the game-over banner with the winning player', () => {
    const out = formatGermanSoloState(makeGermanSoloState({ phase: 5, gameEndFlag: true, winnerPlayer: 1 }));
    expect(out).toContain('Game Over! Winner: Player 1');
  });

  it('renders an explicit message when present', () => {
    const out = formatGermanSoloState(makeGermanSoloState({ message: 'hello world' }));
    expect(out).toContain('hello world');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { cardIndices: [1, 2], reason: 'follow_win' };
    expect(formatGermanSoloState(makeGermanSoloState({ hint, messageCode: 'germansolo.hintRequested' }))).toContain(
      'HINT',
    );
    expect(formatGermanSoloState(makeGermanSoloState({ hint, messageCode: 'germansolo.playing' }))).not.toContain(
      'HINT',
    );
  });
});
