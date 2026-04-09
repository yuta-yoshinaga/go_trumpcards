import { useTranslation } from 'react-i18next';

interface RoundResultEntry {
  playerIdx: number;
  handName: string;
  kickers?: string;
  wonAmount: number;
  mucked?: boolean;
}

interface RoundResultsProps {
  results: RoundResultEntry[] | undefined;
  players: { isHuman: boolean }[];
}

/** Renders round result summary showing each player's hand and winnings. */
export function RoundResults({ results, players }: RoundResultsProps) {
  const { t } = useTranslation('common');
  if (!results || results.length === 0) return null;
  return (
    <div className="bg-black/30 rounded p-2 mb-3 text-white text-xs">
      <div className="font-bold mb-1">{t('label.result')}</div>
      {results.map((r) => (
        <div key={r.playerIdx}>
          {players[r.playerIdx]?.isHuman ? t('player.you') : `CPU ${r.playerIdx}`}
          {r.mucked ? `: ${t('label.mucked')}` : r.handName && `: ${r.handName}`}
          {!r.mucked && r.kickers && ` (${t('label.kicker', { kickers: r.kickers })})`}
          {r.wonAmount > 0 && (
            <span className="text-ds-warning ml-1"> {t('label.chipsWon', { amount: r.wonAmount })}</span>
          )}
        </div>
      ))}
    </div>
  );
}
