import { useEffect, useRef, useState } from 'react';
import { klaberjassApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import type { KlaberjassResponse } from '../types/card';
import { KlaberjassPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { KLABERJASS_HELP, parseKlaberjassCommand } from '../utils/cli/commands/klaberjassCommands';
import { formatKlaberjassState } from '../utils/cli/formatters/klaberjassFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Target-score options for the Klaberjass settings panel. */
const TARGET_OPTIONS = [301, 501, 1000];

/** Suit glyphs by design value (1=Spade … 4=Diamond). */
const SUIT_GLYPHS: Readonly<Record<number, string>> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };
const SUIT_NAME_KEYS: Readonly<Record<number, string>> = {
  1: 'common.suit.spade',
  2: 'common.suit.club',
  3: 'common.suit.heart',
  4: 'common.suit.diamond',
};

/**
 * The trump order with its point values.
 *
 * Shown permanently because it is the one thing that reliably trips people up:
 * in trumps the jack and nine jump above the ace, so a hand that looks strong
 * by ordinary reckoning may not be.
 */
const TRUMP_LADDER = [
  { label: 'J', points: 20 },
  { label: '9', points: 14 },
  { label: 'A', points: 11 },
  { label: '10', points: 10 },
  { label: 'K', points: 4 },
  { label: 'Q', points: 3 },
  { label: '8', points: 0 },
  { label: '7', points: 0 },
];

/** Klaberjass tutorial step definitions. */
const KLABERJASS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="klaberjass-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="klaberjass-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="klaberjass-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="klaberjass-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="klaberjass-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const KLABERJASS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [KlaberjassPhase.BID_TURN_UP]: 'bidTurnUp',
  [KlaberjassPhase.BID_FREE]: 'bidFree',
  [KlaberjassPhase.SCHMEISS]: 'schmeiss',
  [KlaberjassPhase.PLAY]: 'play',
  [KlaberjassPhase.HAND_END]: 'handEnd',
  [KlaberjassPhase.GAME_END]: 'gameEnd',
};

type KlaberjassGame = ReturnType<typeof useKlaberjassGame>;
type KlaberjassView = ReturnType<typeof useKlaberjassView>;
type KlaberjassContext = GamePageContext<KlaberjassResponse, KlaberjassGame, KlaberjassView>;

function useKlaberjassGame() {
  const { state, loading, error, exec, retry } = useGameApi(klaberjassApi.exec);

  const [targetScore, setTargetScore] = useState(501);
  const [allowSchmeiss, setAllowSchmeiss] = useState(true);
  const [selected, setSelected] = useState<number | null>(null);

  useMountReset(exec);

  const handleTargetChange = (value: string) => {
    const next = Number(value);
    setTargetScore(next);
    exec('reset', { config: { targetScore: next, allowSchmeiss } });
  };

  const handleAllowSchmeissChange = (checked: boolean) => {
    setAllowSchmeiss(checked);
    exec('reset', { config: { targetScore, allowSchmeiss: checked } });
  };

  const handlePlay = () => {
    if (selected === null) return;
    exec('play', { cardIndex: selected });
    setSelected(null);
  };

  const handleReset = () => {
    setSelected(null);
    exec('reset', { config: { targetScore, allowSchmeiss } });
  };

  return {
    state,
    loading,
    error,
    exec,
    retry,
    targetScore,
    allowSchmeiss,
    selected,
    setSelected,
    handlePlay,
    handleReset,
    handleTargetChange,
    handleAllowSchmeissChange,
  };
}

