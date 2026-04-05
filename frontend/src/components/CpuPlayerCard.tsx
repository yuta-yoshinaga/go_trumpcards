import { useTranslation } from 'react-i18next';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { handNameBadgeClass } from '../styles/gameConstants';
import type { Card } from '../types/card';
import { deriveAdaptationLevel, deriveStrategyStyle } from '../utils/metaAiAdaptation';
import { CardBack, CardImage } from './CardImage';
import { MetaAiIndicator } from './MetaAiIndicator';
import { MetaAiToast } from './MetaAiToast';

/** Meta-AI data for a CPU player. */
export interface CpuMetaAiProp {
  /** Whether meta-AI is enabled for this CPU. */
  enabled: boolean;
  /** Number of games the meta-AI has observed. */
  gamesPlayed: number;
  /** Bluff rate (used by betting/doubt games). */
  bluffRate?: number;
  /** Fold rate (used by betting games). */
  foldRate?: number;
  /** Edge pick rate (used by OldMaid). */
  edgePickRate?: number;
}

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
  /** When true, show text card count instead of face-down card images. */
  compactFaceDown?: boolean;
  /** Optional meta-AI data for visual feedback. */
  metaAi?: CpuMetaAiProp;
}

/** Renders a CPU player's info area with cards (face-up or face-down) and status. */
export function CpuPlayerCard({
  player,
  showCards,
  faceDownCount,
  showHandName,
  extraInfo,
  compactFaceDown,
  metaAi,
}: CpuPlayerCardProps) {
  const { t } = useTranslation('common');
  const { cpuCardWidth } = useCardDimensions();
  const showFaceUp = showCards && !player.folded && player.cards.length > 0;
  const strategyStyle = metaAi?.enabled ? deriveStrategyStyle(metaAi) : undefined;
  return (
    <div className="mb-3 rounded-lg p-2 bg-black/20 border border-white/10">
      {metaAi?.enabled && <MetaAiToast strategyStyle={strategyStyle} />}
      <div className="text-white text-sm mb-1">
        {t('player.cpu', { id: player.id })} <span className="text-gray-300 text-xs">({player.playStyleName})</span>
        {metaAi?.enabled && strategyStyle && (
          <MetaAiIndicator adaptationLevel={deriveAdaptationLevel(metaAi.gamesPlayed)} strategyStyle={strategyStyle} />
        )}
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
        {!showFaceUp && compactFaceDown && (
          <span className="ml-2 text-game-text-muted text-xs" data-testid="compact-card-count">
            🂠 {faceDownCount}
          </span>
        )}
      </div>
      {showFaceUp ? (
        <div className="flex flex-wrap gap-1">
          {player.cards.map((card) => (
            <CardImage
              key={`${card.design}-${card.value}`}
              card={card}
              width={cpuCardWidth}
              style={{ border: '3px solid transparent' }}
            />
          ))}
        </div>
      ) : !compactFaceDown ? (
        <div className="flex flex-wrap gap-1">
          {Array.from({ length: faceDownCount }).map((_, i) => (
            <CardBack key={i} width={cpuCardWidth} />
          ))}
        </div>
      ) : null}
    </div>
  );
}
