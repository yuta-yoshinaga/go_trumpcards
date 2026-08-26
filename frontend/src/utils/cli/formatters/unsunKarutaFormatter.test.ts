import { describe, expect, it } from 'vitest';
import { makeUnsunKarutaState } from '../../../test/stateFactories';
import { formatUnsunKarutaState } from './unsunKarutaFormatter';

describe('formatUnsunKarutaState', () => {
  it('prints the deal, trick, phase and trump', () => {
    const out = formatUnsunKarutaState(makeUnsunKarutaState());
    expect(out).toContain('deal: 1  trick: 1/9  phase: Play');
    // **切り札のスート名は必須。** 長物と丸物で数札の強弱が逆になるので、
    // どちらが切り札かが分からないと端末から強さを読めない。
    expect(out).toContain('trump: kotsu');
  });

  it('prints the ko count per team and the match score', () => {
    const out = formatUnsunKarutaState(makeUnsunKarutaState({ teamTricks: [4, 2], teamScores: [9, 7] }));
    expect(out).toContain('ko: team0=4 team1=2  match: 9/7');
  });

  it('lists every seat with its team and indexes the human hand', () => {
    const out = formatUnsunKarutaState(makeUnsunKarutaState());
    expect(out).toContain('(team 0 / Player): cards=4 tricks=0');
    expect(out).toContain('(team 1 / Dealer): cards=9 tricks=0');
    expect(out).toContain('[0]');
  });

  it('prints the cards already in the trick', () => {
    const base = makeUnsunKarutaState();
    const out = formatUnsunKarutaState(
      makeUnsunKarutaState({ currentTrick: [{ playerIdx: 1, card: base.players[0].cards[0] }] }),
    );
    expect(out).toContain('trick: CPU 1=棒9');
  });

  it('says when the follow obligation is standing', () => {
    expect(formatUnsunKarutaState(makeUnsunKarutaState({ mustFollow: true, canDeclare: false }))).toContain(
      'declared: must follow the led suit',
    );
    expect(formatUnsunKarutaState(makeUnsunKarutaState({ mustFollow: false, canDeclare: false }))).not.toContain(
      'declared:',
    );
  });

  it('offers the declaration only on a lead', () => {
    expect(formatUnsunKarutaState(makeUnsunKarutaState())).toContain('you lead: meri <idx> declares');
    expect(formatUnsunKarutaState(makeUnsunKarutaState({ canDeclare: false }))).not.toContain('you lead:');
  });

  // 頼んでいないヒントは出さない。
  it('gates the hint on the request', () => {
    const hint = { cardIndices: [2], reason: 'lead_strong' };
    expect(formatUnsunKarutaState(makeUnsunKarutaState({ hint, messageCode: '' }))).not.toContain('HINT:');
    expect(formatUnsunKarutaState(makeUnsunKarutaState({ hint, messageCode: 'unsunkaruta.hintRequested' }))).toContain(
      'HINT: card indices [2] (lead_strong)',
    );
  });

  it('announces the winning team, and a draw', () => {
    expect(formatUnsunKarutaState(makeUnsunKarutaState({ gameEndFlag: true, winnerTeam: 1 }))).toContain(
      'Game Over! Team 1 wins!',
    );
    expect(formatUnsunKarutaState(makeUnsunKarutaState({ gameEndFlag: true, winnerTeam: -1 }))).toContain(
      'Game Over! Draw!',
    );
  });

  // 切り札の札そのものは無いことがある (復元直後など)。スート名だけは必ず出す。
  it('prints the trump suit even with no turned-up card', () => {
    const out = formatUnsunKarutaState(makeUnsunKarutaState({ trumpCard: null }));
    expect(out).toContain('trump: kotsu');
    expect(out).not.toContain('trump: kotsu (');
  });

  // 累計や取り分がまだ無い盤面でも 0 として出す (null を素通しにしない)。
  it('falls back to zero when the team arrays are missing', () => {
    const out = formatUnsunKarutaState(
      makeUnsunKarutaState({
        teamTricks: undefined as unknown as number[],
        teamScores: undefined as unknown as number[],
      }),
    );
    expect(out).toContain('ko: team0=0 team1=0  match: 0/0');
  });

  it('omits the hand line for a seat holding no cards', () => {
    const base = makeUnsunKarutaState();
    const out = formatUnsunKarutaState(
      makeUnsunKarutaState({ players: base.players.map((p) => ({ ...p, cards: [], cardCount: 0 })) }),
    );
    expect(out).toContain('cards=0 tricks=0');
    expect(out).not.toContain('[0]');
  });

  it('prints the server message when there is one', () => {
    expect(formatUnsunKarutaState(makeUnsunKarutaState({ message: '好きな札を出せます。' }))).toContain(
      '好きな札を出せます。',
    );
  });

  it('names every phase it can be in', () => {
    for (const [phase, name] of [
      [0, 'Play'],
      [1, 'TrickEnd'],
      [2, 'RoundEnd'],
      [3, 'GameEnd'],
    ] as const) {
      expect(formatUnsunKarutaState(makeUnsunKarutaState({ phase }))).toContain(`phase: ${name}`);
    }
  });

  // 頼まれたヒントに札番号が無いこともある (次のトリックへ、など)。
  it('prints a requested hint that carries no card indices', () => {
    const out = formatUnsunKarutaState(
      makeUnsunKarutaState({
        hint: { cardIndices: undefined as unknown as number[], reason: 'next_trick' },
        messageCode: 'unsunkaruta.hintRequested',
      }),
    );
    expect(out).toContain('HINT: card indices [] (next_trick)');
  });
});
