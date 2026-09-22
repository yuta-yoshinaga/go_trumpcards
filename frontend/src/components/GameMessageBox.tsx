import { useTranslation } from 'react-i18next';
import { resolveMessageCode } from '../utils/resolveMessageCode';

/** Props for {@link GameMessageBox}. */
export interface GameMessageBoxProps {
  message: string | undefined;
  messageCode?: string;
  messageParams?: Record<string, string>;
  alwaysVisible?: boolean;
  /** "info" (default): polite announcement, "alert": assertive announcement for game results. */
  severity?: 'info' | 'alert';
}

/** Renders a game message box with i18n translation support via messageCode. */
export function GameMessageBox({
  message,
  messageCode,
  messageParams,
  alwaysVisible = false,
  severity = 'info',
}: GameMessageBoxProps) {
  const { t } = useTranslation('common');
  const displayMessage = resolveMessageCode(t, messageCode, messageParams, message);
  if (!alwaysVisible && !displayMessage) return null;
  const role = severity === 'alert' ? 'alert' : 'status';
  const live = severity === 'alert' ? 'assertive' : 'polite';
  return (
    <div
      role={role}
      aria-live={live}
      className="glass-panel rounded-lg text-ds-text-primary text-center px-4 py-2 text-lg font-bold mb-2"
    >
      {displayMessage}
    </div>
  );
}
