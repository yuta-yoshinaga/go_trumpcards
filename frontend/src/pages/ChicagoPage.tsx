import { withTutorial } from '../components/tutorial/withTutorial';
import type { TutorialStep } from '../types/tutorial';
import { SevenCardStudPageContent } from './SevenCardStudPage';

/** Chicago tutorial steps.
 *
 * The deal, the streets and the high hand are identical to plain stud, so the
 * steps that differ are the ones about the other half: which card decides it,
 * that it must be face-down, and what happens when nobody holds one. */
const CHICAGO_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="scs-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="scs-player-hand"]', messageKey: 'tutorial.spade', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="scs-pot-display"]', messageKey: 'tutorial.split', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="scs-action-buttons"]', messageKey: 'tutorial.scoop', placement: 'top', advanceOn: 'next' },
];

/**
 * Renders the Chicago page.
 *
 * The whole page is shared with plain Seven Card Stud — the deal, the streets
 * and the betting are the same game, and only the showdown splits. A second
 * ~600-line copy would be duplication, not a feature, so this passes a game key
 * and the shared content renders the high/spade breakdown when the server sets
 * `isChicago`.
 */
export const ChicagoPage = withTutorial(
  () => <SevenCardStudPageContent gameKey="chicago" />,
  'chicago',
  CHICAGO_TUTORIAL_STEPS,
);
