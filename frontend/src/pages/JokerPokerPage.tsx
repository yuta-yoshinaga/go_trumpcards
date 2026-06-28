import { jokerpokerApi } from '../api/gameApi';
import { withTutorial } from '../components/tutorial/withTutorial';
import { VideoPokerGameContent } from '../components/VideoPokerGameContent';
import { VIDEO_POKER_TUTORIAL_STEPS } from '../constants/videoPokerTutorial';
import { JOKERPOKER_HELP, parseJokerpokerCommand } from '../utils/cli/commands/jokerpokerCommands';
import { formatJokerpokerState } from '../utils/cli/formatters/jokerpokerFormatter';

const CLI_GAME_CONFIG = {
  parseCommand: parseJokerpokerCommand,
  formatResponse: formatJokerpokerState,
  helpText: JOKERPOKER_HELP,
} as const;

function JokerPokerPageContent() {
  return (
    <VideoPokerGameContent
      gameName="jokerpoker"
      i18nNamespace="jokerpoker"
      apiExec={jokerpokerApi.exec}
      gamePath="/jokerpoker"
      cliGameConfig={CLI_GAME_CONFIG}
    />
  );
}

/** Renders the Joker Poker (Kings or Better) game page. */
export const JokerPokerPage = withTutorial(JokerPokerPageContent, 'jokerpoker', VIDEO_POKER_TUTORIAL_STEPS);
