import { redirect } from 'next/navigation';
import { cookies } from 'next/headers';
import { TOKEN_COOKIE_NAME, getRoleFromJwt, getDefaultHomeForRole } from '@/lib/auth';

export default function HomePage() {
  const cookieStore = cookies();
  const token = cookieStore.get(TOKEN_COOKIE_NAME)?.value;

  if (token) {
    const role = getRoleFromJwt(token);
    redirect(getDefaultHomeForRole(role));
  } else {
    redirect('/login');
  }
}
