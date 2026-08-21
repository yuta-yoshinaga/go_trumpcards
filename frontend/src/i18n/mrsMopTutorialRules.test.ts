import { describe, expect, it } from 'vitest';
import en from './locales/en/mrsmop.json';
import ja from './locales/ja/mrsmop.json';

// ミセス・モップの配置規則は2段構えで、チュートリアルはそれを取り違えていた (#5503)。
//
//   canPlaceOnTableau   -- 1枚を1枚の上に置くのは「値が1つ小さい」だけが条件。**スート不問**
//   isValidMrsMopSequence -- 同スート降順が要るのは**まとめて動かすとき**だけ
//
// 旧文言は「同じスートの降順でカードを並べます」で、1枚配置にまで同スート制限が
// あるかのように読める。実際には合法な異スートへの配置をプレイヤーが避けてしまう。
// docs/manual/web/mrsmop.md は最初から正しく書き分けているので、そちらが基準。
describe('mrsMop tutorial text matches the actual placement rules', () => {
  it('says a single card ignores suit', () => {
    expect(ja.tutorial.tableau).toContain('スート不問');
    expect(en.tutorial.tableau.toLowerCase()).toContain('regardless of suit');
  });

  it('confines the same-suit requirement to moving a group', () => {
    // 「同スート降順」という語自体は正しい。誤りは**それが何に掛かるか**なので、
    // まとめ移動を指す語と同じ文に現れることまで見る。
    expect(ja.tutorial.tableau).toMatch(/まとめて.*同スート降順|同スート降順.*まとめて/);
    expect(en.tutorial.tableau.toLowerCase()).toMatch(/group|together/);
  });

  it('states the removal run in the order it is built (K down to A)', () => {
    // 旧文言は「AからK」で、実際に積み上がる向きと逆に読める。
    expect(ja.tutorial.tableau).toContain('K');
    expect(ja.tutorial.tableau).not.toContain('AからK');
    expect(en.tutorial.tableau.toLowerCase()).not.toContain('ace-to-king');
  });
});
