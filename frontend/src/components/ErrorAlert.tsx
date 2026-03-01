interface ErrorAlertProps {
  message: string | null;
}

export function ErrorAlert({ message }: ErrorAlertProps) {
  if (!message) return null;
  return (
    <div
      role="alert"
      className="bg-red-700/90 text-white text-center px-4 py-2 text-[0.95em] font-bold mb-2 rounded-lg"
    >
      {message}
    </div>
  );
}
