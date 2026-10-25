import type { TFunction } from 'i18next';
import { useCallback, useState } from 'react';
import type { oldmaidApi } from '../api/gameApi';
import { CpuAccordion } from '../components/CpuAccordion';
import { ReplaySpeedSettingsPanel } from '../components/common/ReplaySpeedSettingsPanel';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { OldMaidDiscardedArea } from '../components/oldmaid/OldMaidDiscardedArea';
import { OldMaidDrawHistory } from '../components/oldmaid/OldMaidDrawHistory';
import { OldMaidPlayerArea } from '../components/oldmaid/OldMaidPlayerArea';
import { OldMaidSettingsDialog } from '../components/oldmaid/OldMaidSettingsDialog';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { OldMaidMode, useOldMaidGame } from '../hooks/useOldMaidGame';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CpuAction, OldMaidResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { cardLabel } from '../utils/cardUtils';
import { OLDMAID_HELP, parseOldmaidCommand } from '../utils/cli/commands/oldmaidCommands';
import { formatOldmaidState } from '../utils/cli/formatters/oldmaidFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import { findPlayerName } from '../utils/playerUtils';

/** Old Maid tutorial step definitions. */
const OM_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="om-cpu-area"]',
    messageKey: 'tutorial.cpuArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="om-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="om-draw-button"]',
    messageKey: 'tutorial.drawButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="om-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

function useGame() {
  const game = useOldMaidGame();
  return {
    ...game,
    state: game.displayState,
    exec: game.gameExec,
  };
}

type OldMaidGame = ReturnType<typeof useGame>;
type OldMaidContext = GamePageContext<OldMaidResponse, OldMaidGame, OldMaidView>;

function useOldMaidView({ game, t }: { game: OldMaidGame; t: TFunction; tc: TFunction }) {
  const { cardWidth, isMobile } = useCardDimensions();
  const [settingsOpen, setSettingsOpen] = useState(false);

  const phaseNames = {
    play: t('phase.play'),
    end: t('phase.end'),
  };

  const syncSetupFromSettings = useCallback(() => {
    if (game.gameSettings) {
      game.setSetupMode(game.gameSettings.mode);
      game.setSetupStrategy(game.gameSettings.cpuPlacementStrategy);
      game.setSetupMemoryAI(game.gameSettings.cpuMemoryAI);
      game.setSetupHesitation(game.gameSettings.cpuHesitationEnabled);
      game.setSetupMetaAI(game.gameSettings.cpuMetaAI);
    }
  }, [game]);

  return {
    cardWidth,
    isMobile,
    settingsOpen,
    setSettingsOpen,
    phaseNames,
    syncSetupFromSettings,
  };
}

type OldMaidView = ReturnType<typeof useOldMaidView>;

function AboveContent({ state, t }: OldMaidContext) {
  return (
    <>
      {/* **CUI は oldmaid.metaAILine で「自分の癖がどれだけ読まれているか」を
          毎回出しているのに、Web は同じ値をレスポンスで受け取りながら一度も
          描画していなかった。**メタAI を ON にした意味が画面から分からない。
          enabled のときだけ出す -- OFF なら学習していないので数字に意味が無い。 */}
      {state.metaAI?.enabled && (
        <div
          className="mx-4 lg:mx-8 mb-1 text-xs text-ds-text-muted"
          role="status"
          aria-live="polite"
          data-testid="om-metaai"
        >
          {t('metaAI.status', {
            games: state.metaAI.gamesPlayed,
            rate: Math.round(state.metaAI.edgePickRate * 100),
          })}
        </div>
      )}

      <ReplaySpeedSettingsPanel />
    </>
  );
}

