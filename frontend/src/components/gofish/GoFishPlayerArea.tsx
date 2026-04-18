import { useTranslation } from 'react-i18next';
import type { GoFishPlayerData } from '../../types/card';
import { playerName } from '../../utils/playerUtils';

interface GoFishPlayerAreaProps {
  player: GoFishPlayerData;
  isSelected: boolean;
  onSelect: (idx: number) => void;
  disabled: boolean;
}

/** Renders an opponent's card count and book count, clickable to select as ask target. */
export function GoFishPlayerArea({ player, isSelected, onSelect, disabled }: GoFishPlayerAreaProps) {
  const { t } = useTranslation('gofish');
  const name = playerName(player.id, player.isHuman);

  return (
    <button
      type="button"
      onClick={() => onSelect(player.id)}
      disabled={disabled}
      className={`w-full mb-2 p-2 rounded text-left transition-colors ${
        isSelected ? 'bg-ds-warning/30 ring-2 ring-ds-warning' : 'bg-black/30 hover:bg-black/40'
      } ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
      aria-pressed={isSelected}
    >
      <div className="text-ds-text-muted text-sm">
        {name}: {t('deck', { count: player.cardCount })} | {t('books', { count: player.bookCount })}
      </div>
    </button>
  );
}
