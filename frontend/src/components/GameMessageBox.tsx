interface GameMessageBoxProps {
  message: string | undefined;
  alwaysVisible?: boolean;
}

export function GameMessageBox({ message, alwaysVisible = false }: GameMessageBoxProps) {
  if (!alwaysVisible && !message) return null;
  return (
    <div className="bg-black/55 rounded-lg text-white text-center px-4 py-2 text-[1.1em] font-bold mb-2">
      {message ?? ''}
    </div>
  );
}
