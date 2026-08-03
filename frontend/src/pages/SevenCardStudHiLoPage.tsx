import { withTutorial } from '../components/tutorial/withTutorial';
import type { TutorialStep } from '../types/tutorial';
import { SevenCardStudPageContent } from './SevenCardStudPage';

/** Seven Card Stud Hi-Lo tutorial steps.
 *
 * The deal and the streets are identical to plain stud, so the steps that
 * differ are the ones about the split: what makes a low, and what happens when
 * nobody makes one. */
const SCSHL_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="scs-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="scs-player-hand"]', messageKey: 'tutorial.low', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="scs-pot-display"]', messageKey: 'tutorial.split', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="scs-action-buttons"]', messageKey: 'tutorial.scoop', placement: 'top', advanceOn: 'next' },
];

/**
 * Renders the Seven Card Stud Hi-Lo (8 or Better) page.
 *
 * The whole page is shared with plain Seven Card Stud — the deal, the streets
 * and the betting are the same game, and only the showdown splits. A second
 * ~600-line copy would be duplication, not a feature, so this passes a game key
 * and the shared content renders the Hi/Lo breakdown when the server sets
 * `isHiLo`.
 */
export const SevenCardStudHiLoPage = withTutorial(
  () => <SevenCardStudPageContent gameKey="sevencardstudhilo" />,
  'sevencardstudhilo',
  SCSHL_TUTORIAL_STEPS,
);
