'use client';

import React from 'react';
import { useTheme } from '@/contexts/theme-context';
import { Sun, Moon } from 'lucide-react';

interface ThemeToggleProps {
  className?: string;
  showLabel?: boolean;
}

export function ThemeToggle({ className = '', showLabel = false }: ThemeToggleProps) {
  const { theme, toggleTheme } = useTheme();

  const isDark = theme === 'dark';
  const label = isDark ? 'Modo Claro' : 'Modo Oscuro';
  const tooltip = isDark ? 'Cambiar a modo claro' : 'Cambiar a modo oscuro';

  return (
    <button
      type="button"
      onClick={toggleTheme}
      className={`inline-flex items-center gap-2 px-2.5 py-1.5 rounded-lg border border-slate-700/70 bg-slate-800/50 dark:border-slate-800 dark:bg-slate-900 text-slate-300 hover:text-white dark:hover:text-amber-300 active:scale-95 transition-all cursor-pointer no-print ${className}`}
      title={tooltip}
      aria-label={tooltip}
    >
      {isDark ? (
        <Sun className="w-4 h-4 text-amber-400" />
      ) : (
        <Moon className="w-4 h-4 text-indigo-600 dark:text-slate-300" />
      )}
      {showLabel && (
        <span className="text-xs font-semibold select-none">
          {label}
        </span>
      )}
    </button>
  );
}
