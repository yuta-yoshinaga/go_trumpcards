import { useTranslation } from 'react-i18next';
import type { DaifugoExchangeAction } from '../../types/card';
import { cardLabel } from '../../utils/cardUtils';
import { findPlayerName } from '../../utils/playerUtils';

export function DaifugoExchangeLog({
  players,
  actions,
}: {
  players: { id: number; isHuman: boolean }[];
  actions: DaifugoExchangeAction[];
}) {
  const { t } = useTranslation('daifugo');
  return (
    <div className="bg-black/40 rounded-lg text-[#ffd] py-2 px-3.5 my-2 whitespace-pre-line text-[0.85em]">
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
