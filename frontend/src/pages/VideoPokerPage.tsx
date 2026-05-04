import { videopokerApi } from '../api/gameApi';
import { withTutorial } from '../components/tutorial/withTutorial';
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

const CLI_GAME_CONFIG = {
  parseCommand: parseVideopokerCommand,
  formatResponse: formatVideopokerState,
  helpText: VIDEOPOKER_HELP,
} as const;

function VideoPokerPageContent() {
  return (
    <VideoPokerGameContent
      gameName="videopoker"
      i18nNamespace="videopoker"
      apiExec={videopokerApi.exec}
      payoutTableRows={JOB_PAYOUT_ROWS}
      gamePath="/videopoker"
      cliGameConfig={CLI_GAME_CONFIG}
    />
  );
}

/** Renders the Video Poker (Jacks or Better) game page. */
export const VideoPokerPage = withTutorial(VideoPokerPageContent, 'videopoker', VIDEO_POKER_TUTORIAL_STEPS);
