import { withTutorial } from '../components/tutorial/withTutorial';
import type { TutorialStep } from '../types/tutorial';
import { HorsePageContent } from './HorsePage';

/** Eight-Game Mix tutorial step definitions. */
const EIGHT_GAME_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ho-discipline"]',
    messageKey: 'tutorial.discipline',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ho-seats"]', messageKey: 'tutorial.seats', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ho-actions"]', messageKey: 'tutorial.actions', placement: 'top', advanceOn: 'next' },
];

/**
 * Renders the Eight-Game Mix page: eight poker disciplines rotating at one
 * table.
 *
 * The board is the H.O.R.S.E. board — same orchestrator, one rotation longer —
 * so this page supplies the game key and its own tutorial, and the shared
 * content component does the rest.
 */
export const EightGamePage = withTutorial(
  () => <HorsePageContent gameKey="eightgame" />,
  'eightgame',
  EIGHT_GAME_TUTORIAL_STEPS,
);
