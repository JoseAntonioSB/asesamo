'use client';

import React, { useEffect, useState } from 'react';
import QRCode from 'qrcode';
import { QrCode, Maximize2, Copy, Check, SunMedium, X, Sparkles } from 'lucide-react';

interface WorkerQRCodeProps {
  payload: string;
  workerName: string;
  className?: string;
  complianceStatus?: 'granted' | 'granted_with_warning' | 'denied';
}

export function WorkerQRCode({
  payload,
  workerName,
  className = '',
  complianceStatus = 'granted',
}: WorkerQRCodeProps) {
  const [qrDataUrl, setQrDataUrl] = useState<string>('');
  const [largeQrDataUrl, setLargeQrDataUrl] = useState<string>('');
  const [isCopied, setIsCopied] = useState(false);
  const [isScannerModalOpen, setIsScannerModalOpen] = useState(false);

  useEffect(() => {
    if (!payload) return;

    // Generate normal card QR code
    QRCode.toDataURL(payload, {
      width: 240,
      margin: 1,
      errorCorrectionLevel: 'M',
      color: {
        dark: '#0f172a',
        light: '#ffffff',
      },
    })
      .then((url) => setQrDataUrl(url))
      .catch((err) => console.error('Error generating QR code:', err));

    // Generate high-resolution full-screen QR code for turnstile readers
    QRCode.toDataURL(payload, {
      width: 480,
      margin: 2,
      errorCorrectionLevel: 'H',
      color: {
        dark: '#000000',
        light: '#ffffff',
      },
    })
      .then((url) => setLargeQrDataUrl(url))
      .catch((err) => console.error('Error generating large QR code:', err));
  }, [payload]);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(payload);
      setIsCopied(true);
      setTimeout(() => setIsCopied(false), 2000);
    } catch {
      // Fallback if clipboard API is unavailable
    }
  };

  const getBorderColor = () => {
    switch (complianceStatus) {
      case 'granted':
        return 'border-emerald-500/50 shadow-emerald-500/10';
      case 'granted_with_warning':
        return 'border-amber-500/50 shadow-amber-500/10';
      case 'denied':
        return 'border-rose-500/50 shadow-rose-500/10';
      default:
        return 'border-slate-700/50';
    }
  };

  return (
    <>
      <div className={`flex flex-col items-center text-center ${className}`}>
        {/* QR Code Frame with Industrial Tactile Container */}
        <div
          className={`relative p-3 rounded-2xl bg-white border-2 shadow-xl transition-all ${getBorderColor()}`}
        >
          {qrDataUrl ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={qrDataUrl}
              alt={`Código QR de acceso para ${workerName}`}
              className="w-44 h-44 sm:w-48 sm:h-48 object-contain rounded-lg select-none"
            />
          ) : (
            <div className="w-44 h-44 sm:w-48 sm:h-48 flex items-center justify-center bg-slate-100 rounded-lg text-slate-400">
              <QrCode className="w-12 h-12 animate-pulse" />
            </div>
          )}

          {/* Quick Expand Button overlay */}
          <button
            type="button"
            onClick={() => setIsScannerModalOpen(true)}
            className="absolute bottom-2 right-2 p-1.5 rounded-lg bg-slate-900/80 hover:bg-slate-950 text-white backdrop-blur shadow-md active:scale-95 transition-all"
            title="Ampliar QR para escaneo en caseta"
            aria-label="Ampliar QR para escaneo"
          >
            <Maximize2 className="w-3.5 h-3.5" />
          </button>
        </div>

        {/* UUID readout & quick copy */}
        <div className="mt-3 flex items-center gap-2 max-w-full">
          <div className="px-2.5 py-1 rounded-md bg-slate-900/90 dark:bg-slate-900 text-slate-300 border border-slate-800 text-[11px] font-mono tracking-tight truncate max-w-[220px]">
            {payload}
          </div>
          <button
            type="button"
            onClick={handleCopy}
            className="p-1.5 rounded-md border border-slate-700/70 bg-slate-800/60 hover:bg-slate-700 text-slate-300 hover:text-white transition-colors cursor-pointer"
            title="Copiar UUID para validación manual"
            aria-label="Copiar identificador de trabajador"
          >
            {isCopied ? (
              <Check className="w-3.5 h-3.5 text-emerald-400" />
            ) : (
              <Copy className="w-3.5 h-3.5" />
            )}
          </button>
        </div>

        {/* Tap to expand prompt */}
        <button
          type="button"
          onClick={() => setIsScannerModalOpen(true)}
          className="mt-2 text-[11px] font-semibold text-emerald-600 dark:text-emerald-400 hover:underline flex items-center gap-1 cursor-pointer"
        >
          <SunMedium className="w-3 h-3" />
          <span>Presentar en torniquete (Modo Brillo)</span>
        </button>
      </div>

      {/* Full-Screen Scanner Modal for Industrial Optical Readers */}
      {isScannerModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/95 flex flex-col items-center justify-between p-6 animate-in fade-in duration-200">
          {/* Header */}
          <div className="w-full max-w-sm flex items-center justify-between text-white">
            <div className="flex items-center gap-2">
              <SunMedium className="w-5 h-5 text-amber-400" />
              <span className="text-sm font-bold uppercase tracking-wider">Modo Escáner Activo</span>
            </div>
            <button
              type="button"
              onClick={() => setIsScannerModalOpen(false)}
              className="p-2 rounded-full bg-white/10 hover:bg-white/20 text-white transition-colors cursor-pointer"
              aria-label="Cerrar modo escáner"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Central High-Contrast QR */}
          <div className="flex flex-col items-center text-center my-auto">
            <div className="p-6 rounded-3xl bg-white shadow-2xl shadow-emerald-500/20 border-4 border-emerald-500">
              {largeQrDataUrl ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  src={largeQrDataUrl}
                  alt={`QR de alta resolución para ${workerName}`}
                  className="w-64 h-64 sm:w-80 sm:h-80 object-contain select-none"
                />
              ) : (
                <div className="w-64 h-64 flex items-center justify-center">
                  <QrCode className="w-16 h-16 animate-pulse text-slate-800" />
                </div>
              )}
            </div>

            <h3 className="mt-6 text-xl font-bold text-white tracking-tight">{workerName}</h3>
            <p className="mt-1 text-xs font-mono text-emerald-400 tracking-wider">
              ASESAMO PASS: {payload.slice(0, 8)}...{payload.slice(-4)}
            </p>
            <p className="mt-2 text-xs text-slate-400 max-w-xs">
              Acerque esta pantalla al lector óptico del torniquete o a la cámara del guardia en caseta.
            </p>
          </div>

          {/* Footer Close Button */}
          <button
            type="button"
            onClick={() => setIsScannerModalOpen(false)}
            className="w-full max-w-sm py-3 px-4 rounded-xl bg-slate-800 hover:bg-slate-700 text-white font-bold text-sm tracking-wide transition-colors cursor-pointer"
          >
            Volver a la Credencial
          </button>
        </div>
      )}
    </>
  );
}
