import { useTranslation } from 'react-i18next';
import { focusRingWhite } from '../../styles/buttonStyles';

/** SVG icon representing a terminal/CLI. */
function TerminalIcon() {
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
      <polyline points="4 17 10 11 4 5" />
      <line x1="12" y1="19" x2="20" y2="19" />
    </svg>
  );
}

/** SVG icon representing a graphical UI. */
function GuiIcon() {
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
      <rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
      <line x1="8" y1="21" x2="16" y2="21" />
      <line x1="12" y1="17" x2="12" y2="21" />
    </svg>
  );
}

/** Props for the CliToggle component. */
export interface CliToggleProps {
  /** Whether CLI mode is currently enabled. */
  cliEnabled: boolean;
  /** Callback to toggle CLI mode. */
  onToggle: () => void;
}

/** Renders a toggle button to switch between GUI and CLI mode. */
export function CliToggle({ cliEnabled, onToggle }: CliToggleProps) {
  const { t } = useTranslation('common');

  return (
    <button
      type="button"
      onClick={onToggle}
      aria-label={cliEnabled ? t('cli.switchToGui') : t('cli.switchToCli')}
      className={`px-3 py-2 text-xs font-bold rounded min-h-[44px] min-w-[44px] flex items-center justify-center gap-1 transition-colors bg-ds-surface-elevated text-ds-text-primary hover:bg-ds-surface-elevated-hover ${focusRingWhite}`}
    >
      {cliEnabled ? (
        <>
          <GuiIcon />
          <span>GUI</span>
        </>
      ) : (
        <>
          <TerminalIcon />
          <span>CLI</span>
        </>
      )}
    </button>
  );
}
