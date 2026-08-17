import { describe, expect, it } from 'vitest';
import { chineseTenHandBlockedReason, chineseTenLayoutBlockedReason } from './chineseTenDisabledReason';

const playing = { ended: false, isHumanTurn: true, choosing: false };

describe('chineseTenHandBlockedReason', () => {
  it('lets a card be played on your own turn', () => {
    expect(chineseTenHandBlockedReason(playing)).toBeNull();
  });

  it('names the CPU turn', () => {
    expect(chineseTenHandBlockedReason({ ...playing, isHumanTurn: false })).toBe('notYourTurn');
  });

  it('names the pending selection while it is yours to make', () => {
    expect(chineseTenHandBlockedReason({ ...playing, choosing: true })).toBe('chooseFirst');
  });

  // **`choosing` は相手が選んでいる間も真。**手番を先に見ないと、相手の選択中に
  // 「あなたが取る札を選んでください」と言ってしまう。
  it('names the turn, not the selection, while the opponent chooses', () => {
    expect(chineseTenHandBlockedReason({ ended: false, isHumanTurn: false, choosing: true })).toBe('notYourTurn');
  });

  it('names the end of the game before anything else', () => {
    expect(chineseTenHandBlockedReason({ ended: true, isHumanTurn: false, choosing: true })).toBe('gameOver');
  });
});

describe('chineseTenLayoutBlockedReason', () => {
  it('lets a matching card be taken while choosing', () => {
    expect(chineseTenLayoutBlockedReason({ ...playing, choosing: true }, true)).toBeNull();
  });

  // **手を出す前は「取れない」ではなく「まだ訊かれていない」。**同じ場札が
  // 一手あとには取れるので、取れないと言うのは誤解を招く。
  it('asks for a card to be led first', () => {
    expect(chineseTenLayoutBlockedReason(playing, false)).toBe('playFirst');
    expect(chineseTenLayoutBlockedReason(playing, true)).toBe('playFirst');
  });

  it('names the mismatch only once a card is led', () => {
    expect(chineseTenLayoutBlockedReason({ ...playing, choosing: true }, false)).toBe('noMatch');
  });

  it('names the CPU turn and the end of the game first', () => {
    expect(chineseTenLayoutBlockedReason({ ...playing, isHumanTurn: false, choosing: true }, true)).toBe('notYourTurn');
    expect(chineseTenLayoutBlockedReason({ ended: true, isHumanTurn: true, choosing: true }, true)).toBe('gameOver');
  });
});
