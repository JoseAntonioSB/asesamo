'use client';

import React, { useState, useEffect, useRef } from 'react';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/contexts/auth-context';
import { apiClient, ApiError } from '@/lib/api/client';
import { UserProfileDetail } from '@/lib/api/types';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import {
  User,
  ShieldCheck,
  Mail,
  Phone,
  Building2,
  HardHat,
  Camera,
  UploadCloud,
  Lock,
  Eye,
  EyeOff,
  CheckCircle2,
  AlertCircle,
  Loader2,
  ArrowLeft,
  Calendar,
  SunMoon,
  Sparkles,
  QrCode,
  FileText,
  Users,
  LogOut,
  Check,
} from 'lucide-react';

export default function PerfilPage() {
  const { user, logout, updateUserSession } = useAuth();
  const [profile, setProfile] = useState<UserProfileDetail | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [avatarTimestamp, setAvatarTimestamp] = useState<number>(Date.now());

  // Personal Info Form State
  const [fullName, setFullName] = useState<string>('');
  const [phone, setPhone] = useState<string>('');
  const [isSavingInfo, setIsSavingInfo] = useState<boolean>(false);

  // Password Form State
  const [currentPassword, setCurrentPassword] = useState<string>('');
  const [newPassword, setNewPassword] = useState<string>('');
  const [confirmPassword, setConfirmPassword] = useState<string>('');
  const [showCurrentPw, setShowCurrentPw] = useState<boolean>(false);
  const [showNewPw, setShowNewPw] = useState<boolean>(false);
  const [isSavingPassword, setIsSavingPassword] = useState<boolean>(false);

  // Avatar Upload State
  const [isUploadingAvatar, setIsUploadingAvatar] = useState<boolean>(false);
  const [avatarPreview, setAvatarPreview] = useState<string | null>(null);
  const [avatarLoadError, setAvatarLoadError] = useState<boolean>(false);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  // Active Tab State: 'info' | 'security' | 'preferences'
  const [activeTab, setActiveTab] = useState<'info' | 'security' | 'preferences'>('info');

  const fetchProfile = async () => {
    try {
      setIsLoading(true);
      const data = await apiClient.getProfile();
      setProfile(data);
      setFullName(data.full_name || '');
      setPhone(data.phone || '');
      setAvatarLoadError(false);
    } catch (err) {
      const apiErr = err as ApiError;
      toast.error(apiErr.message || 'No se pudo cargar el perfil del usuario');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchProfile();
  }, []);

  // Handle Photo Selection & Upload
  const handlePhotoSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate size (max 5 MB)
    if (file.size > 5 * 1024 * 1024) {
      toast.error('La fotografía excede el tamaño máximo permitido de 5 MB.');
      return;
    }

    // Validate mime type
    const validTypes = ['image/jpeg', 'image/png', 'image/webp'];
    if (!validTypes.includes(file.type)) {
      toast.error('Formato no válido. Solo se permiten imágenes JPEG, PNG o WebP.');
      return;
    }

    // Optimistic preview
    const previewUrl = URL.createObjectURL(file);
    setAvatarPreview(previewUrl);
    setAvatarLoadError(false);

    try {
      setIsUploadingAvatar(true);
      const res = await apiClient.uploadMyAvatar(file);
      toast.success('Fotografía de perfil actualizada con éxito.');
      setAvatarTimestamp(Date.now());
      updateUserSession({ avatar_r2_key: res.avatar_key });
      if (profile) {
        setProfile({ ...profile, avatar_r2_key: res.avatar_key });
      }
    } catch (err) {
      const apiErr = err as ApiError;
      toast.error(apiErr.message || 'Error al subir la fotografía.');
      setAvatarPreview(null);
    } finally {
      setIsUploadingAvatar(false);
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  };

  // Handle Profile Info Update
  const handleSaveInfo = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!fullName.trim()) {
      toast.error('El nombre completo es obligatorio.');
      return;
    }

    try {
      setIsSavingInfo(true);
      const updated = await apiClient.updateProfile({
        full_name: fullName.trim(),
        phone: phone.trim() || undefined,
      });
      setProfile(updated);
      updateUserSession({
        full_name: updated.full_name,
        phone: updated.phone,
      });
      toast.success('Información personal guardada con éxito.');
    } catch (err) {
      const apiErr = err as ApiError;
      toast.error(apiErr.message || 'Error al actualizar información.');
    } finally {
      setIsSavingInfo(false);
    }
  };

  // Handle Password Change
  const handleSavePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentPassword) {
      toast.error('Debe ingresar su contraseña actual.');
      return;
    }
    if (newPassword.length < 8) {
      toast.error('La nueva contraseña debe tener al menos 8 caracteres.');
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error('Las contraseñas no coinciden.');
      return;
    }

    try {
      setIsSavingPassword(true);
      await apiClient.updateProfile({
        full_name: fullName.trim() || profile?.full_name || '',
        current_password: currentPassword,
        new_password: newPassword,
      });
      toast.success('Contraseña actualizada correctamente.');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err) {
      const apiErr = err as ApiError;
      toast.error(apiErr.message || 'Error al cambiar contraseña.');
    } finally {
      setIsSavingPassword(false);
    }
  };

  const getRoleLabel = (role?: string) => {
    switch (role) {
      case 'guard':
        return 'Guardia de Seguridad';
      case 'plant_admin':
        return 'Administrador de Planta';
      case 'ehs_manager':
        return 'Gerente de EHS';
      case 'contractor_admin':
        return 'Administrador de Contratista';
      case 'worker':
        return 'Trabajador Operativo';
      case 'super_admin':
        return 'Super Administrador';
      default:
        return role || 'Usuario';
    }
  };

  const initials = profile?.full_name
    ? profile.full_name
        .split(' ')
        .filter(Boolean)
        .slice(0, 2)
        .map((n) => n[0])
        .join('')
        .toUpperCase()
    : user?.full_name
    ? user.full_name
        .split(' ')
        .filter(Boolean)
        .slice(0, 2)
        .map((n) => n[0])
        .join('')
        .toUpperCase()
    : 'US';

  const avatarSrc = avatarPreview || (profile?.id ? `${apiClient.getUserAvatarUrl(profile.id)}?t=${avatarTimestamp}` : null);

  return (
    <ProtectedRoute>
      <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col font-sans transition-colors duration-200">
        {/* Top Navbar */}
        <header className="sticky top-0 z-40 border-b border-slate-800 bg-slate-900/90 backdrop-blur-md px-4 sm:px-8 py-3.5 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Link
              href="/"
              className="inline-flex items-center gap-2 text-sm text-slate-400 hover:text-white transition-colors py-1 px-2.5 rounded-lg hover:bg-slate-800"
            >
              <ArrowLeft className="w-4 h-4" />
              <span className="font-medium">Inicio</span>
            </Link>
            <div className="h-4 w-px bg-slate-800 hidden sm:block" />
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg bg-emerald-600 text-white flex items-center justify-center font-bold text-xs shadow-md shadow-emerald-600/30">
                <ShieldCheck className="w-4 h-4" />
              </div>
              <div>
                <h1 className="text-sm font-bold text-white tracking-wide leading-none">Mi Perfil</h1>
                <p className="text-[11px] text-slate-400 font-mono mt-0.5">ASESAMO Access Control</p>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2 sm:gap-3">
            {/* Context Navigation Shortcuts based on role */}
            {profile?.role === 'guard' && (
              <Link href="/caseta">
                <Button variant="outline" size="sm" className="border-slate-700 text-xs text-slate-300 hover:text-white h-8">
                  Caseta
                </Button>
              </Link>
            )}
            {profile?.role === 'contractor_admin' && (
              <Link href="/contratistas">
                <Button variant="outline" size="sm" className="border-slate-700 text-xs text-slate-300 hover:text-white h-8">
                  Cuadrilla
                </Button>
              </Link>
            )}
            {profile?.role === 'worker' && (
              <Link href="/credencial">
                <Button variant="outline" size="sm" className="border-slate-700 text-xs text-slate-300 hover:text-white h-8">
                  Credencial
                </Button>
              </Link>
            )}

            <ThemeToggle />

            <Button
              variant="outline"
              size="sm"
              onClick={logout}
              className="border-slate-800 text-slate-400 hover:text-rose-400 hover:bg-rose-950/20 hover:border-rose-900/60 text-xs h-8 gap-1.5"
              title="Cerrar sesión"
            >
              <LogOut className="w-3.5 h-3.5" />
              <span className="hidden sm:inline">Salir</span>
            </Button>
          </div>
        </header>

        {/* Main Content Area */}
        <main className="flex-1 max-w-5xl w-full mx-auto px-4 sm:px-6 py-6 sm:py-10 space-y-8">
          {isLoading ? (
            <div className="flex flex-col items-center justify-center py-20 text-slate-400 space-y-3">
              <Loader2 className="w-8 h-8 animate-spin text-emerald-500" />
              <p className="text-sm font-medium">Cargando perfil de usuario...</p>
            </div>
          ) : (
            <>
              {/* Profile Hero Card */}
              <div className="relative rounded-3xl border border-slate-800 bg-gradient-to-b from-slate-900 via-slate-900/95 to-slate-950 p-6 sm:p-8 shadow-2xl overflow-hidden">
                <div className="absolute top-0 right-0 w-96 h-96 bg-emerald-500/5 rounded-full blur-3xl pointer-events-none" />

                <div className="flex flex-col sm:flex-row items-center sm:items-start gap-6 sm:gap-8">
                  {/* Avatar with Upload Trigger */}
                  <div className="relative group shrink-0">
                    <div className="w-28 h-28 sm:w-32 sm:h-32 rounded-3xl border-4 border-slate-800 bg-slate-800 text-white flex items-center justify-center font-black text-3xl shadow-xl overflow-hidden relative">
                      {avatarSrc && !avatarLoadError ? (
                        <img
                          src={avatarSrc}
                          alt={profile?.full_name || 'Avatar'}
                          className="w-full h-full object-cover"
                          onError={() => setAvatarLoadError(true)}
                        />
                      ) : (
                        <div className="w-full h-full flex items-center justify-center bg-gradient-to-tr from-slate-800 to-slate-700 text-slate-200">
                          {initials}
                        </div>
                      )}

                      {/* Loading spinner overlay */}
                      {isUploadingAvatar && (
                        <div className="absolute inset-0 bg-slate-950/70 backdrop-blur-xs flex items-center justify-center">
                          <Loader2 className="w-7 h-7 text-emerald-400 animate-spin" />
                        </div>
                      )}
                    </div>

                    {/* Camera Button to trigger file picker */}
                    <button
                      type="button"
                      onClick={() => fileInputRef.current?.click()}
                      disabled={isUploadingAvatar}
                      className="absolute -bottom-2 -right-2 p-2.5 rounded-2xl bg-emerald-600 hover:bg-emerald-500 text-white shadow-lg border-2 border-slate-900 transition-all hover:scale-105 active:scale-95 cursor-pointer disabled:opacity-50"
                      title="Cambiar fotografía"
                    >
                      <Camera className="w-4 h-4" />
                    </button>
                    <input
                      ref={fileInputRef}
                      type="file"
                      accept="image/jpeg,image/png,image/webp"
                      className="hidden"
                      onChange={handlePhotoSelect}
                    />
                  </div>

                  {/* User Details */}
                  <div className="flex-1 text-center sm:text-left space-y-3">
                    <div className="flex flex-wrap items-center justify-center sm:justify-start gap-2">
                      <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                        <ShieldCheck className="w-3.5 h-3.5" />
                        {getRoleLabel(profile?.role)}
                      </span>
                      {profile?.is_active && (
                        <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-950/60 text-emerald-300 border border-emerald-800/60">
                          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                          Cuenta Activa
                        </span>
                      )}
                    </div>

                    <h2 className="text-2xl sm:text-3xl font-black text-white tracking-tight">
                      {profile?.full_name}
                    </h2>

                    <div className="flex flex-wrap items-center justify-center sm:justify-start gap-y-2 gap-x-4 text-xs text-slate-400">
                      <div className="flex items-center gap-1.5">
                        <Mail className="w-3.5 h-3.5 text-slate-500" />
                        <span>{profile?.email}</span>
                      </div>
                      {profile?.phone && (
                        <div className="flex items-center gap-1.5">
                          <Phone className="w-3.5 h-3.5 text-slate-500" />
                          <span>{profile.phone}</span>
                        </div>
                      )}
                      {profile?.contractor_name && (
                        <div className="flex items-center gap-1.5 text-cyan-400 font-medium">
                          <Building2 className="w-3.5 h-3.5" />
                          <span>{profile.contractor_name}</span>
                        </div>
                      )}
                    </div>

                    <div className="pt-2 text-[11px] text-slate-500 font-mono">
                      Formato recomendado para fotografía: Cuadrada (1:1), JPEG o PNG, máx. 5 MB.
                    </div>
                  </div>
                </div>
              </div>

              {/* Navigation Tabs */}
              <div className="flex items-center gap-2 border-b border-slate-800 pb-px">
                <button
                  type="button"
                  onClick={() => setActiveTab('info')}
                  className={`px-4 py-2.5 text-sm font-semibold rounded-t-xl transition-all border-b-2 flex items-center gap-2 ${
                    activeTab === 'info'
                      ? 'text-emerald-400 border-emerald-500 bg-slate-900/60'
                      : 'text-slate-400 border-transparent hover:text-slate-200 hover:bg-slate-900/30'
                  }`}
                >
                  <User className="w-4 h-4" />
                  <span>Información Personal</span>
                </button>
                <button
                  type="button"
                  onClick={() => setActiveTab('security')}
                  className={`px-4 py-2.5 text-sm font-semibold rounded-t-xl transition-all border-b-2 flex items-center gap-2 ${
                    activeTab === 'security'
                      ? 'text-emerald-400 border-emerald-500 bg-slate-900/60'
                      : 'text-slate-400 border-transparent hover:text-slate-200 hover:bg-slate-900/30'
                  }`}
                >
                  <Lock className="w-4 h-4" />
                  <span>Seguridad y Contraseña</span>
                </button>
                <button
                  type="button"
                  onClick={() => setActiveTab('preferences')}
                  className={`px-4 py-2.5 text-sm font-semibold rounded-t-xl transition-all border-b-2 flex items-center gap-2 ${
                    activeTab === 'preferences'
                      ? 'text-emerald-400 border-emerald-500 bg-slate-900/60'
                      : 'text-slate-400 border-transparent hover:text-slate-200 hover:bg-slate-900/30'
                  }`}
                >
                  <SunMoon className="w-4 h-4" />
                  <span>Preferencias</span>
                </button>
              </div>

              {/* Tab 1: Personal Info Form */}
              {activeTab === 'info' && (
                <div className="rounded-2xl border border-slate-800 bg-slate-900/80 p-6 sm:p-8 space-y-6">
                  <div>
                    <h3 className="text-lg font-bold text-white">Datos de Perfil</h3>
                    <p className="text-xs text-slate-400 mt-1">
                      Actualice su nombre y número telefónico de contacto registrado en la plataforma.
                    </p>
                  </div>

                  <form onSubmit={handleSaveInfo} className="space-y-5">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                      <div className="space-y-1.5">
                        <label className="text-xs font-semibold text-slate-300">
                          Nombre Completo <span className="text-rose-400">*</span>
                        </label>
                        <Input
                          type="text"
                          value={fullName}
                          onChange={(e) => setFullName(e.target.value)}
                          placeholder="Nombre y Apellidos"
                          required
                          className="bg-slate-950 border-slate-800 text-white focus:border-emerald-500"
                        />
                      </div>

                      <div className="space-y-1.5">
                        <label className="text-xs font-semibold text-slate-300">
                          Teléfono de Contacto
                        </label>
                        <Input
                          type="tel"
                          value={phone}
                          onChange={(e) => setPhone(e.target.value)}
                          placeholder="Ej. 442 123 4567"
                          className="bg-slate-950 border-slate-800 text-white focus:border-emerald-500"
                        />
                      </div>

                      <div className="space-y-1.5">
                        <label className="text-xs font-semibold text-slate-400">
                          Correo Electrónico (Solo Lectura)
                        </label>
                        <Input
                          type="email"
                          value={profile?.email || ''}
                          disabled
                          className="bg-slate-950/60 border-slate-800/80 text-slate-400 cursor-not-allowed"
                        />
                      </div>

                      <div className="space-y-1.5">
                        <label className="text-xs font-semibold text-slate-400">
                          Rol en el Sistema
                        </label>
                        <Input
                          type="text"
                          value={getRoleLabel(profile?.role)}
                          disabled
                          className="bg-slate-950/60 border-slate-800/80 text-slate-400 cursor-not-allowed"
                        />
                      </div>
                    </div>

                    {profile?.contractor_name && (
                      <div className="p-4 rounded-xl bg-cyan-950/20 border border-cyan-800/30 flex items-center gap-3">
                        <Building2 className="w-5 h-5 text-cyan-400 shrink-0" />
                        <div>
                          <div className="text-xs font-bold text-cyan-300">Empresa Contratista Afiliada</div>
                          <div className="text-sm font-semibold text-white">{profile.contractor_name}</div>
                        </div>
                      </div>
                    )}

                    <div className="flex justify-end pt-3">
                      <Button
                        type="submit"
                        disabled={isSavingInfo}
                        className="bg-emerald-600 hover:bg-emerald-500 text-white font-semibold text-sm px-6 h-10 shadow-lg shadow-emerald-600/20 gap-2"
                      >
                        {isSavingInfo ? (
                          <>
                            <Loader2 className="w-4 h-4 animate-spin" />
                            <span>Guardando...</span>
                          </>
                        ) : (
                          <>
                            <Check className="w-4 h-4" />
                            <span>Guardar Cambios</span>
                          </>
                        )}
                      </Button>
                    </div>
                  </form>
                </div>
              )}

              {/* Tab 2: Security & Password */}
              {activeTab === 'security' && (
                <div className="rounded-2xl border border-slate-800 bg-slate-900/80 p-6 sm:p-8 space-y-6">
                  <div>
                    <h3 className="text-lg font-bold text-white">Actualizar Contraseña</h3>
                    <p className="text-xs text-slate-400 mt-1">
                      Para cambiar su clave de acceso, confirme su contraseña actual y defina una nueva (mínimo 8 caracteres).
                    </p>
                  </div>

                  <form onSubmit={handleSavePassword} className="space-y-5 max-w-xl">
                    <div className="space-y-1.5">
                      <label className="text-xs font-semibold text-slate-300">
                        Contraseña Actual <span className="text-rose-400">*</span>
                      </label>
                      <div className="relative">
                        <Input
                          type={showCurrentPw ? 'text' : 'password'}
                          value={currentPassword}
                          onChange={(e) => setCurrentPassword(e.target.value)}
                          placeholder="••••••••"
                          required
                          className="bg-slate-950 border-slate-800 text-white focus:border-emerald-500 pr-10"
                        />
                        <button
                          type="button"
                          onClick={() => setShowCurrentPw(!showCurrentPw)}
                          className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
                        >
                          {showCurrentPw ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                        </button>
                      </div>
                    </div>

                    <div className="space-y-1.5">
                      <label className="text-xs font-semibold text-slate-300">
                        Nueva Contraseña <span className="text-rose-400">*</span>
                      </label>
                      <div className="relative">
                        <Input
                          type={showNewPw ? 'text' : 'password'}
                          value={newPassword}
                          onChange={(e) => setNewPassword(e.target.value)}
                          placeholder="Mínimo 8 caracteres"
                          required
                          minLength={8}
                          className="bg-slate-950 border-slate-800 text-white focus:border-emerald-500 pr-10"
                        />
                        <button
                          type="button"
                          onClick={() => setShowNewPw(!showNewPw)}
                          className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
                        >
                          {showNewPw ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                        </button>
                      </div>
                    </div>

                    <div className="space-y-1.5">
                      <label className="text-xs font-semibold text-slate-300">
                        Confirmar Nueva Contraseña <span className="text-rose-400">*</span>
                      </label>
                      <Input
                        type={showNewPw ? 'text' : 'password'}
                        value={confirmPassword}
                        onChange={(e) => setConfirmPassword(e.target.value)}
                        placeholder="Repita la nueva contraseña"
                        required
                        minLength={8}
                        className="bg-slate-950 border-slate-800 text-white focus:border-emerald-500"
                      />
                    </div>

                    <div className="pt-2 flex justify-end">
                      <Button
                        type="submit"
                        disabled={isSavingPassword}
                        className="bg-emerald-600 hover:bg-emerald-500 text-white font-semibold text-sm px-6 h-10 shadow-lg shadow-emerald-600/20 gap-2"
                      >
                        {isSavingPassword ? (
                          <>
                            <Loader2 className="w-4 h-4 animate-spin" />
                            <span>Actualizando...</span>
                          </>
                        ) : (
                          <>
                            <Lock className="w-4 h-4" />
                            <span>Modificar Contraseña</span>
                          </>
                        )}
                      </Button>
                    </div>
                  </form>
                </div>
              )}

              {/* Tab 3: System Preferences */}
              {activeTab === 'preferences' && (
                <div className="rounded-2xl border border-slate-800 bg-slate-900/80 p-6 sm:p-8 space-y-6">
                  <div>
                    <h3 className="text-lg font-bold text-white">Tema y Apariencia Visual</h3>
                    <p className="text-xs text-slate-400 mt-1">
                      Personalice el aspecto de la plataforma en su navegador. La preferencia se sincroniza en su cuenta.
                    </p>
                  </div>

                  <div className="flex items-center justify-between p-4 rounded-xl border border-slate-800 bg-slate-950 max-w-xl">
                    <div className="space-y-0.5">
                      <div className="text-sm font-semibold text-white">Alternar Modo Oscuro / Claro</div>
                      <div className="text-xs text-slate-400">
                        Cambie entre el modo de alto contraste oscuro y el modo diurno.
                      </div>
                    </div>
                    <ThemeToggle />
                  </div>
                </div>
              )}
            </>
          )}
        </main>
      </div>
    </ProtectedRoute>
  );
}
