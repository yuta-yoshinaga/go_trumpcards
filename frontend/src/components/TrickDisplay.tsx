import type { Card } from '../types/card';
import { playerName } from '../utils/playerUtils';
import { AnimatedCard } from './motion/AnimatedCard';

/** One card played into the current trick. Matches all trick-taking game TrickCard shapes. */
export interface TrickDisplayCard {
  playerIdx: number;
  card: Card;
}

/** Minimal player reference used for display-name resolution. */
export interface TrickDisplayPlayer {
  id: number;
  isHuman: boolean;
  /** Team identifier (0/1) for partner-based games. Omitted for free-for-all games. */
  team?: number;
}

/** Props for {@link TrickDisplay}. */
export interface TrickDisplayProps {
  /** Cards currently on the trick. When empty, the component renders nothing. */
  currentTrick: TrickDisplayCard[];
  /** All players; indexed by `trickCard.playerIdx`. */
  players: TrickDisplayPlayer[];
  /** Card width in px forwarded to {@link AnimatedCard}. */
  cardWidth: number;
  /** Localised label, e.g. `t('currentTrick')`. */
  label: string;
  /** Value for the `data-tutorial` attribute (e.g. `"ht-trick-display"`). */
  dataTutorial?: string;
  /** When set (e.g. the trick winner at TRICK_END), that player's card gets a gold ring + WIN badge. */
  winnerIdx?: number;
  /** Localised badge/announcement text for the winning card (defaults to "WIN"). */
  winnerLabel?: string;
}

/**
 * Shared trick-display area for trick-taking games (Hearts, Spades, Euchre, Bridge,
 * Napoleon, OhHell, TwoTenJack, Whist). Renders each played card with the player's
 * name underneath, or nothing when the trick is empty. AnimatedCard plays the
 * default deal SFX itself, so callers no longer need to thread a sound callback.
 *
 * Partner-aware games (Bridge, Euchre, Whist, Tarneeb, ...) pass a `team` field
 * on each TrickDisplayPlayer and the card is bordered + labeled in the human
 * team's color (blue) or the opposing team's color (red) to make ally/opponent
 * relationships scannable at a glance.
 */
export function TrickDisplay({
  currentTrick,
  players,
  cardWidth,
  label,
  dataTutorial,
  winnerIdx,
  winnerLabel,
}: TrickDisplayProps) {
  if (currentTrick.length === 0) {
    return null;
  }

  const human = players.find((p) => p.isHuman);
  const humanTeam = human?.team;
  const hasTeams = humanTeam !== undefined && players.some((p) => p.team !== undefined && p.team !== humanTeam);

  return (
    <div className="my-3 p-3 rounded bg-black/40" data-tutorial={dataTutorial}>
      <div className="text-ds-text-muted text-sm mb-1">{label}</div>
      <div className="flex gap-2">
        {currentTrick.map((trickCard) => {
          const player = players[trickCard.playerIdx];
          const team = player?.team;
          const isAlly = hasTeams && team !== undefined && team === humanTeam;
          const isFoe = hasTeams && team !== undefined && team !== humanTeam;
          const isWinner = winnerIdx !== undefined && winnerIdx === trickCard.playerIdx;
          // The winning card's gold ring takes visual priority over the ally/foe team rings.
          const wrapperClass = isWinner
            ? 'ring-2 ring-ds-warning rounded motion-safe:animate-pulse'
            : isAlly
              ? 'ring-2 ring-ds-info rounded'
              : isFoe
                ? 'ring-2 ring-ds-error rounded'
                : '';
          const labelClass = isAlly
            ? 'text-ds-info font-semibold'
            : isFoe
              ? 'text-ds-error font-semibold'
              : 'text-game-text-muted';
          return (
            <div
              key={`trick-${trickCard.playerIdx}`}
              className="relative text-center"
              data-team={team ?? undefined}
              data-team-role={isAlly ? 'ally' : isFoe ? 'foe' : undefined}
              data-trick-winner={isWinner || undefined}
            >
              <AnimatedCard card={trickCard.card} width={cardWidth} wrapperClassName={wrapperClass || undefined} />
              {isWinner && (
                <span
                  data-testid="trick-winner-badge"
                  className="absolute top-0 right-0 px-1 rounded bg-ds-warning text-ds-text-on-accent text-[8px] font-extrabold tracking-wider shadow-md pointer-events-none"
                >
                  {winnerLabel ?? 'WIN'}
                </span>
              )}
              <div className={`text-xs mt-1 ${labelClass}`}>
                {playerName(player?.id ?? trickCard.playerIdx, player?.isHuman ?? false)}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
