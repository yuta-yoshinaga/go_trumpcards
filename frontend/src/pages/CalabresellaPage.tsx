import { useEffect } from 'react';
import type { calabresellaApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useCalabresellaGame } from '../hooks/useCalabresellaGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { CalabresellaResponse } from '../types/card';
import { CalabresellaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { CALABRESELLA_HELP, parseCalabresellaCommand } from '../utils/cli/commands/calabresellaCommands';
import { formatCalabresellaState } from '../utils/cli/formatters/calabresellaFormatter';
import { formatSignedDelta } from '../utils/formatSignedDelta';
import { isRequestedHint } from '../utils/hintRequest';
import { findPlayerName, playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Calabresella tutorial step definitions. */
const CALABRESELLA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="calabresella-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="calabresella-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="calabresella-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="calabresella-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="calabresella-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const CALABRESELLA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CalabresellaPhase.BID]: 'bid',
  [CalabresellaPhase.DISCARD]: 'discard',
  [CalabresellaPhase.PLAY]: 'play',
  [CalabresellaPhase.TRICK_END]: 'trickEnd',
  [CalabresellaPhase.ROUND_END]: 'roundEnd',
  [CalabresellaPhase.GAME_END]: 'gameEnd',
};

/** Bid labels indexed by bid value (0=pass/none, 1=chiamo, 2=solo). */
const BID_KEYS = ['bidNone', 'bidChiamo', 'bidSolo'] as const;

/** Shared suit translation keys indexed by card design. */
const SUIT_I18N_KEYS: Readonly<Record<'SPADE' | 'CLOVER' | 'HEART' | 'DIAMOND', string>> = {
  SPADE: 'common.suit.spade',
  CLOVER: 'common.suit.club',
  HEART: 'common.suit.heart',
  DIAMOND: 'common.suit.diamond',
};

type CalabresellaGame = ReturnType<typeof useCalabresellaGame>;
type CalabresellaContext = GamePageContext<CalabresellaResponse, CalabresellaGame, CalabresellaView>;

function useCalabresellaView({ game }: { game: CalabresellaGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('calabresella', CALABRESELLA_PHASE_KEYS);
  return { cardWidth, isMobile, phaseNames };
}

type CalabresellaView = ReturnType<typeof useCalabresellaView>;

/** Values derived from the state that the board, footer hand and footer actions all read. */
function derive({ game, state }: Pick<CalabresellaContext, 'game' | 'state'>) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isBidPhase = state.phase === CalabresellaPhase.BID;
  const isDiscardPhase = state.phase === CalabresellaPhase.DISCARD;
  const isPlayPhase = state.phase === CalabresellaPhase.PLAY;
  const isTrickEnd = state.phase === CalabresellaPhase.TRICK_END;
  const isRoundEnd = state.phase === CalabresellaPhase.ROUND_END;
  const isGameEnd = state.phase === CalabresellaPhase.GAME_END || state.gameEndFlag;

  const canBid = isBidPhase && state.currentBidderIdx === humanIdx;
  const canSubmitBid = (bid: number) => bid > state.highestBid;
  const submitBid = (bid: number) => {
    if (bid === 0 || canSubmitBid(bid)) game.handleBid(bid);
  };
  const canDiscard = isDiscardPhase && state.soloistIdx === humanIdx;
  const canPlay = isPlayPhase && isHumanTurn;
  // The soloist takes the 4-card monte (16 cards) and discards down to the regulation
  // 12-card hand (CalabresellaHandSize in the backend).
  const REGULATION_HAND_SIZE = 12;
  const discardRemaining = humanPlayer ? Math.max(0, humanPlayer.cards.length - REGULATION_HAND_SIZE) : 0;
  const discardSelectionRemaining = Math.max(0, discardRemaining - game.selectedCardIndices.length);

  return {
    humanPlayer,
    humanIdx,
    isBidPhase,
    isTrickEnd,
    isRoundEnd,
    isGameEnd,
    canBid,
    canSubmitBid,
    submitBid,
    canDiscard,
    canPlay,
    discardRemaining,
    discardSelectionRemaining,
  };
}

function Board({ state, t, tc, view, game }: CalabresellaContext) {
  const { cardWidth, isMobile } = view;
  const { isTrickEnd, isRoundEnd, isGameEnd } = derive({ game, state });
  const firstTrickCard = state.currentTrick[0]?.card;
  const leadSuit = firstTrickCard
    ? tc(SUIT_I18N_KEYS[firstTrickCard.design as keyof typeof SUIT_I18N_KEYS])
    : undefined;

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span className="mr-4">{t('winningBid', { bid: t(BID_KEYS[state.winningBid] ?? 'bidNone') })}</span>
        <span>{t('target', { points: state.config.targetPoints })}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          {/* Monte (widow): the 4 cards the Soloist took, revealed to all players. */}
          {state.monte && state.monte.length > 0 && (
            <div
              className="mb-3 p-2 rounded bg-black/30"
              data-testid="calabresella-monte"
              data-tutorial="calabresella-monte"
            >
              <div className="text-ds-text-muted text-sm mb-1">{t('monteLabel')}</div>
              <div className="flex flex-wrap gap-1">
                {state.monte.map((c, i) => (
                  <CardImage key={`monte-${c.design}-${c.value}-${i}`} card={c} width={Math.round(cardWidth * 0.75)} />
                ))}
              </div>
            </div>
          )}

          <TrickDisplay
            currentTrick={state.currentTrick}
            leadSuit={leadSuit}
            leadSuitLabel={t('leadSuit')}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            winnerIdx={isTrickEnd ? state.lastTrickWinner : undefined}
            winnerLabel={
              isTrickEnd
                ? t('previousTrickWinner', { name: findPlayerName(state.players, state.lastTrickWinner) })
                : undefined
            }
            dataTutorial="calabresella-trick-display"
          />
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="calabresella-info">
          {/* Per-player match scores with Soloist badge */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            {state.players.map((p) => (
              <div key={p.id} className="py-0.5 flex items-center gap-2">
                <span className={p.isSoloist ? 'text-ds-warning font-semibold' : ''}>
                  {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                </span>
                {p.isSoloist && (
                  <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('soloistBadge')}</span>
                )}
              </div>
            ))}
          </div>

          {/* Players: cards / tricks */}
          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
              <div className="mt-1">
                {state.players.map((p) => (
                  <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                    {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                    {t('tricks', { count: p.trickCount })}
                  </div>
                ))}
              </div>
            </details>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players.map((p) => (
                <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                  {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                  {t('tricks', { count: p.trickCount })}
                </div>
              ))}
            </div>
          )}

          {/* Round result: per-player thirds captured */}
          {(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              {state.players.map((p) => (
                <div key={p.id}>
                  {t('roundResult.thirds', {
                    name: playerName(p.id, p.isHuman),
                    thirds: state.roundThirds[p.id] ?? 0,
                  })}
                </div>
              ))}
              <div>{t(state.soloistWon ? 'roundResult.soloistWon' : 'roundResult.soloistLost')}</div>
              {state.players.map((p) => {
                const change = state.roundScoreChanges[p.id];
                const signedChange = formatSignedDelta(change);
                return (
                  <div key={`settlement-${p.id}`}>
                    {t('roundResult.settlement', {
                      name: playerName(p.id, p.isHuman),
                      change: signedChange,
                    })}
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ game, state, t, view }: CalabresellaContext) {
  const { humanPlayer, isBidPhase, canDiscard, canPlay, discardRemaining, discardSelectionRemaining } = derive({
    game,
    state,
  });
  const { selectedCardIndices, toggleCard } = game;
  return (
    <>
      {/* 領域は**常設**。中身だけ差し替える ── 出現と同時に付けた領域は
          変化として扱われず読み上げられない (#5955、hint-live と同じ形)。 */}
      <div data-testid="calabresella-prompt-live" role="status" aria-live="polite">
        {isBidPhase && (
          <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="calabresella-bid-prompt">
            {t('bidPhase')}
          </div>
        )}
        {canDiscard && discardRemaining > 0 && (
          <div
            className="mb-1 text-center text-sm text-ds-accent font-semibold"
            data-testid="calabresella-discard-prompt"
          >
            {t('discardPhaseRemaining', { count: discardSelectionRemaining })}
          </div>
        )}
      </div>
      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={selectedCardIndices}
          toggleCard={(index) => {
            if (canDiscard && !selectedCardIndices.includes(index) && selectedCardIndices.length >= discardRemaining)
              return;
            toggleCard(index);
          }}
          cardWidth={view.cardWidth}
          isMobile={view.isMobile}
          dataTutorialPrefix="calabresella"
          validIndices={canPlay ? state.playableIndices : undefined}
          restrictedTooltip={t('playButton')}
        />
      )}
    </>
  );
}

function FooterStatus({ state, t, frontendHint }: CalabresellaContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="calabresella-hint-live" role="status" aria-live="polite">
        {state.hint && isRequestedHint(state) && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
            {state.hint.cardIndices &&
              state.hint.cardIndices.length > 0 &&
              ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, t }: CalabresellaContext) {
  const { loading, selectedCardIndices, handleDiscard, handlePlay, handleNextTrick, handleNextRound } = game;
  const { canBid, canSubmitBid, submitBid, canDiscard, canPlay, isTrickEnd, isRoundEnd, discardRemaining } = derive({
    game,
    state,
  });
  return (
    <>
      {canBid && (
        <>
          <span id="calabresella-bid-unavailable-reason" className="sr-only">
            {t('bidUnavailable')}
          </span>
          <button
            type="button"
            className={btnPrimary}
            onClick={() => submitBid(1)}
            disabled={loading}
            aria-disabled={!canSubmitBid(1)}
            aria-describedby={!canSubmitBid(1) ? 'calabresella-bid-unavailable-reason' : undefined}
          >
            {t('bidChiamo')}
          </button>
          <button
            type="button"
            className={btnPrimary}
            onClick={() => submitBid(2)}
            disabled={loading}
            aria-disabled={!canSubmitBid(2)}
            aria-describedby={!canSubmitBid(2) ? 'calabresella-bid-unavailable-reason' : undefined}
          >
            {t('bidSolo')}
          </button>
          <button type="button" className={btnSecondary} onClick={() => submitBid(0)} disabled={loading}>
            {t('bidPass')}
          </button>
        </>
      )}
      {canDiscard && (
        <button
          type="button"
          className={btnPrimary}
          onClick={handleDiscard}
          disabled={loading || selectedCardIndices.length !== discardRemaining}
          data-testid="calabresella-discard-button"
        >
          {discardRemaining > 0 ? t('discardCards', { count: discardRemaining }) : t('discardCard')}
        </button>
      )}
      {canPlay && (
        <button
          type="button"
          className={btnPrimary}
          onClick={handlePlay}
          disabled={loading || selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}
      {isTrickEnd && (
        <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
          {t('nextTrick')}
        </button>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
          {t('nextRound')}
        </button>
      )}
    </>
  );
}

const calabresellaDefinition: GameDefinition<
  CalabresellaResponse,
  CalabresellaGame,
  Parameters<typeof calabresellaApi.exec>,
  CalabresellaView
> = {
  key: 'calabresella',
  useGame: useCalabresellaGame,
  useView: useCalabresellaView,
  tutorial: CALABRESELLA_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 12 },
  cli: () => ({
    parseCommand: parseCalabresellaCommand,
    formatResponse: formatCalabresellaState,
    helpText: CALABRESELLA_HELP,
  }),
  phase: ({ state, view, game }) => {
    const { isGameEnd, humanIdx } = derive({ game, state });
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerPlayer === humanIdx,
    };
  },
  shell: ({ state }) => ({ isHumanTurn: state.isHumanTurn }),
  reset: ({ game }) => {
    game.reset();
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.calabresellaConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetPoints',
          label: t('settings.targetPoints'),
          value: game.calabresellaConfig.targetPoints,
          options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetPoints', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'calabresella-action-buttons',
  resetTutorialId: 'calabresella-reset-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Calabresella (Terziglio) game page: a Calabrian/Italian 3-player 40-card Tressette-family trick-taker with bidding and a monte exchange. */
export const CalabresellaPage = createGamePage(calabresellaDefinition);
