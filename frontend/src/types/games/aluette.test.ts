import { describe, expect, it } from 'vitest';
import type { AluetteLuette } from './aluette';
import { ALUETTE_HAND_SIZE, ALUETTE_PLAYER_COUNT, aluetteLuetteName, aluetteTeamOf } from './aluette';

const LUETTES: AluetteLuette[] = [
  { design: 'DIAMOND', value: 3, name: 'Monsieur' },
  { design: 'HEART', value: 3, name: 'Madame' },
  { design: 'HEART', value: 2, name: 'Borgne' },
  { design: 'DIAMOND', value: 2, name: 'Vache' },
  { design: 'HEART', value: 9, name: 'GrandNeuf' },
  { design: 'DIAMOND', value: 9, name: 'PetitNeuf' },
];

describe('aluetteTeamOf', () => {
  // 対面同士が組む。ここがずれると味方のトリックを奪う手が正しく見えてしまう。
  it('pairs the opposite seats', () => {
    expect(aluetteTeamOf(0)).toBe(aluetteTeamOf(2));
    expect(aluetteTeamOf(1)).toBe(aluetteTeamOf(3));
    expect(aluetteTeamOf(0)).not.toBe(aluetteTeamOf(1));
  });

  it('produces exactly two teams across every seat', () => {
    const teams = new Set(Array.from({ length: ALUETTE_PLAYER_COUNT }, (_, i) => aluetteTeamOf(i)));
    expect(teams.size).toBe(2);
  });
});

describe('aluetteLuetteName', () => {
  // **強さは値ではなく札で決まる。**同じ 3 でも金貨は最強、剣はただの札。
  // 値だけで引くとこの区別が消える。
  it('matches on suit and value together, not on value alone', () => {
    expect(aluetteLuetteName(LUETTES, { design: 'DIAMOND', value: 3 })).toBe('Monsieur');
    expect(aluetteLuetteName(LUETTES, { design: 'SPADE', value: 3 })).toBeUndefined();
    expect(aluetteLuetteName(LUETTES, { design: 'CLOVER', value: 9 })).toBeUndefined();
  });

  it('names every luette in the table', () => {
    for (const l of LUETTES) {
      expect(aluetteLuetteName(LUETTES, { design: l.design, value: l.value })).toBe(l.name);
    }
  });

  it('returns undefined against an empty table', () => {
    expect(aluetteLuetteName([], { design: 'DIAMOND', value: 3 })).toBeUndefined();
  });
});

describe('mene arithmetic', () => {
  // 5 トリック中 3 勝でメーヌを取る。過半数であることを固定する。
  it('needs a strict majority of the tricks to take a mene', () => {
    const toWin = Math.floor(ALUETTE_HAND_SIZE / 2) + 1;
    expect(toWin).toBe(3);
    expect(toWin * 2).toBeGreaterThan(ALUETTE_HAND_SIZE);
  });
});
