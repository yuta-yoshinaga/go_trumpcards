import type { ShengJiResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { ShengJiPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Sheng Ji (升级 / 拖拉机), or null when
 * there is nothing useful to say.
 *
 * The advice is phase-shaped rather than card-shaped, because the decisions a
 * newcomer gets wrong here are structural, not tactical:
 *
 * - **Declaring** is a race to show a level card, not an auction, and only a
 *   stronger showing overrides. If the player holds one, saying so is worth
 *   more than any card-level suggestion.
 * - **Burying** is where the game is most often lost outright: points and
 *   trumps left in the kitty come back multiplied if the defenders take the
 *   last trick.
 * - **In play**, which side you are on inverts your goal — the defenders
 *   collect, the declarers hold them under the target. The tooltip names the
 *   side's goal rather than a card, since the trump group (every level card in
 *   all four suits, plus the jokers) is the thing players misjudge.
 *
 * Card-level choice is left to the player: it depends on the whole trick
 * history and the partner's holdings, which the client cannot model.
 */
export function getShengJiHint(state: ShengJiResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  switch (state.phase) {
    case ShengJiPhase.DECLARE: {
      // Only worth prompting when the player actually can declare.
      const canDeclare = Object.keys(state.declarableSuits).length > 0;
      return {
        targetAction: 'declare',
        reason: canDeclare ? 'hint.shengji_declare' : 'hint.shengji_pass',
        confidence: canDeclare ? 'strong' : 'moderate',
      };
    }
    case ShengJiPhase.KITTY:
      return {
        targetAction: 'bury',
        reason: 'hint.shengji_bury',
        confidence: 'strong',
      };
    case ShengJiPhase.PLAY: {
      const human = state.players.find((p) => p.isHuman);
      if (!human) return null;
      return {
        targetAction: 'play',
        reason: human.isDeclarer ? 'hint.shengji_hold' : 'hint.shengji_collect',
        confidence: 'moderate',
      };
    }
    default:
      return null;
  }
}