function useKlaberjassView({ game, t }: { game: KlaberjassGame; t: KlaberjassContext['t'] }) {
  const { state } = game;
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('klaberjass', KLABERJASS_PHASE_KEYS);

  const [settlementAnnouncement, setSettlementAnnouncement] = useState('');
  const lastSettlementKey = useRef<string | null>(null);

  const settlementKey =
    state && (state.phase === KlaberjassPhase.HAND_END || state.phase === KlaberjassPhase.GAME_END)
      ? `${state.dealNumber}-${state.phase}`
      : null;

  useEffect(() => {
    if (!state || settlementKey === null) {
      lastSettlementKey.current = null;
      setSettlementAnnouncement('');
      return;
    }
    if (lastSettlementKey.current === settlementKey) return;

    lastSettlementKey.current = settlementKey;
    const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));
    const details = state.players.map(
      (p) =>
        `${playerLabel(p.id, p.isHuman)}: ${t('pointsBreakdown', {
          card: p.cardPoints,
          sequence: p.sequencePoints,
          bela: p.belaPoints,
          lastTrick: p.lastTrickPoints,
          total: p.handPoints,
        })} ${t('scoreUpdate', { score: p.score })}`,
    );
    setSettlementAnnouncement(`${t('settlementTitle')}: ${details.join(t('listSeparator'))}`);
  }, [settlementKey, state, t]);

  return {
    cardWidth,
    isMobile,
    phaseNames,
    settlementAnnouncement,
  };
}

function LiveRegion({ view }: KlaberjassContext) {
  return (
    <div className="sr-only" role="status" aria-live="polite" data-testid="klaberjass-settlement-announcement">
      {view.settlementAnnouncement}
    </div>
  );
}

