import { useTranslation } from 'react-i18next';
import type { Card } from '../../types/card';
import { cardLabel } from '../../utils/cardUtils';
import { findPlayerName } from '../../utils/playerUtils';

/** i18n namespaces that carry `exchange.title` / `exchange.entry`. */
export type ExchangeLogNamespace = 'daifugo' | 'president';

/** One "player A handed these cards to player B" record. */
export interface ExchangeLogAction {
  fromPlayerIdx: number;
  toPlayerIdx: number;
  cards: Card[];
}

/**
 * Renders the round-start card exchange log between ranked players.
 *
 * **大富豪とプレジデントは同じ交換ルールを持ち、レスポンスの型も同形なのに、
 * 表示は Daifugo にしかなかった (#4745)。**文言はゲームごとに違うので、
 * 名前空間だけ受け取って `exchange.title` / `exchange.entry` を引く。
 */
export function ExchangeLog({
  ns,
  players,
  actions,
}: {
  /**
   * i18n namespace holding `exchange.title` / `exchange.entry`.
   *
   * 素の string にすると綴り違い (`preisdent` 等) が型検査を素通りして、
   * i18next のキー欠落フォールバックとして黙って出てしまう。ゲームを増やす
   * ときは、ここに足す = その locale に `exchange.*` を用意する合図。
   */
  ns: ExchangeLogNamespace;
  players: { id: number; isHuman: boolean }[];
  actions: ExchangeLogAction[];
}) {
  const { t } = useTranslation(ns);
  return (
    <div
      className="bg-black/40 rounded-lg text-game-cpu-label-light py-2 px-3.5 my-2 whitespace-pre-line text-xs"
      data-testid="exchange-log"
    >
      {[
        t('exchange.title'),
        ...actions.map((a) => {
          const from = findPlayerName(players, a.fromPlayerIdx);
          const to = findPlayerName(players, a.toPlayerIdx);
          const cards = a.cards.map(cardLabel).join(', ');
          return t('exchange.entry', { from, to, cards });
        }),
      ].join('\n')}
    </div>
  );
}
