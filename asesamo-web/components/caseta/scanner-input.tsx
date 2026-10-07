// asesamo-web/components/caseta/scanner-input.tsx
'use client';

import React, { useState, useRef, useEffect, useCallback } from 'react';
import { ScanBarcode, Loader2, X, ArrowRight, Keyboard } from 'lucide-react';

export interface ScannerInputProps {
  onScan: (workerId: string) => void | Promise<void>;
  isLoading?: boolean;
  disabled?: boolean;
}

export function ScannerInput({ onScan, isLoading = false, disabled = false }: ScannerInputProps) {
  const [inputValue, setInputValue] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);

  // Auto-focus on component mount
  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  // Restore focus automatically when loading finishes
  useEffect(() => {
    if (!isLoading && !disabled) {
      inputRef.current?.focus();
    }
  }, [isLoading, disabled]);

  // Global keydown recovery: If user is typing and not focused on another input, redirect focus
  useEffect(() => {
    const handleGlobalKeyDown = (e: KeyboardEvent) => {
      // Ignore if user is already inside an input or textarea
      const target = e.target as HTMLElement;
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') {
        return;
      }

      // If F2 or '/' is pressed, or if typing alphanumeric characters
      if (e.key === 'F2' || e.key === '/') {
        e.preventDefault();
        inputRef.current?.focus();
      }
    };

    window.addEventListener('keydown', handleGlobalKeyDown);
    return () => window.removeEventListener('keydown', handleGlobalKeyDown);
  }, []);

  const handleTriggerScan = useCallback(() => {
    const clean = inputValue.trim();
    if (!clean || isLoading || disabled) return;

    onScan(clean);
    setInputValue('');
    // Ensure focus is retained for the next scan immediately
    setTimeout(() => {
      inputRef.current?.focus();
    }, 10);
  }, [inputValue, isLoading, disabled, onScan]);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    // Standard HID barcode scanner hardware ends its payload with Enter
    if (e.key === 'Enter') {
      e.preventDefault();
      handleTriggerScan();
    }
  };

  return (
    <div className="w-full rounded-2xl border-2 border-slate-700 bg-slate-900/90 p-5 shadow-2xl backdrop-blur transition-all focus-within:border-emerald-500 focus-within:ring-4 focus-within:ring-emerald-500/20">
      <div className="flex flex-col gap-3">
        {/* Label & Instructions */}
        <div className="flex items-center justify-between text-xs text-slate-400">
          <label htmlFor="caseta-scanner-input" className="flex items-center gap-2 font-bold text-slate-200 uppercase tracking-wider">
            <ScanBarcode className="h-4 w-4 text-emerald-400" />
            Lector de Código de Barras / QR (HID Wedge)
          </label>
          <span className="hidden sm:inline-flex items-center gap-1 text-[11px] font-mono text-slate-400 bg-slate-800/80 px-2 py-0.5 rounded border border-slate-700">
            <Keyboard className="h-3 w-3" />
            Auto-enfoque activo • Presione [Enter] para verificar
          </span>
        </div>

        {/* Primary Input Container */}
        <div className="relative flex items-center">
          <div className="pointer-events-none absolute left-4 text-slate-400">
            {isLoading ? (
              <Loader2 className="h-6 w-6 animate-spin text-emerald-400" />
            ) : (
              <ScanBarcode className="h-6 w-6 text-slate-400" />
            )}
          </div>

          <input
            id="caseta-scanner-input"
            ref={inputRef}
            type="text"
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={disabled || isLoading}
            placeholder="Escanee el código de barras o ingrese UUID del trabajador..."
            className="w-full rounded-xl border border-slate-700 bg-slate-950/80 py-4 pl-14 pr-32 font-mono text-base sm:text-lg text-white placeholder-slate-500 outline-none transition-all focus:border-emerald-500 focus:bg-slate-950"
            autoComplete="off"
            autoCorrect="off"
            spellCheck="false"
          />

          {/* Right Action Buttons */}
          <div className="absolute right-2.5 flex items-center gap-1.5">
            {inputValue && !isLoading && (
              <button
                type="button"
                onClick={() => {
                  setInputValue('');
                  inputRef.current?.focus();
                }}
                className="rounded-lg p-2 text-slate-400 hover:bg-slate-800 hover:text-white transition-colors cursor-pointer"
                title="Limpiar campo"
              >
                <X className="h-4 w-4" />
              </button>
            )}

            <button
              type="button"
              onClick={handleTriggerScan}
              disabled={!inputValue.trim() || isLoading || disabled}
              className="inline-flex items-center gap-1.5 rounded-lg bg-emerald-600 px-4 py-2.5 text-sm font-bold text-white shadow-lg transition-all hover:bg-emerald-500 active:scale-95 disabled:pointer-events-none disabled:opacity-40 cursor-pointer"
            >
              {isLoading ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <span>Validando...</span>
                </>
              ) : (
                <>
                  <span>Verificar</span>
                  <ArrowRight className="h-4 w-4" />
                </>
              )}
            </button>
          </div>
        </div>

        {/* Sub-instruction note */}
        <div className="flex items-center justify-between text-[11px] text-slate-400">
          <span>Acepta códigos QR de gafetes industriales, pistolas láser USB y entrada manual.</span>
          <span className="font-mono text-slate-400">Atajo: F2 para enfocar</span>
        </div>
      </div>
    </div>
  );
}
