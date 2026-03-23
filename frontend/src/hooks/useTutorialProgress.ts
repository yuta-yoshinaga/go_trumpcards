import { gameRoutes } from '../constants/gameRoutes';
import { TUTORIAL_COMPLETED_PREFIX } from '../constants/tutorialKeys';

/** Progress state for a single game's tutorial. */
export interface GameProgress {
  /** Internal game name derived from route path. */
  gameName: string;
  /** Route path for navigation. */
  path: string;
  /** i18n key for the game label. */
  labelKey: string;
  /** Whether the tutorial has been completed. */
  completed: boolean;
}

/** Converts a route path to a game name. */
function pathToGameName(path: string): string {
  return path === '/' ? 'blackjack' : path.slice(1);
}

/** Aggregates tutorial completion state for all games from localStorage. Re-evaluated on each render. */
export function useTutorialProgress() {
  const games: GameProgress[] = gameRoutes.map((route) => {
    const gameName = pathToGameName(route.path);
    return {
      gameName,
      path: route.path,
      labelKey: route.labelKey,
      completed: localStorage.getItem(`${TUTORIAL_COMPLETED_PREFIX}${gameName}`) === 'true',
    };
  });

  const completedCount = games.filter((g) => g.completed).length;
  const totalCount = games.length;

  return { games, completedCount, totalCount };
}
