import { useMemo } from 'react';
import { videopokerApi } from '../api/gameApi';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { VideoPokerGameContent } from '../components/VideoPokerGameContent';
import { VIDEO_POKER_TUTORIAL_STEPS } from '../constants/videoPokerTutorial';
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
    <TutorialWrapper gameName="videopoker" steps={VIDEO_POKER_TUTORIAL_STEPS}>
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
