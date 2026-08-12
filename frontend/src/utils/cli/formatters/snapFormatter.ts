import type { SnapResponse } from '../../../types/card';
import { SnapEventKind, SnapPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [SnapPhase.PLAY]: 'PLAY',
  [SnapPhase.GAME_END]: 'GAME END',
};

/** Format a Snap game state as terminal text. */
export function formatSnapState(state: SnapResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Snap'));
  lines.push(`pile ${state.centerPileSize} | ${PHASE_NAMES[state.phase] ?? state.phase}`);
  // **トリガーが動くことが規則そのもの。** 毎回書く。
  lines.push('call when the top card matches the one before it — one card showing is never a snap');

  lines.push(state.topCard ? `top card: ${formatCard(state.topCard)}` : 'pile: empty');
  // **成立しているかは一目で分かる必要がある。** 反射ゲームなので。
  if (state.snapAvailable) {
    lines.push('*** SNAP IS ON — press n ***');
  }

  lines.push('----------');

  state.players.forEach((p) => {
    const marker = p.id === state.currentTurnIdx && !state.gameEndFlag ? '>' : ' ';
    lines.push(`${marker}${formatPlayerName(p.id, p.isHuman)}: ${p.stockSize} in stock`);
  });

  // **直近に何が起きたかを出す。** 盤面だけでは誰が取ったのか読めない。
  const who = formatPlayerName(state.lastEventPlayerIdx, state.lastEventPlayerIdx === 0);
  switch (state.lastEventKind) {
    case SnapEventKind.SNAP_CORRECT:
      lines.push(`${who} called snap and took the pile`);
      break;
    case SnapEventKind.SNAP_WRONG:
      lines.push(`${who} called wrongly and paid a card`);
      break;
    case SnapEventKind.ELIMINATED:
      lines.push(`${who} has run out of stock`);
      break;
    case SnapEventKind.STEP:
      lines.push(`${who} turned a card over`);
      break;
    default:
      break;
  }

  if (state.gameEndFlag) {
    lines.push('----------');
    lines.push(
      state.winnerIdx >= 0
        ? `game over — ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)} wins`
        : 'game over — play could not continue',
    );
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
