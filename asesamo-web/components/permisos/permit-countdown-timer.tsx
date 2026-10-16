'use client';

import React, { useState, useEffect } from 'react';
import { Clock, AlertCircle, CheckCircle2 } from 'lucide-react';
import { PermitStatus } from '@/lib/api/types';

interface PermitCountdownTimerProps {
  startTime: string;
  endTime: string;
  status: PermitStatus | string;
  showIcon?: boolean;
  className?: string;
}

interface TimeRemaining {
  isExpired: boolean;
  isNearExpiry: boolean; // < 2 hours
  hours: number;
  minutes: number;
  seconds: number;
  totalSeconds: number;
  displayText: string;
}

export function PermitCountdownTimer({
  startTime,
  endTime,
  status,
  showIcon = true,
  className = '',
}: PermitCountdownTimerProps) {
  const [now, setNow] = useState<number>(Date.now());

  useEffect(() => {
    // Only tick live if status is active
    if (status !== 'approved_active') return;

    const interval = setInterval(() => {
      setNow(Date.now());
    }, 1000);

    return () => clearInterval(interval);
  }, [status]);

  const calculateRemaining = (): TimeRemaining => {
    const endMs = new Date(endTime).getTime();
    const diffMs = endMs - now;

    if (isNaN(diffMs) || diffMs <= 0) {
      return {
        isExpired: true,
        isNearExpiry: false,
        hours: 0,
        minutes: 0,
        seconds: 0,
        totalSeconds: 0,
        displayText: 'Vigencia Concluida (Vencido)',
      };
    }

    const totalSeconds = Math.floor(diffMs / 1000);
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;

    const isNearExpiry = hours < 2;

    const formattedH = String(hours).padStart(2, '0');
    const formattedM = String(minutes).padStart(2, '0');
    const formattedS = String(seconds).padStart(2, '0');

    return {
      isExpired: false,
      isNearExpiry,
      hours,
      minutes,
      seconds,
      totalSeconds,
      displayText: `${formattedH}h ${formattedM}m ${formattedS}s restantes`,
    };
  };

  // If status is closed or cancelled
  if (status === 'closed') {
    return (
      <span className={`inline-flex items-center gap-1.5 text-xs text-slate-400 font-mono ${className}`}>
        <CheckCircle2 className="w-3.5 h-3.5 text-slate-500" />
        <span>Concluido</span>
      </span>
    );
  }

  if (status === 'cancelled') {
    return (
      <span className={`inline-flex items-center gap-1.5 text-xs text-slate-500 font-mono ${className}`}>
        <span>Cancelado</span>
      </span>
    );
  }

  if (status === 'suspended') {
    return (
      <span className={`inline-flex items-center gap-1.5 text-xs text-rose-400 font-semibold font-mono ${className}`}>
        <AlertCircle className="w-3.5 h-3.5 text-rose-400 animate-pulse" />
        <span>Operación Pausada</span>
      </span>
    );
  }

  // If status is draft or pending_approval, show scheduled duration
  if (status === 'draft' || status === 'pending_approval') {
    try {
      const start = new Date(startTime).getTime();
      const end = new Date(endTime).getTime();
      const durationHours = Math.max(0, Math.round((end - start) / (1000 * 3600)));
      return (
        <span className={`inline-flex items-center gap-1.5 text-xs text-slate-400 font-mono ${className}`}>
          {showIcon && <Clock className="w-3.5 h-3.5 text-slate-500" />}
          <span>Programado ({durationHours} hrs)</span>
        </span>
      );
    } catch {
      return <span className="text-xs text-slate-400">Programado</span>;
    }
  }

  // Active status countdown
  const remaining = calculateRemaining();

  if (remaining.isExpired) {
    return (
      <span
        className={`inline-flex items-center gap-1.5 text-xs font-mono font-bold text-rose-400 bg-rose-950/70 border border-rose-800/80 px-2.5 py-1 rounded-md ${className}`}
      >
        {showIcon && <AlertCircle className="w-3.5 h-3.5 text-rose-400 shrink-0" />}
        <span>{remaining.displayText}</span>
      </span>
    );
  }

  if (remaining.isNearExpiry) {
    return (
      <span
        className={`inline-flex items-center gap-1.5 text-xs font-mono font-bold text-amber-300 bg-amber-950/60 border border-amber-700/70 px-2.5 py-1 rounded-md animate-pulse ${className}`}
      >
        {showIcon && <Clock className="w-3.5 h-3.5 text-amber-400 shrink-0" />}
        <span>{remaining.displayText}</span>
      </span>
    );
  }

  return (
    <span
      className={`inline-flex items-center gap-1.5 text-xs font-mono font-medium text-emerald-300 bg-emerald-950/50 border border-emerald-700/60 px-2.5 py-1 rounded-md ${className}`}
    >
      {showIcon && <Clock className="w-3.5 h-3.5 text-emerald-400 shrink-0" />}
      <span>{remaining.displayText}</span>
    </span>
  );
}