function Board({ state, game, t, tc, view }: OldMaidContext) {
  const { cardWidth, isMobile } = view;
  const loading = game.loading;
  const isHumanTurn = !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  const cpuPlayers = state.players.filter((p) => !p.isHuman);

  const statusLines: string[] = [];
  if (!state.gameEndFlag && state.hasDrawn) {
    const from = findPlayerName(state.players, state.lastDrawPlayerIdx);
    const target = findPlayerName(state.players, state.lastDrawFromIdx);
    let msg = state.lastDrawCard
      ? t('drewCardWithLabel', { from, target, card: cardLabel(state.lastDrawCard) })
      : t('drewCard', { from, target });
    if (state.lastDiscardedPairs > 0) msg += t('discardedPairs', { count: state.lastDiscardedPairs });
    statusLines.push(msg);
  }
  if (isHumanTurn) {
    statusLines.push(t('yourTurn', { target: findPlayerName(state.players, state.nextDrawTargetIdx) }));
  }

  return (
    <>
      {/* Mode badge */}
      {state.mode === OldMaidMode.JijiNuki && (
        <div className="text-center mb-1">
          <span className="inline-block rounded-md bg-ds-error px-2.5 py-0.5 text-sm font-bold text-white">
            {t('badge.jijiNuki')}
          </span>
        </div>
      )}

      {/* CPU row */}
      <CpuAccordion playerCount={cpuPlayers.length} dataTutorial="om-cpu-area">
        <div className="flex gap-2 flex-wrap mb-2 justify-center sm:grid sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 sm:items-start">
          {cpuPlayers.map((player) => {
            // Surface a transient bubble on the CPU that just drew so the
            // player can visually track who acted without reading the log.
            // See issue #1490.
            const isDrawer =
              !state.gameEndFlag && state.hasDrawn && state.lastDrawPlayerIdx === player.id && !player.isHuman;
            const bubble = isDrawer
              ? {
                  message:
                    state.lastDiscardedPairs > 0
                      ? t('bubble.drewAndPaired', {
                          from: findPlayerName(state.players, state.lastDrawFromIdx),
                        })
                      : t('bubble.drewFrom', {
                          from: findPlayerName(state.players, state.lastDrawFromIdx),
                        }),
                  // Prepend the monotonic drawHistory length so identical
                  // back-to-back draws (same player, same target, same
                  // card bouncing back, e.g. Joker) still re-trigger the
                  // animation. Go Fish gets this for free via turnNumber.
                  triggerKey: `${state.drawHistory?.length ?? 0}-${state.lastDrawPlayerIdx}-${state.lastDrawFromIdx}-${state.lastDrawCard?.design ?? 'x'}-${state.lastDrawCard?.value ?? 0}`,
                }
              : undefined;
            return (
              <OldMaidPlayerArea
                key={player.id}
                player={player}
                isTarget={state.nextDrawTargetIdx === player.id}
                isHumanTurn={isHumanTurn}
                gameEndFlag={state.gameEndFlag}
                loading={loading}
                isSuspect={game.suspectPins.has(player.id)}
                compactNonTarget={isMobile}
                compactCpuCards={!isMobile}
                onToggleSuspect={() =>
                  game.setSuspectPins((prev) => {
                    const next = new Set(prev);
                    if (next.has(player.id)) {
                      next.delete(player.id);
                    } else {
                      next.add(player.id);
                    }
                    return next;
                  })
                }
                onDraw={(drawIdx) => game.exec('draw', drawIdx)}
                bubble={bubble}
              />
            );
          })}
        </div>
      </CpuAccordion>

      {/* Discarded Area */}
      <OldMaidDiscardedArea cards={state.lastDiscardedCards} />

      {/* Card reveal area */}
      {state.lastDrawCard && !state.gameEndFlag && (
        <div className="flex justify-center my-2" data-testid="card-reveal-area">
          {game.revealedCard ? (
            <div className="animate-flipIn">
              <AnimatedCard card={game.revealedCard} width={cardWidth} />
            </div>
          ) : (
            <AnimatedCardBack width={cardWidth} />
          )}
        </div>
      )}

      {/* Status */}
      {statusLines.length > 0 && (
        <div className="bg-black/50 rounded-lg text-ds-text-primary py-2 px-3 my-2 whitespace-pre-line text-sm">
          {statusLines.join('\n')}
        </div>
      )}

      {/* CPU log */}
      {state.cpuActions && state.cpuActions.length > 0 && (
        <div className="bg-black/40 rounded-lg text-game-text-muted py-1.5 px-2.5 my-1.5 whitespace-pre-line text-xs max-h-[120px] overflow-y-auto">
          {[
            tc('label.cpuActions'),
            ...state.cpuActions.map((action: CpuAction) => {
              const from = findPlayerName(state.players, action.drawPlayerIdx);
              const target = findPlayerName(state.players, action.drawFromIdx);
              let msg = t('drewCard', { from, target });
              // CPU drawn card is intentionally hidden to preserve game fairness
              if (action.discardedPairs > 0) {
                msg += t('discardedPairs', { count: action.discardedPairs });
                if (action.discardedCards?.length) {
                  msg += t('discardedCardNames', {
                    cards: action.discardedCards.map(cardAlt).join(t('listSeparator')),
                  });
                }
              }
              return msg;
            }),
          ].join('\n')}
        </div>
      )}

      {/* Draw History Timeline */}
      <OldMaidDrawHistory entries={state.drawHistory ?? []} players={state.players} suspectPins={game.suspectPins} />

      {/* Result */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

      {/* JijiNuki: show removed card at game end */}
      {state.gameEndFlag && state.removedCard && (
        <div className="text-center my-2 text-ds-text-primary text-sm">
          {t('removedCard', { card: cardLabel(state.removedCard) })}
        </div>
      )}

      <OldMaidSettingsDialog
        open={view.settingsOpen}
        mode={game.setupMode}
        cpuPlacementStrategy={game.setupStrategy}
        cpuMemoryAI={game.setupMemoryAI}
        cpuHesitationEnabled={game.setupHesitation}
        cpuMetaAI={game.setupMetaAI}
        onModeChange={game.setSetupMode}
        onStrategyChange={game.setSetupStrategy}
        onMemoryAIChange={game.setSetupMemoryAI}
        onHesitationChange={game.setSetupHesitation}
        onMetaAIChange={game.setSetupMetaAI}
        onApply={() => {
          view.setSettingsOpen(false);
          game.handleStart();
        }}
        onClose={() => {
          view.syncSetupFromSettings();
          view.setSettingsOpen(false);
        }}
      />
    </>
  );
}

function FooterHand({ state, game }: OldMaidContext) {
  const isHumanTurn = !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  const humanPlayer = state.players.find((p) => p.isHuman);

  // When the human just drew a card into their own hand, locate where it landed
  // so the footer hand can briefly ring-highlight that position (issue #2984).
  // Returns -1 when the card was immediately discarded as a pair (not in hand),
  // when the drawer was not the human, or at game end.
  const drawnCard = state.lastDrawCard;
  const humanDrawnCardIdx =
    humanPlayer && drawnCard && !state.gameEndFlag && state.hasDrawn && state.lastDrawPlayerIdx === humanPlayer.id
      ? (humanPlayer.cards ?? []).findIndex((c) => c.design === drawnCard.design && c.value === drawnCard.value)
      : -1;

  if (!humanPlayer) return null;
  return (
    <div className="mb-2" data-tutorial="om-player-hand">
      <OldMaidPlayerArea
        player={humanPlayer}
        isTarget={false}
        isHumanTurn={isHumanTurn}
        gameEndFlag={state.gameEndFlag}
        loading={game.loading}
        drawnCardIdx={humanDrawnCardIdx}
        onDraw={(drawIdx) => game.exec('draw', drawIdx)}
        onReorder={game.handleReorder}
      />
    </div>
  );
}

function FooterStatus({ frontendHint, t }: OldMaidContext) {
  return <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />;
}

function FooterActions({ game, state, t, busy, view, renderResetButton }: OldMaidContext) {
  const isHumanTurn = !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  return (
    <>
      <button
        type="button"
        className={`${btnSecondary} min-w-[80px]`}
        disabled={busy}
        onClick={() => {
          view.syncSetupFromSettings();
          view.setSettingsOpen(true);
        }}
      >
        {t('button.settings')}
      </button>
      {renderResetButton('min-w-[80px]')}
      <span data-tutorial="om-draw-button">
        <button
          type="button"
          className={`${btnPrimary} min-w-[110px]`}
          disabled={busy || !isHumanTurn || state.gameEndFlag}
          onClick={() => game.exec('draw')}
          // Lowercase to mean the unmodified key; only advertised while the
          // keyboard bindings are actually active (the human's turn).
          aria-keyshortcuts={isHumanTurn ? 'd' : undefined}
        >
          {t('button.drawRandom')}
        </button>
      </span>
      <button
        type="button"
        className={`${btnSecondary} min-w-[110px]`}
        disabled={busy || state.gameEndFlag}
        onClick={() => game.exec('shuffle')}
        aria-keyshortcuts={isHumanTurn ? 's' : undefined}
      >
        {t('button.shuffle')}
      </button>
      {/* Keyboard shortcut hints, shown on the human's turn (matches the d/s
          bindings in useActionKeyboardNav above). */}
      {isHumanTurn && (
        <p className="text-center text-game-text-muted text-xs mt-2" data-testid="oldmaid-key-hints">
          {t('keyHints')}
        </p>
      )}
    </>
  );
}

const oldMaidDefinition: GameDefinition<
  OldMaidResponse,
  OldMaidGame,
  Parameters<typeof oldmaidApi.exec>,
  OldMaidView
> = {
  key: 'oldmaid',
  useGame,
  useView: useOldMaidView,
  tutorial: OM_TUTORIAL_STEPS,
  skeleton: {
    kind: 'trick-taking',
    titleBar: false,
    opponents: 3,
    opponentStyle: 'hand',
    opponentHandSize: 3,
    footerHandSize: 5,
    footerButton: 'wide',
  },
  cli: ({ frontendHint }) => ({
    parseCommand: parseOldmaidCommand,
    formatResponse: formatOldmaidState,
    helpText: OLDMAID_HELP,
    localCommand: hintLocalCommand(frontendHint.hint),
  }),
  phase: ({ state, view }) => ({
    name: state.gameEndFlag ? view.phaseNames.end : view.phaseNames.play,
    ended: !!state.gameEndFlag,
  }),
  shortcuts: ({ game }) => [
    { key: 'd', action: () => void game.exec('draw'), label: 'draw' },
    { key: 's', action: () => void game.exec('shuffle'), label: 'shuffle' },
  ],
  shortcutsEnabled: ({ state }) => !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman,
  shortcutsTestId: 'old-maid-kbd-shortcuts',
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  actionsClassName: 'text-center',
  resetButton: 'custom',
  resetTutorialId: 'om-reset-button',
  shell: ({ state, game }) => ({
    isHumanTurn: !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman,
    outerKey: game.shakeKey,
    gameThemeBg: `${gameTheme.oldmaid.bg}${game.shakeKey > 0 ? ' animate-shake' : ''}`,
  }),
  AboveContent,
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Old Maid game page with settings dialog, player areas, and draw history. */
export const OldMaidPage = createGamePage(oldMaidDefinition);
