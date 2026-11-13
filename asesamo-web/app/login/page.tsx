'use client';

import React, { useState, useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/contexts/auth-context';
import { ApiError } from '@/lib/api/client';
import { getDefaultHomeForRole } from '@/lib/auth';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { ShieldCheck, AlertTriangle, KeyRound, UserCheck, Loader2, Lock, ArrowRight, ShieldAlert, Cpu } from 'lucide-react';

function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { login, isAuthenticated, user } = useAuth();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errorBanner, setErrorBanner] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (isAuthenticated) {
      const redirect = searchParams.get('redirect') || getDefaultHomeForRole(user?.role);
      router.replace(redirect);
    }
  }, [isAuthenticated, user, router, searchParams]);

  useEffect(() => {
    const err = searchParams.get('error');
    if (err === 'session_expired') {
      setErrorBanner('Su sesión ha vencido. Por favor ingrese sus credenciales nuevamente.');
    }
  }, [searchParams]);

  const handleQuickFill = (userEmail = 'guard@asesamo.com') => {
    setEmail(userEmail);
    setPassword('Password123!');
    setErrorBanner(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorBanner(null);

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
    <div className="min-h-screen bg-zinc-950 flex flex-col lg:flex-row text-zinc-100 selection:bg-emerald-500 selection:text-white relative overflow-hidden">
      {/* Botón Flotante de Tema */}
      <div className="absolute top-6 right-6 z-30">
        <ThemeToggle showLabel />
      </div>

      {/* Lado Izquierdo: Formulario de Acceso y Cuentas Demo */}
      <div className="flex-1 flex flex-col justify-center px-6 sm:px-12 lg:px-20 py-12 z-20 relative">
        <div className="w-full max-w-md mx-auto space-y-8">
          {/* Logo y Encabezado */}
          <div className="space-y-3">
            <div className="inline-flex items-center gap-2.5 px-3 py-1.5 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-semibold tracking-wide">
              <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
              Sistema Central de Seguridad Industrial
            </div>
            <h1 className="text-3xl sm:text-4xl font-extrabold tracking-tight text-white font-mono">
              ASESAMO
            </h1>
            <p className="text-sm text-zinc-400 leading-relaxed">
              Autenticación para supervisores de caseta, administradores de planta y coordinadores EHS.
            </p>
          </div>

          {/* Tarjeta de Formulario */}
          <div className="bg-zinc-900/80 border border-zinc-800/90 rounded-2xl p-6 sm:p-8 backdrop-blur-xl shadow-2xl space-y-6">
            {errorBanner && (
              <div
                role="alert"
                className="bg-red-950/40 border border-red-500/40 rounded-xl p-4 flex items-start gap-3 text-red-200 text-xs font-medium"
              >
                <AlertTriangle className="w-4 h-4 text-red-400 shrink-0 mt-0.5" />
                <span>{errorBanner}</span>
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-1.5">
                <label
                  htmlFor="email"
                  className="block text-xs font-bold text-zinc-300 uppercase tracking-wider"
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
                  className="w-full px-4 py-3 bg-zinc-950/80 border border-zinc-800 rounded-xl text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 transition-all text-sm font-mono"
                />
              </div>

              <div className="space-y-1.5">
                <label
                  htmlFor="password"
                  className="block text-xs font-bold text-zinc-300 uppercase tracking-wider"
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
                  className="w-full px-4 py-3 bg-zinc-950/80 border border-zinc-800 rounded-xl text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/50 focus:border-emerald-500 transition-all text-sm font-mono"
                />
              </div>

              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full mt-3 py-3.5 px-6 rounded-xl bg-white hover:bg-zinc-200 active:bg-zinc-300 text-zinc-950 font-bold text-sm shadow-xl shadow-white/5 transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
              >
                {isSubmitting ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin text-zinc-950" />
                    <span>Verificando credenciales...</span>
                  </>
                ) : (
                  <>
                    <span>Ingresar a Plataforma</span>
                    <ArrowRight className="w-4 h-4" />
                  </>
                )}
              </button>
            </form>

            {/* Presets Demo */}
            <div className="pt-4 border-t border-zinc-800/80 space-y-3">
              <div className="flex items-center justify-between text-xs text-zinc-400 font-semibold">
                <span className="flex items-center gap-1.5 text-zinc-300">
                  <UserCheck className="w-4 h-4 text-emerald-400" />
                  Roles de Demostración:
                </span>
                <span className="text-[11px] text-zinc-400 font-mono">Password123!</span>
              </div>

              <div className="grid grid-cols-2 gap-2">
                <button
                  type="button"
                  onClick={() => handleQuickFill('guard@asesamo.com')}
                  disabled={isSubmitting}
                  className={`p-2.5 rounded-xl border text-left transition-all ${
                    email === 'guard@asesamo.com'
                      ? 'border-emerald-500 bg-emerald-950/40 text-emerald-300 shadow-sm'
                      : 'border-zinc-800/80 bg-zinc-950/60 hover:border-zinc-700 text-zinc-300'
                  }`}
                >
                  <div className="font-bold text-xs text-emerald-400">Caseta de Guardia</div>
                  <div className="truncate text-zinc-400 text-[10px] font-mono">guard@asesamo.com</div>
                </button>

                <button
                  type="button"
                  onClick={() => handleQuickFill('contratista@mantenimientonorte.com')}
                  disabled={isSubmitting}
                  className={`p-2.5 rounded-xl border text-left transition-all ${
                    email === 'contratista@mantenimientonorte.com'
                      ? 'border-amber-500 bg-amber-950/40 text-amber-300 shadow-sm'
                      : 'border-zinc-800/80 bg-zinc-950/60 hover:border-zinc-700 text-zinc-300'
                  }`}
                >
                  <div className="font-bold text-xs text-amber-400">Admin Contratista</div>
                  <div className="truncate text-zinc-400 text-[10px] font-mono">contratista@...</div>
                </button>

                <button
                  type="button"
                  onClick={() => handleQuickFill('ehs@asesamo.com')}
                  disabled={isSubmitting}
                  className={`p-2.5 rounded-xl border text-left transition-all ${
                    email === 'ehs@asesamo.com'
                      ? 'border-blue-500 bg-blue-950/40 text-blue-300 shadow-sm'
                      : 'border-zinc-800/80 bg-zinc-950/60 hover:border-zinc-700 text-zinc-300'
                  }`}
                >
                  <div className="font-bold text-xs text-blue-400">Gerencia EHS</div>
                  <div className="truncate text-zinc-400 text-[10px] font-mono">ehs@asesamo.com</div>
                </button>

                <button
                  type="button"
                  onClick={() => handleQuickFill('admin@asesamo.com')}
                  disabled={isSubmitting}
                  className={`p-2.5 rounded-xl border text-left transition-all ${
                    email === 'admin@asesamo.com'
                      ? 'border-purple-500 bg-purple-950/40 text-purple-300 shadow-sm'
                      : 'border-zinc-800/80 bg-zinc-950/60 hover:border-zinc-700 text-zinc-300'
                  }`}
                >
                  <div className="font-bold text-xs text-purple-400">Admin de Planta</div>
                  <div className="truncate text-zinc-400 text-[10px] font-mono">admin@asesamo.com</div>
                </button>
              </div>
            </div>
          </div>

          <p className="text-center text-xs text-zinc-400 font-mono">
            ASESAMO • Planta Querétaro • Conforme a Normas Oficiales Mexicanas
          </p>
        </div>
      </div>

      {/* Lado Derecho: Artefacto Visual Editorial y Candados Inteligentes */}
      <div className="hidden lg:flex flex-1 relative bg-gradient-to-br from-zinc-900 via-zinc-950 to-black items-center justify-center p-12 border-l border-zinc-800/60 overflow-hidden">
        {/* Luces Ambientales de Fondo */}
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-emerald-500/10 rounded-full blur-3xl pointer-events-none" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-amber-500/10 rounded-full blur-3xl pointer-events-none" />

        <div className="w-full max-w-lg space-y-8 z-10">
          <div className="space-y-4">
            <span className="text-xs font-mono uppercase tracking-widest text-emerald-400">
              Protocolo de Vida o Muerte
            </span>
            <h2 className="text-3xl xl:text-4xl font-extrabold tracking-tight text-white leading-tight">
              Cero accidentes en planta mediante candados digitales en tiempo real.
            </h2>
            <p className="text-sm text-zinc-400 leading-relaxed">
              Validación automatizada de constancias DC-3, certificados médicos y permisos de alto riesgo antes de permitir el ingreso o la apertura de labores en campo.
            </p>
          </div>

          {/* Tarjeta Mockup Viva */}
          <div className="p-6 rounded-2xl bg-zinc-950/90 border border-zinc-800/80 shadow-2xl backdrop-blur-md space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Cpu className="w-4 h-4 text-emerald-400" />
                <span className="text-xs font-mono font-bold text-zinc-200">MOTOR CANDADO INTELIGENTE</span>
              </div>
              <span className="px-2 py-0.5 rounded text-[10px] font-bold font-mono bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
                ACTIVO
              </span>
            </div>

            <div className="space-y-2 text-xs font-mono">
              <div className="p-3 rounded-xl bg-zinc-900/80 border border-zinc-800 flex items-center justify-between">
                <span className="text-zinc-400">NOM-009 Trabajos en Altura:</span>
                <span className="text-emerald-400 font-bold">DC-3 VIGENTE</span>
              </div>
              <div className="p-3 rounded-xl bg-zinc-900/80 border border-zinc-800 flex items-center justify-between">
                <span className="text-zinc-400">NOM-033 Espacios Confinados:</span>
                <span className="text-emerald-400 font-bold">EXAMEN MÉDICO APTO</span>
              </div>
              <div className="p-3 rounded-xl bg-zinc-900/80 border border-zinc-800 flex items-center justify-between">
                <span className="text-zinc-400">Caseta de Control:</span>
                <span className="text-emerald-400 font-bold">ESCÁNER QR SINCRONIZADO</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-zinc-950 flex items-center justify-center text-zinc-400">
          <Loader2 className="w-8 h-8 animate-spin text-emerald-500" />
        </div>
      }
    >
      <LoginForm />
    </Suspense>
  );
}
