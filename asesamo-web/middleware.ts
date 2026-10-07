/**
 * ASESAMO — Edge Route Protection Middleware
 * Location: asesamo-web/middleware.ts
 */

import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { TOKEN_COOKIE_NAME } from './lib/auth';

export function middleware(request: NextRequest) {
  const token = request.cookies.get(TOKEN_COOKIE_NAME)?.value;
  const { pathname } = request.nextUrl;

  const isAuthRoute = pathname === '/login' || pathname.startsWith('/login/');
  const isProtectedRoute =
    pathname === '/caseta' ||
    pathname.startsWith('/caseta/') ||
    pathname === '/dashboard' ||
    pathname.startsWith('/dashboard/');

  // 1. Root route '/' redirect
  if (pathname === '/') {
    if (token) {
      return NextResponse.redirect(new URL('/caseta', request.url));
    } else {
      return NextResponse.redirect(new URL('/login', request.url));
    }
  }

  // 2. Unauthenticated user trying to access protected gate/caseta routes
  if (isProtectedRoute && !token) {
    const loginUrl = new URL('/login', request.url);
    loginUrl.searchParams.set('redirect', pathname);
    return NextResponse.redirect(loginUrl);
  }

  // 3. Authenticated user visiting /login
  if (isAuthRoute && token) {
    return NextResponse.redirect(new URL('/caseta', request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except:
     * 1. /api/* (backend or internal API routes)
     * 2. /_next/static (static JS, CSS chunks)
     * 3. /_next/image (image optimization files)
     * 4. /favicon.ico and static assets (.svg, .png, .jpg)
     */
    '/((?!api|_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
};
