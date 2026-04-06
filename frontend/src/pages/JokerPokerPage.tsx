import { useMemo } from 'react';
import { jokerpokerApi } from '../api/gameApi';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { VideoPokerGameContent } from '../components/VideoPokerGameContent';
import type { TutorialStep } from '../types/tutorial';
import { JOKERPOKER_HELP, parseJokerpokerCommand } from '../utils/cli/commands/jokerpokerCommands';
import { formatJokerpokerState } from '../utils/cli/formatters/jokerpokerFormatter';

/** Joker Poker payout table rows. */
const JP_PAYOUT_ROWS = [
  'naturalRoyalFlush5',
  'naturalRoyalFlush',
  'fiveOfAKind',
  'wildRoyalFlush',
  'straightFlush',
  'fourOfAKind',
  'fullHouse',
  'flush',
  'straight',
  'threeOfAKind',
  'twoPair',
  'kingsOrBetter',
];

/** Joker Poker tutorial step definitions. */
const JP_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="vp-bet-controls"]',
    messageKey: 'tutorial.betControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="vp-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="vp-draw-button"]',
    messageKey: 'tutorial.drawButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="vp-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Joker Poker (Kings or Better) game page. */
export function JokerPokerPage() {
  const cliGameConfig = useMemo(
    () => ({
      parseCommand: parseJokerpokerCommand,
      formatResponse: formatJokerpokerState,
      helpText: JOKERPOKER_HELP,
    }),
    [],
  );
  return (
    <TutorialWrapper gameName="jokerpoker" steps={JP_TUTORIAL_STEPS}>
      <VideoPokerGameContent
        gameName="jokerpoker"
        i18nNamespace="jokerpoker"
        apiExec={jokerpokerApi.exec}
        payoutTableRows={JP_PAYOUT_ROWS}
        gamePath="/jokerpoker"
        cliGameConfig={cliGameConfig}
      />
    </TutorialWrapper>
  );
}
