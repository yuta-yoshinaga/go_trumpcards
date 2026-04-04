import { useTranslation } from 'react-i18next';
import { focusRingWhite } from '../styles/buttonStyles';

interface SoundToggleProps {
  muted: boolean;
  onToggle: () => void;
}

/** SVG icon for volume on state. */
function VolumeOnIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M11 5 6 9H2v6h4l5 4zM19.07 4.93a10 10 0 0 1 0 14.14M15.54 8.46a5 5 0 0 1 0 7.07" />
    </svg>
  );
}

/** SVG icon for volume muted state. */
function VolumeMutedIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M11 5 6 9H2v6h4l5 4zM22 9l-6 6M16 9l6 6" />
    </svg>
  );
}

/** Renders a mute/unmute toggle button for the navigation bar. */
export function SoundToggle({ muted, onToggle }: SoundToggleProps) {
  const { t } = useTranslation('common');

  return (
    <button
      type="button"
      onClick={onToggle}
      aria-label={muted ? t('sound.unmute') : t('sound.mute')}
      className={`px-3 py-2 text-xs font-bold rounded min-h-[44px] min-w-[44px] flex items-center justify-center transition-colors bg-ds-surface-elevated text-ds-text-primary hover:bg-ds-surface ${focusRingWhite}`}
    >
      {muted ? <VolumeMutedIcon /> : <VolumeOnIcon />}
    </button>
  );
}