function Board({ state, t, view }: KlaberjassContext) {
  const { cardWidth, isMobile } = view;
  const isHandEnd = state.phase === KlaberjassPhase.HAND_END;
  const isGameEnd = state.phase === KlaberjassPhase.GAME_END || state.gameEndFlag;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));
  const suitGlyph = (suit: number): string => SUIT_GLYPHS[suit] ?? t('trumpNone');

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2" data-tutorial="klaberjass-info">
        <span className="mr-4">{t('deal', { n: state.dealNumber })}</span>
        <span className="mr-4">{t('target', { n: state.targetScore })}</span>
        <span className="mr-4">
          {t('trump')}: {suitGlyph(state.trumpSuit)}
        </span>
        {state.turnUpCard && state.trumpSuit === 0 && (
          <span data-testid="klaberjass-turnup">
            {t('turnUp')}: {suitGlyph(suitOf(state.turnUpCard.design))}
          </span>
        )}
      </div>

      {/* The trump ladder — the one rule people reliably get wrong. */}
      <div className="mb-2 p-2 rounded bg-black/20 text-xs" data-testid="klaberjass-ladder">
        <div className="mb-1 text-ds-text-primary">{t('ladderTitle')}</div>
        <div className="flex flex-wrap gap-x-3 gap-y-1">
          {TRUMP_LADDER.map((r) => (
            <span key={r.label} className={r.points >= 14 ? 'text-ds-warning font-semibold' : 'text-ds-text-muted'}>
              {r.label} ({r.points})
            </span>
          ))}
        </div>
        <div className="mt-1 text-ds-text-muted">{t('ladderNote')}</div>
        <div className="text-ds-text-muted">{t('sequenceNote')}</div>
      </div>

      {/* Turn-up card while bidding */}
      {state.turnUpCard && state.trumpSuit === 0 && (
        <div className="mb-2 flex items-center gap-2" data-testid="klaberjass-turnup-card">
          <span className="text-ds-text-muted text-sm">{t('turnUp')}</span>
          <CardImage card={state.turnUpCard} width={cardWidth} />
        </div>
      )}

      {/* Players */}
      <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="klaberjass-players">
        <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
        {state.players.map((p) => (
          <div
            key={`player-${p.id}`}
            data-testid="klaberjass-player"
            className={`text-sm py-0.5 flex items-center gap-2 ${
              p.isCurrentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
            } ${p.isHuman ? 'font-semibold' : ''}`}
          >
            <span>{playerLabel(p.id, p.isHuman)}</span>
            {p.isDealer && <span className="text-ds-accent">[{t('dealer')}]</span>}
            {p.isMaker && <span className="text-ds-success">[{t('maker')}]</span>}
            <span>{t('score', { n: p.score })}</span>
            <span>{t('handPoints', { n: p.handPoints })}</span>
            {!p.isHuman && p.cards.length === 0 && <span>{t('hiddenHand', { count: p.cardCount })}</span>}
          </div>
        ))}
      </div>

      <section className="mb-2 p-2 rounded bg-black/20 text-sm" data-testid="klaberjass-trick-history">
        <h2 className="mb-1 text-ds-text-primary">{t('trickHistory')}</h2>
        {state.trickHistory.map((entry, index) => (
          <div key={`trick-history-${index}`} className="text-ds-text-muted">
            {t('trickHistoryEntry', {
              number: index + 1,
              name: playerLabel(entry.winnerIdx, entry.winnerIdx === 0),
              points: entry.points,
            })}
          </div>
        ))}
      </section>

      {/* Current trick */}
      {state.trick.length > 0 && (
        <div className="mb-2 flex items-center gap-2" data-testid="klaberjass-trick">
          <span className="text-ds-text-muted text-sm">{t('currentTrick')}</span>
          {state.trick.map((c, i) => (
            <CardImage key={`trick-${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />
          ))}
        </div>
      )}

      {/* Settlement */}
      {(isHandEnd || isGameEnd) && (
        <div className={`mb-2 p-2 rounded text-sm ${badgeWarningColors}`} data-testid="klaberjass-settlement">
          <h2 className="mb-1 font-semibold">{t('settlementTitle')}</h2>
          <div>{state.bete ? t('beteLine') : t('madeLine')}</div>
          <ul>
            {state.players.map((p) => (
              <li key={p.id} data-testid="klaberjass-points-breakdown">
                {playerLabel(p.id, p.isHuman)}:{' '}
                {t('pointsBreakdown', {
                  card: p.cardPoints,
                  sequence: p.sequencePoints,
                  bela: p.belaPoints,
                  lastTrick: p.lastTrickPoints,
                  total: p.handPoints,
                })}{' '}
                {t('scoreUpdate', { score: p.score })}
              </li>
            ))}
          </ul>
          <div>
            {state.sequenceWinner >= 0
              ? t('sequenceWinner', { name: playerLabel(state.sequenceWinner, state.sequenceWinner === 0) })
              : t('sequenceNobody')}
          </div>
          {state.belaScored && state.belaHolder >= 0 && (
            <div>{t('belaLine', { name: playerLabel(state.belaHolder, state.belaHolder === 0) })}</div>
          )}
          {state.dixUsed && <div>{t('dixLine')}</div>}
          {/* **最終トリックには 10 点が付く。**書かないと、ベラや宣言点を
              足しても handPoints と合わない理由が説明できない (#4937)。 */}
          {state.lastTrickWinner >= 0 && (
            <div data-testid="klaberjass-last-trick-bonus">
              {t('lastTrickBonus', {
                name: playerLabel(state.lastTrickWinner, state.lastTrickWinner === 0),
                points: state.lastTrickBonus,
              })}
            </div>
          )}
        </div>
      )}

      {isMobile && (
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
      )}
    </>
  );
}

function FooterHand({ game, state, t, view }: KlaberjassContext) {
  const human = state.players.find((p) => p.isHuman);
  const isPlay = state.phase === KlaberjassPhase.PLAY;
  const canPlay = (i: number) => state.validPlays.includes(i);

  return (
    <div className="mb-2" data-tutorial="klaberjass-hand">
      <div className="text-ds-text-muted text-xs mb-1">{t('yourHand')}</div>
      <div className="flex flex-wrap gap-1">
        {(human?.cards ?? []).map((c, i) => {
          const playable = !isPlay || canPlay(i);
          return (
            <button
              key={`hand-${c.design}-${c.value}-${i}`}
              type="button"
              data-hint-action="play"
              // **リングは目にしか届かない。**どの札を選んでいるかを
              // 音声でも確かめられないと、出す前に検証できない (#6523)。
              aria-pressed={game.selected === i}
              onClick={() => game.setSelected(i)}
              disabled={game.loading || (isPlay && !canPlay(i))}
              className={`rounded ${game.selected === i ? 'ring-2 ring-ds-accent' : ''} ${
                playable ? '' : 'opacity-40'
              }`}
            >
              <CardImage card={c} width={view.cardWidth} />
            </button>
          );
        })}
      </div>
    </div>
  );
}

function FooterStatus({ state, t, frontendHint }: KlaberjassContext) {
  const isBidTurnUp = state.phase === KlaberjassPhase.BID_TURN_UP;
  const isBidFree = state.phase === KlaberjassPhase.BID_FREE;
  const isSchmeiss = state.phase === KlaberjassPhase.SCHMEISS;
  const isPlay = state.phase === KlaberjassPhase.PLAY;
  const isGameEnd = state.phase === KlaberjassPhase.GAME_END || state.gameEndFlag;
  const isHumanBid = (isBidTurnUp || isBidFree) && state.bidPlayerIdx === 0 && !isGameEnd;
  const isHumanSchmeissAnswer = isSchmeiss && state.bidPlayerIdx === 0 && !isGameEnd;
  const isHumanPlay = isPlay && state.currentPlayerIdx === 0 && !isGameEnd;

  return (
    <>
      {isHumanBid && (
        <div className="text-ds-text-muted text-xs mb-2" data-testid="klaberjass-bid-notice">
          {isBidTurnUp ? t('bidNotice') : t('bidFreeNotice')}
        </div>
      )}
      {isHumanSchmeissAnswer && (
        <div className="text-ds-text-muted text-xs mb-2" data-testid="klaberjass-schmeiss-notice">
          {t('schmeissNotice')}
        </div>
      )}
      {isHumanPlay && (
        <div className="text-ds-text-muted text-xs mb-2" data-testid="klaberjass-play-notice">
          {t('playNotice')}
        </div>
      )}

      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, t, tc }: KlaberjassContext) {
  const isBidTurnUp = state.phase === KlaberjassPhase.BID_TURN_UP;
  const isBidFree = state.phase === KlaberjassPhase.BID_FREE;
  const isSchmeiss = state.phase === KlaberjassPhase.SCHMEISS;
  const isPlay = state.phase === KlaberjassPhase.PLAY;
  const isHandEnd = state.phase === KlaberjassPhase.HAND_END;
  const isGameEnd = state.phase === KlaberjassPhase.GAME_END || state.gameEndFlag;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  const isHumanBid = (isBidTurnUp || isBidFree) && state.bidPlayerIdx === 0 && !isGameEnd;
  const isHumanSchmeissAnswer = isSchmeiss && state.bidPlayerIdx === 0 && !isGameEnd;
  const isHumanPlay = isPlay && state.currentPlayerIdx === 0 && !isGameEnd;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  return (
    <>
      {isHumanBid && isBidTurnUp && (
        <button type="button" className={btnSuccess} onClick={() => void game.exec('accept')} disabled={game.loading}>
          {t('acceptButton')}
        </button>
      )}

      {isHumanBid &&
        isBidFree &&
        [1, 2, 3, 4]
          // **断られた表向きのスートは選べない。**第1ラウンドで流れている。
          .filter((s) => s !== (state.turnUpCard ? suitOf(state.turnUpCard.design) : 0))
          .map((s) => (
            <button
              key={`call-${s}`}
              type="button"
              className={btnPrimary}
              onClick={() => void game.exec('call', { suit: s })}
              disabled={game.loading}
            >
              {t('callButton', { suit: SUIT_GLYPHS[s], suitName: tc(SUIT_NAME_KEYS[s]) })}
            </button>
          ))}

      {isHumanBid && (
        <>
          <button type="button" className={btnPrimary} onClick={() => void game.exec('pass')} disabled={game.loading}>
            {t('passButton')}
          </button>
          <button
            type="button"
            className={btnWarning}
            onClick={() => void game.exec('schmeiss')}
            disabled={game.loading}
          >
            {t('schmeissButton')}
          </button>
        </>
      )}

      {isHumanSchmeissAnswer && (
        <>
          <button
            type="button"
            className={btnSuccess}
            onClick={() => void game.exec('answerschmeiss', { accept: true })}
            disabled={game.loading}
          >
            {t('schmeissAcceptButton')}
          </button>
          <button
            type="button"
            className={btnWarning}
            onClick={() => void game.exec('answerschmeiss', { accept: false })}
            disabled={game.loading}
          >
            {t('schmeissRefuseButton')}
          </button>
        </>
      )}

      {isHumanPlay && (
        <button
          type="button"
          className={btnSuccess}
          onClick={game.handlePlay}
          disabled={game.loading || game.selected === null}
        >
          {t('playButton')}
        </button>
      )}

      {isHandEnd && !isGameEnd && (
        <button type="button" className={btnSuccess} onClick={() => void game.exec('next')} disabled={game.loading}>
          {t('nextDeal')}
        </button>
      )}

      {isGameEnd && (
        <span className="text-ds-text-primary text-sm font-semibold mr-1">
          {humanWon ? t('win') : t('lose', { name: playerLabel(state.winnerIdx, state.winnerIdx === 0) })}
        </span>
      )}
    </>
  );
}

const klaberjassDefinition: GameDefinition<
  KlaberjassResponse,
  KlaberjassGame,
  Parameters<typeof klaberjassApi.exec>,
  KlaberjassView
> = {
  key: 'klaberjass',
  useGame: useKlaberjassGame,
  useView: useKlaberjassView,
  tutorial: KLABERJASS_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 9 },
  cli: ({ frontendHint }) => ({
    parseCommand: parseKlaberjassCommand,
    formatResponse: formatKlaberjassState,
    helpText: KLABERJASS_HELP,
    localCommand: hintLocalCommand(frontendHint.hint),
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === KlaberjassPhase.GAME_END || state.gameEndFlag;
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerIdx === 0,
    };
  },
  shell: ({ state }) => {
    const isBidTurnUp = state.phase === KlaberjassPhase.BID_TURN_UP;
    const isBidFree = state.phase === KlaberjassPhase.BID_FREE;
    const isSchmeiss = state.phase === KlaberjassPhase.SCHMEISS;
    const isPlay = state.phase === KlaberjassPhase.PLAY;
    const isGameEnd = state.phase === KlaberjassPhase.GAME_END || state.gameEndFlag;
    const isHumanBid = (isBidTurnUp || isBidFree) && state.bidPlayerIdx === 0 && !isGameEnd;
    // 投げの提案に答えるのは、提案した「相手」ではないほう。
    const isHumanSchmeissAnswer = isSchmeiss && state.bidPlayerIdx === 0 && !isGameEnd;
    const isHumanPlay = isPlay && state.currentPlayerIdx === 0 && !isGameEnd;
    return {
      isHumanTurn: isHumanBid || isHumanSchmeissAnswer || isHumanPlay,
    };
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'targetScore',
          label: t('settings.targetScore'),
          value: game.targetScore,
          options: TARGET_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: game.handleTargetChange,
        },
        {
          type: 'checkbox',
          id: 'allowSchmeiss',
          label: t('settings.allowSchmeiss'),
          checked: game.allowSchmeiss,
          onToggle: game.handleAllowSchmeissChange,
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  message: ({ state, view }) =>
    !view.isMobile ? (
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    ) : null,
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'klaberjass-actions',
  resetTutorialId: 'klaberjass-reset-button',
  LiveRegion,
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Klaberjass game page: the two-player ancestor of the Jass family. */
export const KlaberjassPage = createGamePage(klaberjassDefinition);

/** Maps a wire design string back to the domain's 1-based suit value. */
function suitOf(design: string): number {
  switch (design) {
    case 'SPADE':
      return 1;
    case 'CLOVER':
      return 2;
    case 'HEART':
      return 3;
    case 'DIAMOND':
      return 4;
    default:
      return 0;
  }
}
