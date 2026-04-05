import { useMemo } from 'react';
import { videopokerApi } from '../api/gameApi';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { VideoPokerGameContent } from '../components/VideoPokerGameContent';
import type { TutorialStep } from '../types/tutorial';
import { parseVideopokerCommand, VIDEOPOKER_HELP } from '../utils/cli/commands/videopokerCommands';
import { formatVideopokerState } from '../utils/cli/formatters/videopokerFormatter';

/** Jacks or Better payout table rows. */
const JOB_PAYOUT_ROWS = [
  'royalFlush5',
  'royalFlush',
  'straightFlush',
  'fourOfAKind',
  'fullHouse',
  'flush',
  'straight',
  'threeOfAKind',
  'twoPair',
  'jacksOrBetter',
];

/** Video Poker tutorial step definitions. */
const VP_TUTORIAL_STEPS: TutorialStep[] = [
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

/** Renders the Video Poker (Jacks or Better) game page. */
export function VideoPokerPage() {
  const cliGameConfig = useMemo(
    () => ({
      parseCommand: parseVideopokerCommand,
      formatResponse: formatVideopokerState,
      helpText: VIDEOPOKER_HELP,
    }),
    [],
  );
  return (
    <TutorialWrapper gameName="videopoker" steps={VP_TUTORIAL_STEPS}>
      <VideoPokerGameContent
        gameName="videopoker"
        i18nNamespace="videopoker"
        apiExec={videopokerApi.exec}
        payoutTableRows={JOB_PAYOUT_ROWS}
        gamePath="/videopoker"
        cliGameConfig={cliGameConfig}
      />
    </TutorialWrapper>
  );
}
