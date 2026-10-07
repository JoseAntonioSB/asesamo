/**
 * ASESAMO — Security Guard Login Screen
 * Location: asesamo-web/app/login/page.tsx
 */

'use client';

import React, { useState, useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/contexts/auth-context';
import { ApiError } from '@/lib/api/client';
import { ShieldCheck, AlertTriangle, KeyRound, UserCheck, Loader2 } from 'lucide-react';

function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { login, isAuthenticated } = useAuth();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errorBanner, setErrorBanner] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // If already authenticated, redirect immediately to /caseta
  useEffect(() => {
    if (isAuthenticated) {
      const redirect = searchParams.get('redirect') || '/caseta';
      router.replace(redirect);
    }
  }, [isAuthenticated, router, searchParams]);

  // Handle URL error param (e.g. session expired)
  useEffect(() => {
    const err = searchParams.get('error');
    if (err === 'session_expired') {
      setErrorBanner('Su sesión ha vencido. Por favor ingrese sus credenciales nuevamente.');
    }
  }, [searchParams]);

  // Quick-fill demo credentials
  const handleQuickFill = () => {
    setEmail('guard@asesamo.com');
    setPassword('Password123!');
    setErrorBanner(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorBanner(null);

    // Client-side validation
    if (!email.trim() || !password) {
      setErrorBanner('Por favor complete todos los campos obligatorios.');
      return;
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email.trim())) {
      setErrorBanner('Ingrese un correo electrónico con formato válido.');
      return;
    }

    try {
      setIsSubmitting(true);
      await login(email.trim(), password);

      const destination = searchParams.get('redirect') || '/caseta';
      router.push(destination);
    } catch (err: unknown) {
      setIsSubmitting(false);
      if (err instanceof ApiError) {
        if (err.code === 'INVALID_CREDENTIALS') {
          setErrorBanner('Credenciales inválidas. Verifique su correo y contraseña.');
        } else if (err.code === 'TOO_MANY_REQUESTS') {
          setErrorBanner('Demasiados intentos fallidos. Espere un momento antes de volver a intentar.');
        } else if (err.code === 'USER_INACTIVE') {
          setErrorBanner('Su cuenta de oficial de caseta se encuentra inactiva.');
        } else {
          setErrorBanner(err.message || 'Error al iniciar sesión.');
        }
      } else {
        setErrorBanner('No se pudo establecer conexión con el servidor. Intente más tarde.');
      }
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col justify-center items-center px-4 sm:px-6 lg:px-8 selection:bg-emerald-500 selection:text-white relative">
      {/* Background radial accent */}
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,_var(--tw-gradient-stops))] from-slate-900 via-slate-950 to-slate-950 pointer-events-none" />

      <div className="relative w-full max-w-md space-y-6">
        {/* Brand Header */}
        <div className="text-center space-y-2">
          <div className="inline-flex items-center justify-center w-14 h-14 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 mb-2 shadow-lg shadow-emerald-950/40">
            <ShieldCheck className="w-8 h-8" />
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-white uppercase sm:text-3xl font-mono">
            ASESAMO
          </h1>
          <p className="text-sm text-slate-400">
            Control de Acceso Industrial y Verificación de Cumplimiento
          </p>
          <div className="inline-block px-3 py-1 text-xs font-semibold uppercase tracking-wider text-emerald-400 bg-emerald-950/60 border border-emerald-800/60 rounded-full">
            Módulo de Caseta de Vigilancia
          </div>
        </div>

        {/* Login Card */}
        <div className="bg-slate-900/90 border border-slate-800 shadow-2xl rounded-xl p-6 sm:p-8 backdrop-blur-sm space-y-6">
          {/* Spanish Error Alert Banner */}
          {errorBanner && (
            <div
              role="alert"
              className="bg-red-950/50 border border-red-500/50 rounded-lg p-3.5 flex items-start gap-3 text-red-200 text-sm animate-in fade-in duration-200"
            >
              <AlertTriangle className="w-5 h-5 text-red-400 shrink-0 mt-0.5" />
              <div className="flex-1 font-medium">{errorBanner}</div>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Email Field */}
            <div className="space-y-1.5">
              <label
                htmlFor="email"
                className="block text-xs font-semibold text-slate-300 uppercase tracking-wider"
              >
                Correo Electrónico
              </label>
              <input
                id="email"
                name="email"
                type="email"
                autoComplete="email"
                required
                disabled={isSubmitting}
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="guard@asesamo.com"
                className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-700/80 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/60 focus:border-emerald-500 transition-colors text-sm"
              />
            </div>

            {/* Password Field */}
            <div className="space-y-1.5">
              <label
                htmlFor="password"
                className="block text-xs font-semibold text-slate-300 uppercase tracking-wider"
              >
                Contraseña
              </label>
              <input
                id="password"
                name="password"
                type="password"
                autoComplete="current-password"
                required
                disabled={isSubmitting}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••••••"
                className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-700/80 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/60 focus:border-emerald-500 transition-colors text-sm"
              />
            </div>

            {/* Submit Button */}
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full mt-2 py-3 px-4 rounded-lg bg-emerald-600 hover:bg-emerald-500 active:bg-emerald-700 text-white font-semibold text-sm shadow-md shadow-emerald-950/50 transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin text-white" />
                  <span>Verificando credenciales...</span>
                </>
              ) : (
                <>
                  <KeyRound className="w-4 h-4" />
                  <span>Ingresar a Caseta</span>
                </>
              )}
            </button>
          </form>

          {/* Quick-Fill Demo Helpers */}
          <div className="pt-4 border-t border-slate-800/80 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-xs text-slate-400 font-medium flex items-center gap-1.5">
                <UserCheck className="w-3.5 h-3.5 text-emerald-400" />
                Credenciales de demostración:
              </span>
              <button
                type="button"
                onClick={handleQuickFill}
                disabled={isSubmitting}
                className="text-xs text-emerald-400 hover:text-emerald-300 font-medium underline underline-offset-4 transition-colors cursor-pointer"
              >
                Autocompletar
              </button>
            </div>

            <div className="bg-slate-950/80 border border-slate-800 rounded-lg p-2.5 text-xs font-mono text-slate-400 space-y-1">
              <div className="flex justify-between">
                <span className="text-slate-500">Usuario:</span>
                <span className="text-slate-200">guard@asesamo.com</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-500">Contraseña:</span>
                <span className="text-slate-200">Password123!</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-500">Rol:</span>
                <span className="text-emerald-400">Oficial de Caseta</span>
              </div>
            </div>
          </div>
        </div>

        {/* Industrial Footer Note */}
        <p className="text-center text-xs text-slate-400 font-mono">
          ASESAMO v1.0 • Planta Querétaro Norte • NOM-STPS
        </p>
      </div>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-slate-950 flex items-center justify-center text-slate-400">
          <Loader2 className="w-8 h-8 animate-spin text-emerald-500" />
        </div>
      }
    >
      <LoginForm />
    </Suspense>
  );
}
