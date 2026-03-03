interface CpuActionLogProps {
  actions: { playerIdx: number; action: number; amount: number }[] | undefined;
  actionNames: Record<number, string>;
}

export function CpuActionLog({ actions, actionNames }: CpuActionLogProps) {
  if (!actions || actions.length === 0) return null;
  return (
    <div className="bg-black/30 rounded p-2 mb-3 text-white text-[0.85em]">
      <div className="font-bold mb-1">CPU行動:</div>
      {actions.map((a, i) => (
        <div key={`${i}-${a.playerIdx}-${a.action}`}>
          Player {a.playerIdx}: {actionNames[a.action] ?? '不明'}
          {a.amount > 0 && ` (${a.amount})`}
        </div>
      ))}
    </div>
  );
}
