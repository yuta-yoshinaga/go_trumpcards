import { useTranslation } from 'react-i18next';

interface SoundToggleProps {
  muted: boolean;
  onToggle: () => void;
}

/** Renders a mute/unmute toggle button for the navigation bar. */
export function SoundToggle({ muted, onToggle }: SoundToggleProps) {
  const { t } = useTranslation('common');

  return (
    <button
      type="button"
      onClick={onToggle}
      aria-label={muted ? t('sound.unmute') : t('sound.mute')}
      className="px-3 py-2 text-xs font-bold rounded min-h-[44px] transition-colors bg-gray-600 text-gray-200 hover:bg-gray-500"
    >
      {muted ? '🔇' : '🔊'}
    </button>
  );
}
