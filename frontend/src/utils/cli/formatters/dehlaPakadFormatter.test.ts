import { describe, expect, it } from 'vitest';
import { makeDehlaPakadState } from '../../../test/stateFactories';
import { formatDehlaPakadState } from './dehlaPakadFormatter';

const playState = makeDehlaPakadState({
  phase: 'play',
  isTrumpPhase: false,
  trumpSuit: 3,
  trumpSuitName: 'heart',
  currentPlayerIdx: 0,
  playableIndices: [0, 1, 2, 3, 4],
});

describe('formatDehlaPakadState', () => {
  it('prints the hand, phase and match target', () => {
    expect(formatDehlaPakadState(makeDehlaPakadState())).toContain('hand: 1  phase: selectTrump  target: 2 kot(s)');
  });

  it('prints the trump and trick once the trump is called', () => {
    const out = formatDehlaPakadState(playState);
    expect(out).toContain('trump: heart  trick: 1/13');
  });

  it('omits the trump line while it is still being called', () => {
    expect(formatDehlaPakadState(makeDehlaPakadState())).not.toContain('trump: ');
  });

  it('prints the tens and kots from the human seat', () => {
    const out = formatDehlaPakadState(
      makeDehlaPakadState({ ...playState, humanTeam: 1, teamTens: [3, 1], teamKots: [1, 0] }),
    );
    expect(out).toContain('tens: yours=1 theirs=3  kots: 0/1');
  });

  it('indexes the human hand', () => {
    expect(formatDehlaPakadState(playState)).toContain('[0]');
  });

  it('prints the cards already in the trick', () => {
    const base = makeDehlaPakadState();
    const out = formatDehlaPakadState(
      makeDehlaPakadState({ ...playState, currentTrick: [{ playerIdx: 1, card: base.players[0].cards[0] }] }),
    );
    expect(out).toContain('trick: CPU 1=');
  });

  // **これがこのゲームの心臓部。** 取っただけでは札は手に入らない。
  it('prints the centre pile and who would collect it', () => {
    const out = formatDehlaPakadState(
      makeDehlaPakadState({ ...playState, centrePileCount: 8, centrePileTens: 2, prevTrickWinner: 1 }),
    );
    expect(out).toContain('centre: 8 card(s), 2 ten(s)');
    expect(out).toContain('takes the pile by winning the next trick too');
  });

  it('omits the centre pile while nothing is waiting', () => {
    expect(formatDehlaPakadState(playState)).not.toContain('centre:');
  });

  it('says who calls the trump during that phase', () => {
    expect(formatDehlaPakadState(makeDehlaPakadState())).toContain('calls the trump');
    expect(formatDehlaPakadState(playState)).not.toContain('calls the trump');
  });

  it('reports a winning streak', () => {
    expect(formatDehlaPakadState(makeDehlaPakadState({ ...playState, streakTeam: 1, streakCount: 4 }))).toContain(
      'streak: team 1 has won 4 in a row',
    );
    expect(formatDehlaPakadState(makeDehlaPakadState({ ...playState, streakTeam: 1, streakCount: 1 }))).not.toContain(
      'streak:',
    );
  });

  it('prints the last hand, and names a kot', () => {
    const out = formatDehlaPakadState(
      makeDehlaPakadState({
        lastHand: { winnerTeam: 0, teamTens: [4, 0], kot: true, kotReason: 'allTens', dealerIdx: 3, trumpSuit: 3 },
      }),
    );
    expect(out).toContain('last hand: team 0 wins, tens 4-0 (KOT: allTens)');
  });

  // 頼んでいないヒントは出さない。
  it('gates the hint on the request', () => {
    const hint = { cardIndices: [2], reason: 'take_the_ten' };
    expect(formatDehlaPakadState(makeDehlaPakadState({ ...playState, hint, messageCode: '' }))).not.toContain('HINT:');
    expect(
      formatDehlaPakadState(makeDehlaPakadState({ ...playState, hint, messageCode: 'dehlapakad.hintRequested' })),
    ).toContain('HINT: card indices [2] (take_the_ten)');
  });

  it('gates the trump hint on the request too', () => {
    expect(
      formatDehlaPakadState(makeDehlaPakadState({ hintTrumpSuit: 3, messageCode: 'dehlapakad.hintRequested' })),
    ).toContain('HINT: call trump 3');
    expect(formatDehlaPakadState(makeDehlaPakadState({ hintTrumpSuit: 3, messageCode: '' }))).not.toContain('HINT:');
  });

  it('announces the winning team at the end', () => {
    expect(formatDehlaPakadState(makeDehlaPakadState({ gameEndFlag: true, winnerTeam: 1 }))).toContain(
      'Game Over! Team 1 takes the match!',
    );
  });

  it('prints the server message', () => {
    expect(formatDehlaPakadState(makeDehlaPakadState({ message: '札を出してください。' }))).toContain(
      '札を出してください。',
    );
  });

  // 集計がまだ無い盤面でも 0 として出す (null を素通しにしない)。
  it('falls back to zero when the team arrays are missing', () => {
    const out = formatDehlaPakadState(
      makeDehlaPakadState({
        teamTens: undefined as unknown as number[],
        teamKots: undefined as unknown as number[],
      }),
    );
    expect(out).toContain('tens: yours=0 theirs=0  kots: 0/0');
  });

  it('omits the hand line for a seat holding no cards', () => {
    const base = makeDehlaPakadState();
    const out = formatDehlaPakadState(
      makeDehlaPakadState({ players: base.players.map((p) => ({ ...p, cards: [], cardCount: 0 })) }),
    );
    expect(out).toContain('cards=0');
    expect(out).not.toContain('[0]');
  });

  // 最初のトリックの前は、山があっても「誰が取る」相手が居ない。
  it('prints the pile without a collector before the first trick resolves', () => {
    const out = formatDehlaPakadState(
      makeDehlaPakadState({ ...playState, centrePileCount: 4, centrePileTens: 1, prevTrickWinner: -1 }),
    );
    expect(out).toContain('centre: 4 card(s), 1 ten(s)');
    expect(out).not.toContain('takes the pile');
  });

  it('prints a requested hint that carries no card indices', () => {
    const out = formatDehlaPakadState(
      makeDehlaPakadState({
        ...playState,
        hint: { cardIndices: undefined as unknown as number[], reason: 'next_hand' },
        messageCode: 'dehlapakad.hintRequested',
      }),
    );
    expect(out).toContain('HINT: card indices [] (next_hand)');
  });
});
