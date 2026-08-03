import type { ShengJiResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Declare', 'Kitty', 'Play', 'HandEnd', 'GameEnd'];

const COMBO_NAMES = ['-', 'single', 'pair', 'tractor'];

const SUIT_NAMES: Readonly<Record<number, string>> = { 0: 'none', 1: 'S', 2: 'C', 3: 'H', 4: 'D' };

/** Name a level. **2-A**, so the face levels need letters. */
function levelName(level: number): string {
  return { 11: 'J', 12: 'Q', 13: 'K', 14: 'A' }[level] ?? String(level);
}

/** Format a Sheng Ji game state as terminal text. */
export function formatShengJiState(state: ShengJiResponse): string {
  const lines: string[] = [];
  const defenders = 1 - state.declarerTeam;

  lines.push(formatHeader('Sheng Ji'));
  lines.push(`phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(
    `hand ${state.handNumber} | level: ${levelName(state.level)} | trump: ${SUIT_NAMES[state.trumpSuit] ?? '?'} | ` +
      `declarers: team ${state.declarerTeam} | team0: ${levelName(state.teamLevels[0])} | ` +
      `team1: ${levelName(state.teamLevels[1])}`,
  );
  // **切札は切札スートだけではない。**これが読めないと序列が分からない。
  lines.push(
    `TRUMPS ARE NOT JUST THE TRUMP SUIT: every ${levelName(state.level)} in all four suits, ` +
      'plus all four jokers, are trumps too',
  );
  // **点を集めるのは守備側。**
  lines.push(
    `THE DEFENDERS (team ${defenders}) COLLECT: ${state.teamPoints[defenders] ?? 0} of ${state.totalPoints}; ` +
      `${state.defenderTarget} takes the deal`,
  );
  lines.push('----------');

  state.players.forEach((p, i) => {
    const name = formatPlayerName(i, p.isHuman);
    // **味方の手札も伏せたまま。**
    const hand = p.cards.length > 0 ? p.cards.map(formatCard).join(' ') : `hidden (${p.cardCount})`;
    const turn = p.isCurrentTurn && !state.gameEndFlag ? ' <- turn' : '';
    const role = p.isDeclarer ? 'declarer' : 'defender';
    lines.push(`seat ${i} ${name}(T${p.team} ${role})${turn}: ${p.cardCount} cards  ${hand}`);
  });
  lines.push('----------');

  if (state.trick.length === 0) {
    lines.push('trick: nobody has played yet');
  } else {
    if (state.leadCombo) {
      const kind = COMBO_NAMES[state.leadCombo.kind] ?? String(state.leadCombo.kind);
      lines.push(`trick: a ${kind} (${state.leadCombo.size} cards) was led`);
    }
    for (const play of state.trick) {
      lines.push(`  seat ${play.seat}: ${play.cards.map(formatCard).join(' ')}`);
    }
  }

  if (state.phase === 0) {
    if (state.declaration) {
      const d = state.declaration;
      lines.push(`declared: seat ${d.seat} showed ${SUIT_NAMES[d.suit] ?? '?'} (strength ${d.strength})`);
    } else {
      lines.push('nobody has declared yet -- if everyone passes the hand is played with NO TRUMP SUIT');
    }
    const offers = Object.entries(state.declarableSuits).map(
      ([suit, strength]) => `${suit}=${SUIT_NAMES[Number(suit)] ?? '?'}(x${strength})`,
    );
    lines.push(offers.length > 0 ? `you can declare: ${offers.join(' ')}` : 'you hold no level card to declare with');
    lines.push('("d <0-4>" to declare; 0 passes, and only a stronger showing overrides)');
  }

  if (state.phase === 1) {
    lines.push(`("b <idx x${state.kittySizeMax}>" to bury -- keep your points and trumps out of the kitty)`);
  }

  if (state.phase === 2 && state.currentPlayerIdx === 0 && !state.gameEndFlag) {
    lines.push('(your turn -- "p <idx...>"; you must follow the led suit while you hold it)');
  }

  if (state.phase === 3 && state.lastResult) {
    const r = state.lastResult;
    lines.push(
      r.declarerHeld
        ? `the declarers held (${r.defenderPoints} of ${state.defenderTarget}); team ${r.advancingTeam} climbs ${r.advance}`
        : `the defenders collected ${r.defenderPoints}; the deal changes hands and team ${r.advancingTeam} climbs ${r.advance}`,
    );
    // **底牌の倍率は最終トリックを取った側にしか掛からない。**
    if (r.kittyMultiplier > 0) {
      lines.push(`  kitty: ${r.kittyPoints} points (x${r.kittyMultiplier}) went to the defenders`);
    }
    lines.push('("n" for the next hand)');
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push(`Game over! Winning team: ${state.winnerTeam}`);

  lines.push(formatSeparator());
  return lines.join('\n');
}
