import type React from 'react';

export function selectedCardStyle(isSelected: boolean): React.CSSProperties {
  return {
    border: isSelected ? '3px solid var(--color-game-card-selected)' : '3px solid transparent',
    transform: isSelected ? 'translateY(-8px)' : 'none',
    transition: 'transform 0.15s, border 0.15s',
    boxShadow: isSelected ? '0 4px 12px rgba(59, 130, 246, 0.4)' : 'none',
  };
}

export function playableCardStyle(isPlayable: boolean): React.CSSProperties {
  return {
    border: isPlayable ? '3px solid var(--color-game-status-active)' : '3px solid transparent',
  };
}
