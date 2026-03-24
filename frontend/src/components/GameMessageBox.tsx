import { useTranslation } from 'react-i18next';

interface GameMessageBoxProps {
  message: string | undefined;
  messageCode?: string;
  messageParams?: Record<string, string>;
  alwaysVisible?: boolean;
}

/** Renders a game message box with i18n translation support via messageCode. */
export function GameMessageBox({ message, messageCode, messageParams, alwaysVisible = false }: GameMessageBoxProps) {
  const { t } = useTranslation('common');
  let displayMessage = message ?? '';
  if (messageCode) {
    const translated = t(`messageCode.${messageCode}`, messageParams ?? {});
    if (translated !== `messageCode.${messageCode}`) {
      displayMessage = translated;
    }
  }
  if (!alwaysVisible && !displayMessage) return null;
  return (
    <div
      role="status"
      aria-live="polite"
      className="glass-panel rounded-lg text-white text-center px-4 py-2 text-lg font-bold mb-2"
    >
      {displayMessage}
    </div>
  );
}
