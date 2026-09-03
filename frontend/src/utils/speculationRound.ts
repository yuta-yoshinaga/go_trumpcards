import { SpeculationPhase } from '../types/phases';

/**
 * Returns the round number to display for a Speculation state.
 *
 * `roundNo` は決着で 1 進むので、そのまま +1 すると round 1 の結果画面に
 * 「ラウンド 2/5」と出て、まだ始めていない回の結果を見ているように読める。
 * GAME_END では rounds を超えた番号 (6/5) にもなる。
 *
 * **同じ分岐が3面にある。** Go の CUI (`speculationDisplayRound`)、React の
 * ページ、CLI モードのフォーマッタ。フロント側の2面が手で同期を取っていて
 * 1面だけ直し漏れる事故が起きたので、ここに1つだけ置いて両方から呼ぶ (#6607)。
 */
export function speculationDisplayRound(phase: number | undefined, roundNo: number): number {
  return phase === SpeculationPhase.RESULT || phase === SpeculationPhase.GAME_END ? roundNo : roundNo + 1;
}
