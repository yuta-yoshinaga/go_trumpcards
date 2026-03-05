import { useTranslation } from 'react-i18next';

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
  const { t } = useTranslation('common');
  if (!results || results.length === 0) return null;
  return (
    <div className="bg-black/30 rounded p-2 mb-3 text-white text-[0.85em]">
      <div className="font-bold mb-1">{t('label.result')}</div>
      {results.map((r) => (
        <div key={r.playerIdx}>
          {players[r.playerIdx]?.isHuman ? t('player.you') : `CPU ${r.playerIdx}`}
          {r.handName && `: ${r.handName}`}
          {r.wonAmount > 0 && (
            <span className="text-yellow-300 ml-1"> {t('label.chipsWon', { amount: r.wonAmount })}</span>
          )}
        </div>
      ))}
    </div>
  );
}
