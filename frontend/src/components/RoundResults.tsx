interface RoundResultEntry {
  playerIdx: number;
  handName: string;
  wonAmount: number;
}

interface RoundResultsProps {
  results: RoundResultEntry[] | undefined;
  players: { isHuman: boolean }[];
}

export function RoundResults({ results, players }: RoundResultsProps) {
  if (!results || results.length === 0) return null;
  return (
    <div className="bg-black/30 rounded p-2 mb-3 text-white text-[0.85em]">
      <div className="font-bold mb-1">結果:</div>
      {results.map((r) => (
        <div key={r.playerIdx}>
          {players[r.playerIdx]?.isHuman ? 'あなた' : `CPU ${r.playerIdx}`}
          {r.handName && `: ${r.handName}`}
          {r.wonAmount > 0 && <span className="text-yellow-300 ml-1"> +{r.wonAmount}チップ</span>}
        </div>
      ))}
    </div>
  );
}
