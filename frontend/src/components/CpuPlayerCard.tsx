import { useTranslation } from 'react-i18next';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { handNameBadgeClass } from '../styles/gameConstants';
import type { Card } from '../types/card';
import { CardBack, CardImage } from './CardImage';

interface CpuPlayerCardProps {
  player: {
    id: number;
    playStyleName: string;
    chips: number;
    currentBet: number;
    folded: boolean;
    allIn: boolean;
    handName: string;
    cards: Card[];
  };
  showCards: boolean;
  faceDownCount: number;
  showHandName: boolean;
  extraInfo?: React.ReactNode;
}

/** Renders a CPU player's info area with cards (face-up or face-down) and status. */
export function CpuPlayerCard({ player, showCards, faceDownCount, showHandName, extraInfo }: CpuPlayerCardProps) {
  const { t } = useTranslation('common');
  const { cpuCardWidth } = useCardDimensions();
  return (
    <div className="mb-3 rounded-lg p-2 bg-black/20 border border-white/10">
      <div className="text-white text-sm mb-1">
        {t('player.cpu', { id: player.id })} <span className="text-gray-300 text-xs">({player.playStyleName})</span>
        <span className="ml-2 text-xs">
          {t('betting.chips')} {player.chips}
        </span>
        {extraInfo}
        {player.currentBet > 0 && (
          <span className="ml-2 text-xs">
            {t('betting.currentBet')} {player.currentBet}
          </span>
        )}
        {player.folded && <span className="ml-2 text-red-300 text-xs">[{t('status.folded')}]</span>}
        {player.allIn && <span className="ml-2 text-yellow-300 text-xs">[{t('status.allIn')}]</span>}
        {showHandName && !player.folded && player.handName && (
          <span className={`inline-block ml-2 text-xs font-bold rounded px-2 py-0.5 ${handNameBadgeClass}`}>
            {player.handName}
          </span>
        )}
      </div>
      <div className="flex flex-wrap gap-1">
        {showCards && !player.folded && player.cards.length
          ? player.cards.map((card) => (
              <CardImage
                key={`${card.design}-${card.value}`}
                card={card}
                width={cpuCardWidth}
                style={{ border: '3px solid transparent' }}
              />
            ))
          : Array.from({ length: faceDownCount }).map((_, i) => <CardBack key={i} width={cpuCardWidth} />)}
      </div>
    </div>
  );
}
