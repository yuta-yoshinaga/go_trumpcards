import type { CanfieldResponse } from '../types/games/canfield';

/**
 * Whether Canfield's auto-complete can run right now.
 *
 * Sync: `Canfield.AutoComplete` (`internal/domain/Canfield.go`), which rejects the
 * command unless the reserve, stock and waste are all empty.
 *
 * **サーバが弾く条件をそのまま読む。**別の条件で光らせると、押せる見た目のまま
 * エラーになる (#4787)。
 */
export function canfieldAutoCompleteReady(state: Pick<CanfieldResponse, 'reserve' | 'stockCount' | 'waste'>): boolean {
  return state.reserve.length === 0 && state.stockCount === 0 && state.waste.length === 0;
}
