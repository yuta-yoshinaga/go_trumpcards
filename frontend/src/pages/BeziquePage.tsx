import { useEffect, useRef, useState } from 'react';
import type { beziqueApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { LiveAnnouncement } from '../components/LiveAnnouncement';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { TARGET_SCORE_OPTIONS, useBeziqueGame } from '../hooks/useBeziqueGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { BeziqueMeld, BeziqueResponse } from '../types/card';
import { BeziquePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { suitSymbolAt } from '../utils/cardAlt';
import { cardLabel } from '../utils/cardUtils';
import { BEZIQUE_HELP, parseBeziqueCommand } from '../utils/cli/commands/beziqueCommands';
import { formatBeziqueState } from '../utils/cli/formatters/beziqueFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit-name i18n key suffixes indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 unused). */
const SUIT_KEYS = ['', 'spade', 'club', 'heart', 'diamond'] as const;

/** Meld-name i18n key suffixes indexed by meld type (0=marriage … 5=four jacks). */
const MELD_NAME_KEYS = ['marriage', 'bezique', 'fourAces', 'fourKings', 'fourQueens', 'fourJacks'] as const;

/**
 * Stock (talon) count at or below which the pre-endgame warning is shown. When the
 * stock empties the game jumps to phase 2 (strict must-follow, no melds), so this
 * gives the player a few turns' notice to finish declaring melds.
 */
const STOCK_LOW_THRESHOLD = 4;

// Keep these values synchronized with domain.AceTenCardPoints. The frontend
// bundle cannot import Go domain code, so the i18n sentence receives them here.
const BEZIQUE_TRICK_POINT_VALUES = {
  ace: 11,
  ten: 10,
  king: 4,
  queen: 3,
  jack: 2,
  other: 0,
} as const;

/** Bezique tutorial step definitions. */
const BEZIQUE_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="bezique-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="bezique-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="bezique-player-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="bezique-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const BEZIQUE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BeziquePhase.PLAY]: 'play',
  [BeziquePhase.MELD]: 'meld',
  [BeziquePhase.ROUND_END]: 'roundEnd',
  [BeziquePhase.GAME_END]: 'gameEnd',
};

type BeziqueGame = ReturnType<typeof useBeziqueGame>;
type BeziqueView = ReturnType<typeof useBeziqueView>;
type BeziqueContext = GamePageContext<BeziqueResponse, BeziqueGame, BeziqueView>;

function useBeziqueView({ game, t }: { game: BeziqueGame; t: BeziqueContext['t'] }) {
  const { state } = game;

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const previousEndgame = useRef<boolean | undefined>(undefined);
  const [endgameAnnouncement, setEndgameAnnouncement] = useState('');

  useEffect(() => {
    if (state?.isEndgame && previousEndgame.current === false) {
      setEndgameAnnouncement(t('endgameLive'));
    } else if (!state?.isEndgame) {
      setEndgameAnnouncement('');
    }
    if (state) previousEndgame.current = state.isEndgame;
  }, [state, t]);

  // A meld selection belongs to the meld list it was made on, so a new list
  // (another meld turn, a reset) drops it without an effect.
  const availableMeldsKey = JSON.stringify(state?.availableMelds ?? []);
  const [meldSelection, setMeldSelection] = useState<{ key: string; index: number } | null>(null);
  const selectedMeldIndex = meldSelection?.key === availableMeldsKey ? meldSelection.index : null;
  const setSelectedMeldIndex = (index: number | null) =>
    setMeldSelection(index == null ? null : { key: availableMeldsKey, index });

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('bezique', BEZIQUE_PHASE_KEYS);

  return {
    endgameAnnouncement,
    selectedMeldIndex,
    setSelectedMeldIndex,
    cardWidth,
    isMobile,
    phaseNames,
  };
}

function LiveRegion({ view }: BeziqueContext) {
  return <LiveAnnouncement message={view.endgameAnnouncement} testId="bezique-endgame-live" />;
}

