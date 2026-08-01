import type { GuandanResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Tribute', 'Play', 'HandEnd', 'GameEnd'];

const COMBO_NAMES = [
  '-',
  'single',
  'pair',
  'triple',
  'full house',
  'straight',
  'plate (consecutive triples)',
  'tube (consecutive pairs)',
  'bomb',
  'straight flush',
  'joker bomb',
];

/** Name a level. **2-A**, so the face levels need letters. */
function levelName(level: number): string {
  return { 11: 'J', 12: 'Q', 13: 'K', 14: 'A' }[level] ?? String(level);
}

/** Format a Guandan game state as terminal text. */
export function formatGuandanState(state: GuandanResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Guandan'));
  lines.push(`phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(
    `hand ${state.handNumber} | level: ${levelName(state.level)} (set by team ${state.declarerTeam}) | ` +
      `team0: ${levelName(state.teamLevels[0])} | team1: ${levelName(state.teamLevels[1])}`,
  );
  // **レベル札が A より強い**のがこのゲームの肝。書かないと読めない。
  lines.push(
    `the ${levelName(state.level)}s BEAT ACES and lose only to the jokers; the two hearts among them are WILD`,
  );
  lines.push(
    `going out 1st+2nd climbs ${state.advanceFirstSecond}, 1st+3rd ${state.advanceFirstThird}, ` +
      `1st+4th ${state.advanceFirstFourth} — there is no climb of three`,
  );
  lines.push('----------');

  state.players.forEach((p, i) => {
    const name = formatPlayerName(i, p.isHuman);
    // **味方の手札も伏せたまま。**
    const hand = p.cards.length > 0 ? p.cards.map(formatCard).join(' ') : `hidden (${p.cardCount})`;
    const turn = p.isCurrentTurn && !state.gameEndFlag ? ' <- turn' : '';
    const out = p.finishedRank > 0 ? ` out#${p.finishedRank}` : '';
    lines.push(`seat ${i} ${name}(T${p.team}${out})${turn}: ${p.cardCount} cards  ${hand}`);
  });
  lines.push('----------');

  if (state.lastCombo) {
    const kind = COMBO_NAMES[state.lastCombo.kind] ?? String(state.lastCombo.kind);
    lines.push(`table: ${kind} (${state.lastCombo.size} cards) played by seat ${state.lastPlayerIdx}`);
  } else {
    lines.push('table: clear — lead anything you like');
  }

  if (state.phase === 0) {
    if (state.tributeCancelled) {
      lines.push('tribute is off this hand: a payer held both red jokers');
    } else {
      for (const x of state.tributes) {
        const paid = x.card ? formatCard(x.card) : '?';
        const back = x.returned ? formatCard(x.returned) : null;
        lines.push(
          back
            ? `tribute: seat ${x.from} -> seat ${x.to}, ${paid} / returned: ${back}`
            : `tribute: seat ${x.from} -> seat ${x.to}, ${paid} (awaiting the return)`,
        );
      }
    }
  }

  if (state.phase === 2 && state.lastResult) {
    const r = state.lastResult;
    const climbed = r.firstSecond ? ' (first AND second — the only way to climb four)' : '';
    lines.push(`team ${r.winnerTeam} advances ${r.advance} level(s)${climbed}`);
    lines.push('("n" for the next hand)');
  }

  if (state.phase === 1 && state.currentPlayerIdx === 0 && !state.gameEndFlag) {
    lines.push('(your turn — "p <index...>" to play, "ps" to pass)');
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push(`Game over! Winning team: ${state.winnerTeam}`);

  lines.push(formatSeparator());
  return lines.join('\n');
}
