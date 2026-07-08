import { redirect } from "@sveltejs/kit";
import type { Handle } from "@sveltejs/kit";
import { userStore } from "$lib/stores/user";

const publicRoutes = ['/login', '/register', '/demo'];

export const handle: Handle = async ({ event, resolve }) => {
    // 1. Check authentication status via cookie
    const sessionCookie = event.cookies.get('nextup_session');
    event.locals.isAuthenticated = !!sessionCookie;

    const { pathname } = event.url;
    
    // Allow public routes. For '/demo' we might want to allow subpaths, for login/register exact match or subpaths is fine.
    const isPublicRoute = publicRoutes.some(route => pathname.startsWith(route));

    // 2. Redirect unauthenticated users away from protected routes
    if (!event.locals.isAuthenticated && !isPublicRoute) {
        throw redirect(303, '/login');
    }

    // 3. Redirect authenticated users away from public auth routes and the root
    if (event.locals.isAuthenticated && (pathname === '/' || pathname === '/login' || pathname === '/register')) {
        throw redirect(303, '/dashboard');
    }

    return resolve(event);
};