function Board({ state, t, view }: BeziqueContext) {
  const { cardWidth } = view;
  const isRoundEnd = state.phase === BeziquePhase.ROUND_END;
  const isGameEnd = state.phase === BeziquePhase.GAME_END || state.gameEndFlag;

  const trumpSymbol = state.trumpSuit >= 1 && state.trumpSuit <= 4 ? suitSymbolAt(state.trumpSuit, '') : t('noTrump');

  // Warn as the stock nears empty (but before the endgame actually begins): once it
  // hits 0 the game switches to phase 2 and the existing endgameNotice takes over.
  const stockLow = !state.isEndgame && state.stockRemaining > 0 && state.stockRemaining <= STOCK_LOW_THRESHOLD;

  return (
    <>
      <div
        className="flex flex-wrap justify-center gap-x-4 gap-y-1 text-ds-text-primary text-center mb-2"
        data-tutorial="bezique-info"
      >
        <span className="whitespace-nowrap">{t('deal', { n: state.roundNumber })}</span>
        <span className="whitespace-nowrap">{t('trick', { n: state.trickNumber })}</span>
        <span className="whitespace-nowrap">{t('trump', { suit: trumpSymbol })}</span>
        <span
          className={`whitespace-nowrap ${stockLow ? 'text-ds-warning font-semibold motion-safe:animate-pulse' : ''}`}
        >
          {t('stock', { count: state.stockRemaining })}
        </span>
        <span className="whitespace-nowrap">{t('target', { points: state.config.targetScore })}</span>
      </div>

      <div className="text-ds-text-muted text-center mb-2 text-xs" data-testid="bezique-trick-points">
        {t('trickPoints', BEZIQUE_TRICK_POINT_VALUES)}
      </div>

      {stockLow && (
        <div
          role="status"
          aria-live="polite"
          data-testid="bezique-stock-warning"
          className="text-ds-warning text-center mb-2 text-sm font-semibold motion-safe:animate-pulse"
        >
          {t('stockLowWarning', { count: state.stockRemaining })}
        </div>
      )}

      {state.isEndgame && (
        <div className="text-ds-warning text-center mb-2 text-sm font-semibold">{t('endgameNotice')}</div>
      )}

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="bezique-trick-display"
          />
          {state.trumpCard && !state.isEndgame && (
            <div className="mt-2 text-center text-ds-text-muted text-sm">
              {t('trumpCard', { card: cardLabel(state.trumpCard) })}
            </div>
          )}
        </div>

        {/* Right: score sidebar */}
        <div>
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            <div className="mb-1 text-ds-text-primary text-xs">{t('matchScoreTitle')}</div>
            {state.players.map((p) => (
              <div key={p.id} className={p.isHuman ? 'text-ds-text-primary' : ''}>
                <div>
                  {t('scoreLine', {
                    name: p.isHuman ? t('you') : t('cpu', { id: p.id }),
                    match: state.matchScore[p.id] ?? 0,
                    deal: state.dealPoints[p.id] ?? 0,
                    tricks: p.trickCount,
                  })}
                </div>
                {/* The match is decided by reaching targetScore, not by any one
                    deal, so how far off it is outranks every other number here. */}
                <div
                  className="text-ds-text-muted text-xs"
                  data-testid={`bezique-match-progress-${p.id}`}
                  {...(p.isHuman ? { role: 'status', 'aria-live': 'polite' as const } : {})}
                >
                  {(() => {
                    const current = state.matchScore[p.id];
                    const target = state.config.targetScore;
                    const remaining = target - current;
                    return remaining > 0
                      ? t('matchProgress', { current, target, remaining })
                      : t('matchProgressReached', { current, target });
                  })()}
                </div>
                <div className="text-ds-text-muted text-xs" data-testid={`bezique-deal-breakdown-${p.id}`}>
                  {t('dealBreakdown', {
                    trick: state.dealPoints[p.id] - state.dealMeldPoints[p.id],
                    meld: state.dealMeldPoints[p.id],
                    total: state.dealPoints[p.id],
                  })}
                </div>
              </div>
            ))}
          </div>

          {(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              {state.players.map((p) => (
                <div key={p.id}>
                  {t('roundResult.line', {
                    name: p.isHuman ? t('you') : t('cpu', { id: p.id }),
                    points: state.dealPoints[p.id] ?? 0,
                  })}
                  {state.lastTrickBonus[p.id] > 0 && (
                    <div data-testid={`bezique-last-trick-bonus-${p.id}`}>
                      {t('roundResult.lastTrickBonus', { points: state.lastTrickBonus[p.id] })}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ game, state, t, view }: BeziqueContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isPlayPhase = state.phase === BeziquePhase.PLAY;
  const isMeldPhase = state.phase === BeziquePhase.MELD;
  const isHumanCurrent = state.currentPlayerIdx === humanIdx;
  const isHumanPlayTurn = isPlayPhase && isHumanCurrent;
  const isHumanMeldTurn = isMeldPhase && isHumanCurrent;

  const selectedMeld = view.selectedMeldIndex == null ? undefined : state.availableMelds[view.selectedMeldIndex];
  const meldCardIndices = isHumanMeldTurn ? selectedMeld?.cardIndices : undefined;

  // The server computes legality from internal/domain/Bezique.go validatePlay (ADR-0045).
  // Keep the existing endgame follower display gate; the backend validates every play.
  const leadCard = state.currentTrick[0]?.card ?? null;
  const legalIndices = isHumanPlayTurn && state.isEndgame && leadCard != null ? state.validPlays : undefined;

  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="bezique"
      restrictedTooltip={t('playButton')}
      legalIndices={legalIndices}
      trumpIndices={meldCardIndices}
      trumpTitle={t('meldCardMarker')}
      trumpAccessibleLabel={t('meldCardMarker')}
    />
  );
}

function FooterStatus({ frontendHint, state, t }: BeziqueContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="bezique-hint-live" role="status" aria-live="polite">
        {state.hint && isRequestedHint(state) && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
            {state.hint.cardIndex != null && ` ([${state.hint.cardIndex}])`}
            {state.hint.meldIndex != null &&
              state.hint.meldIndex >= 0 &&
              ` ([${t('meldShort')} ${state.hint.meldIndex}])`}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

/** Builds a localized label for one declarable meld. */
function meldLabel(m: BeziqueMeld, t: BeziqueContext['t']): string {
  const name = t(`meldName.${MELD_NAME_KEYS[m.type] ?? 'marriage'}`);
  const suit = m.type === 0 && m.suit >= 1 && m.suit <= 4 ? ` ${suitSymbolAt(m.suit, '')}` : '';
  return `${name}${suit} (+${m.points})`;
}

/** Builds a screen-reader label naming the suit (a marriage is suit-specific) and points. */
function meldAriaLabel(m: BeziqueMeld, t: BeziqueContext['t']): string {
  const name = t(`meldName.${MELD_NAME_KEYS[m.type] ?? 'marriage'}`);
  const fullName =
    m.type === 0 && m.suit >= 1 && m.suit <= 4
      ? t('meldMarriageName', { suit: t(`suitName.${SUIT_KEYS[m.suit]}`), name })
      : name;
  return t('meldDeclareLabel', { name: fullName, points: m.points });
}

function FooterActions({ game, state, t, view }: BeziqueContext) {
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isPlayPhase = state.phase === BeziquePhase.PLAY;
  const isMeldPhase = state.phase === BeziquePhase.MELD;
  const isRoundEnd = state.phase === BeziquePhase.ROUND_END;
  const isHumanCurrent = state.currentPlayerIdx === humanIdx;
  const isHumanPlayTurn = isPlayPhase && isHumanCurrent;
  const isHumanMeldTurn = isMeldPhase && isHumanCurrent;
  const canPlay = isHumanPlayTurn;

  const { selectedMeldIndex, setSelectedMeldIndex } = view;
  const selectedMeld = selectedMeldIndex == null ? undefined : state.availableMelds[selectedMeldIndex];

  return (
    <>
      {canPlay && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handlePlay}
          disabled={game.loading || game.selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}
      {isHumanMeldTurn && (
        <>
          {/* Fieldset groups the meld buttons; the sr-only legend names the group
              (the visible prompt is aria-hidden to avoid a duplicate announcement). */}
          <fieldset className="contents">
            <legend className="sr-only">{t('meldPrompt')}</legend>
            <span aria-hidden="true" className="text-xs text-ds-text-muted self-center mr-1">
              {t('meldPrompt')}
            </span>
            {state.availableMelds.map((m, i) => (
              <button
                key={`${m.type}-${m.suit}`}
                type="button"
                className="px-3 py-2 rounded-lg bg-ds-info text-ds-text-on-accent text-sm disabled:opacity-40"
                onClick={() => setSelectedMeldIndex(i)}
                disabled={game.loading}
                data-testid={`meld-${i}`}
                aria-label={meldAriaLabel(m, t)}
                aria-pressed={selectedMeldIndex === i}
              >
                {meldLabel(m, t)}
              </button>
            ))}
          </fieldset>
          {selectedMeld && selectedMeldIndex != null && (
            <button
              type="button"
              className={btnPrimary}
              onClick={() => {
                setSelectedMeldIndex(null);
                game.handleMeld(selectedMeldIndex);
              }}
              disabled={game.loading}
              data-testid="meld-declare-selected"
            >
              {t('declareSelectedMeld')}
            </button>
          )}
          <button
            type="button"
            className={btnSuccess}
            onClick={() => {
              setSelectedMeldIndex(null);
              game.handleSkipMeld();
            }}
            disabled={game.loading}
            data-testid="meld-skip"
          >
            {t('skipMeld')}
          </button>
        </>
      )}
      <LiveAnnouncement
        message={
          isHumanMeldTurn
            ? state.availableMelds.length > 0
              ? t('meldLiveCount', { count: state.availableMelds.length })
              : t('meldLiveNone')
            : ''
        }
        testId="bezique-meld-live"
      />
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={game.loading}>
          {t('nextRound')}
        </button>
      )}
    </>
  );
}

const beziqueDefinition: GameDefinition<
  BeziqueResponse,
  BeziqueGame,
  Parameters<typeof beziqueApi.exec>,
  BeziqueView
> = {
  key: 'bezique',
  useGame: useBeziqueGame,
  useView: useBeziqueView,
  tutorial: BEZIQUE_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 9 },
  cli: () => ({
    parseCommand: parseBeziqueCommand,
    formatResponse: formatBeziqueState,
    helpText: BEZIQUE_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === BeziquePhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerIdx === humanIdx,
    };
  },
  shell: ({ state }) => {
    const isPlayPhase = state.phase === BeziquePhase.PLAY;
    const isMeldPhase = state.phase === BeziquePhase.MELD;
    const isGameEnd = state.phase === BeziquePhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    const isHumanCurrent = state.currentPlayerIdx === humanIdx;
    const isHumanPlayTurn = isPlayPhase && isHumanCurrent;
    const isHumanMeldTurn = isMeldPhase && isHumanCurrent;
    return {
      isHumanTurn: (isHumanPlayTurn || isHumanMeldTurn) && !isGameEnd,
    };
  },
  reset: ({ game }) => {
    game.reset();
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'targetScore',
          label: t('settings.targetScore'),
          value: game.beziqueConfig.targetScore,
          options: TARGET_SCORE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetScore', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'bezique-action-buttons',
  resetTutorialId: 'bezique-reset-button',
  LiveRegion,
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Bezique game page: a 2-player melding trick-taker (ancestor of Pinochle). */
export const BeziquePage = createGamePage(beziqueDefinition);
