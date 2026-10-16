/**
 * ASESAMO — Security Guard Login Screen
 * Location: asesamo-web/app/login/page.tsx
 */

'use client';

import React, { useState, useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/contexts/auth-context';
import { ApiError } from '@/lib/api/client';
import { getDefaultHomeForRole } from '@/lib/auth';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { ShieldCheck, AlertTriangle, KeyRound, UserCheck, Loader2 } from 'lucide-react';

function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { login, isAuthenticated, user } = useAuth();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errorBanner, setErrorBanner] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // If already authenticated, redirect immediately to role home
  useEffect(() => {
    if (isAuthenticated) {
      const redirect = searchParams.get('redirect') || getDefaultHomeForRole(user?.role);
      router.replace(redirect);
    }
  }, [isAuthenticated, user, router, searchParams]);

  // Handle URL error param (e.g. session expired)
  useEffect(() => {
    const err = searchParams.get('error');
    if (err === 'session_expired') {
      setErrorBanner('Su sesión ha vencido. Por favor ingrese sus credenciales nuevamente.');
    }
  }, [searchParams]);

  // Quick-fill demo credentials
  const handleQuickFill = (userEmail = 'guard@asesamo.com') => {
    setEmail(userEmail);
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
      const res = await login(email.trim(), password);

      const redirectParam = searchParams.get('redirect');
      const targetHome = redirectParam || getDefaultHomeForRole(res.user.role);
      router.push(targetHome);
    } catch (err: unknown) {
      setIsSubmitting(false);
      if (err instanceof ApiError) {
        if (err.code === 'INVALID_CREDENTIALS') {
          setErrorBanner('Credenciales inválidas. Verifique su correo y contraseña.');
        } else if (err.code === 'TOO_MANY_REQUESTS') {
          setErrorBanner('Demasiados intentos fallidos. Espere un momento antes de volver a intentar.');
        } else if (err.code === 'USER_INACTIVE') {
          setErrorBanner('Su cuenta se encuentra inactiva.');
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
      {/* Top right Theme Toggle */}
      <div className="absolute top-4 right-4 sm:top-6 sm:right-6 z-20">
        <ThemeToggle showLabel />
      </div>

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
                Cuentas de demostración (Click para autollenar):
              </span>
            </div>

            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={() => handleQuickFill('guard@asesamo.com')}
                disabled={isSubmitting}
                className={`p-2 rounded-lg border text-left transition-colors text-xs font-mono ${
                  email === 'guard@asesamo.com'
                    ? 'border-emerald-500 bg-emerald-950/40 text-emerald-300'
                    : 'border-slate-800 bg-slate-950/80 hover:border-slate-700 text-slate-300'
                }`}
              >
                <div className="font-semibold text-[11px] text-emerald-400">👮 Guardia / Caseta</div>
                <div className="truncate text-slate-400 text-[10px]">guard@asesamo.com</div>
              </button>

              <button
                type="button"
                onClick={() => handleQuickFill('contratista@mantenimientonorte.com')}
                disabled={isSubmitting}
                className={`p-2 rounded-lg border text-left transition-colors text-xs font-mono ${
                  email === 'contratista@mantenimientonorte.com'
                    ? 'border-emerald-500 bg-emerald-950/40 text-emerald-300'
                    : 'border-slate-800 bg-slate-950/80 hover:border-slate-700 text-slate-300'
                }`}
              >
                <div className="font-semibold text-[11px] text-amber-400">🏢 Admin Contratista</div>
                <div className="truncate text-slate-400 text-[10px]">contratista@mantenimiento...</div>
              </button>

              <button
                type="button"
                onClick={() => handleQuickFill('ehs@asesamo.com')}
                disabled={isSubmitting}
                className={`p-2 rounded-lg border text-left transition-colors text-xs font-mono ${
                  email === 'ehs@asesamo.com'
                    ? 'border-emerald-500 bg-emerald-950/40 text-emerald-300'
                    : 'border-slate-800 bg-slate-950/80 hover:border-slate-700 text-slate-300'
                }`}
              >
                <div className="font-semibold text-[11px] text-blue-400">🛡️ EHS Manager</div>
                <div className="truncate text-slate-400 text-[10px]">ehs@asesamo.com</div>
              </button>

              <button
                type="button"
                onClick={() => handleQuickFill('admin@asesamo.com')}
                disabled={isSubmitting}
                className={`p-2 rounded-lg border text-left transition-colors text-xs font-mono ${
                  email === 'admin@asesamo.com'
                    ? 'border-emerald-500 bg-emerald-950/40 text-emerald-300'
                    : 'border-slate-800 bg-slate-950/80 hover:border-slate-700 text-slate-300'
                }`}
              >
                <div className="font-semibold text-[11px] text-purple-400">⚙️ Admin Planta</div>
                <div className="truncate text-slate-400 text-[10px]">admin@asesamo.com</div>
              </button>
            </div>

            <div className="text-center text-[11px] text-slate-500 font-mono">
              Contraseña única: <span className="text-slate-300 font-semibold">Password123!</span>
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
