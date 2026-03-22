interface ErrorAlertProps {
  message: string | null;
  onRetry?: () => void;
}

/** Renders an error alert banner with optional retry button, hidden when message is null. */
export function ErrorAlert({ message, onRetry }: ErrorAlertProps) {
  if (!message) return null;
  return (
    <div
      role="alert"
      className="bg-red-700/90 text-white text-center px-4 py-2 text-sm font-bold mb-2 rounded-lg flex items-center justify-center gap-2"
    >
      <span>{message}</span>
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className="px-2 py-0.5 text-xs bg-white/20 hover:bg-white/30 rounded transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/80"
        >
          Retry
        </button>
      )}
    </div>
  );
}
