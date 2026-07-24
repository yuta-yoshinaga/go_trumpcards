import { useTranslation } from 'react-i18next';

interface ErrorAlertProps {
  message: string | null;
  onRetry?: () => void;
}

/** Renders an error alert banner with optional retry button, hidden when message is null. */
export function ErrorAlert({ message, onRetry }: ErrorAlertProps) {
  const { t } = useTranslation('common');
  if (!message) return null;
  return (
    <div
      role="alert"
      className="bg-ds-error text-white text-center px-4 py-2 text-sm font-bold mb-2 rounded-lg flex items-center justify-center gap-2"
    >
      <span>{message}</span>
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className="px-3 py-1 text-xs font-medium bg-white text-ds-error hover:bg-ds-text-primary hover:text-ds-error-hover rounded transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/80 min-h-[44px] min-w-[44px] inline-flex items-center justify-center"
        >
          {t('button.retry')}
        </button>
      )}
    </div>
  );
}
