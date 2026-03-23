import { useCallback, useState } from 'react';
import { GAME_VISITED_PREFIX, TUTORIAL_COMPLETED_PREFIX, TUTORIAL_NO_SUGGEST_KEY } from '../constants/tutorialKeys';

/** Manages first-visit detection for a game, controlling whether to show a tutorial suggestion dialog. */
export function useFirstVisit(gameName: string) {
  const visitedKey = `${GAME_VISITED_PREFIX}${gameName}`;
  const completedKey = `${TUTORIAL_COMPLETED_PREFIX}${gameName}`;

  const [shouldShowDialog, setShouldShowDialog] = useState(() => {
    if (localStorage.getItem(visitedKey) === 'true') return false;
    if (localStorage.getItem(completedKey) === 'true') return false;
    if (localStorage.getItem(TUTORIAL_NO_SUGGEST_KEY) === 'true') return false;
    return true;
  });

  /** Mark the game as visited, hiding the dialog for this game. */
  const dismiss = useCallback(() => {
    localStorage.setItem(visitedKey, 'true');
    setShouldShowDialog(false);
  }, [visitedKey]);

  /** Mark the game as visited and suppress the dialog for all future games. */
  const dismissPermanently = useCallback(() => {
    localStorage.setItem(visitedKey, 'true');
    localStorage.setItem(TUTORIAL_NO_SUGGEST_KEY, 'true');
    setShouldShowDialog(false);
  }, [visitedKey]);

  return { shouldShowDialog, dismiss, dismissPermanently };
}
