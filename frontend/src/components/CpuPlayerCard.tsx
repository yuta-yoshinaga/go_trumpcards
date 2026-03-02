import { handNameBadgeStyle } from '../styles/gameConstants';
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

export function CpuPlayerCard({ player, showCards, faceDownCount, showHandName, extraInfo }: CpuPlayerCardProps) {
  return (
    <div className="mb-3">
      <div className="text-white text-[0.95em] mb-1">
        CPU {player.id} <span className="text-gray-300 text-[0.85em]">({player.playStyleName})</span>
        <span className="ml-2 text-[0.85em]">チップ: {player.chips}</span>
        {extraInfo}
        {player.currentBet > 0 && <span className="ml-2 text-[0.85em]">ベット: {player.currentBet}</span>}
        {player.folded && <span className="ml-2 text-red-300 text-[0.85em]">[フォールド]</span>}
        {player.allIn && <span className="ml-2 text-yellow-300 text-[0.85em]">[オールイン]</span>}
        {showHandName && !player.folded && player.handName && (
          <span className="inline-block ml-2 text-[0.85em] font-bold rounded px-2 py-0.5" style={handNameBadgeStyle}>
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
                width={50}
                style={{ border: '3px solid transparent' }}
              />
            ))
          : Array.from({ length: faceDownCount }).map((_, i) => (
              // biome-ignore lint/suspicious/noArrayIndexKey: placeholder
              <CardBack key={i} width={50} />
            ))}
      </div>
    </div>
  );
}
