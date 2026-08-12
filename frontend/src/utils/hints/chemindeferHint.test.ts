import { describe, expect, it } from 'vitest';
import type { ChemindeFerResponse } from '../../types/card';
import { ChemindeFerPhase } from '../../types/phases';
import { getChemindeferHint } from './chemindeferHint';

const base = {
  phase: ChemindeFerPhase.STAKE,
  isHumanTurn: true,
  gameEndFlag: false,
  punterMayChoose: false,
} as unknown as ChemindeFerResponse;

const at = (over: Partial<ChemindeFerResponse>) => ({ ...base, ...over }) as ChemindeFerResponse;

describe('getChemindeferHint', () => {
  it('終局後は助言しない', () => {
    expect(getChemindeferHint(at({ gameEndFlag: true, phase: ChemindeFerPhase.BANKER_DRAW }))).toBeNull();
  });

  it('自分の手番でなければ助言しない', () => {
    expect(getChemindeferHint(at({ isHumanTurn: false, phase: ChemindeFerPhase.BANKER_DRAW }))).toBeNull();
  });

  it('張りや賭けの場面では助言しない', () => {
    expect(getChemindeferHint(at({ phase: ChemindeFerPhase.STAKE }))).toBeNull();
    expect(getChemindeferHint(at({ phase: ChemindeFerPhase.BET }))).toBeNull();
    expect(getChemindeferHint(at({ phase: ChemindeFerPhase.ROUND_END }))).toBeNull();
  });

  // **選べない合計で助言しない。**
  //
  // 0-4 と 6-7 はサーバが規則どおりに進めるので、そこに助言を出すと
  // 「プレイヤーが決められる」という誤った印象を与える。
  it('子が選べない合計では助言しない', () => {
    expect(getChemindeferHint(at({ phase: ChemindeFerPhase.PUNTER_DRAW, punterMayChoose: false }))).toBeNull();
  });

  it('子の合計 5 では引くのを薦める', () => {
    const hint = getChemindeferHint(at({ phase: ChemindeFerPhase.PUNTER_DRAW, punterMayChoose: true }));
    expect(hint).not.toBeNull();
    expect(hint?.targetAction).toBe('draw');
    expect(hint?.reason).toBe('frontendHint.chemindeferPunterFive');
  });

  // **親はどの合計でも選べるので、常に助言の余地がある。**
  it('親の判断では常に助言する', () => {
    const hint = getChemindeferHint(at({ phase: ChemindeFerPhase.BANKER_DRAW, punterMayChoose: false }));
    expect(hint).not.toBeNull();
    expect(hint?.reason).toBe('frontendHint.chemindeferBankerFree');
  });
});
