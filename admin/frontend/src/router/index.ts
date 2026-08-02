import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import type { MenuTree, MeResp } from '@/api/auth';
import { useAuthStore } from '@/stores/auth';
import { getAccessToken } from '@/utils/auth';

// import.meta.glob makes Vite resolve every matching .vue file at build time
// and emit one chunk per file. The returned map's keys are absolute paths
// like "/src/views/system/user/index.vue" — we use them as a lookup table so
// the dynamic router.addRoute never hands the browser a "@/..." alias it
// can't resolve at runtime.
const viewModules = import.meta.glob('@/views/**/*.vue');

// Static, always-mounted routes. /dashboard and /profile are children of the
// root layout — they don't come from /auth/me. Dynamic routes from /auth/me
// are appended as additional children of the `root` named route below.
const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/login/index.vue'),
    meta: { public: true, title: 'Login' },
  },
  {
    path: '/',
    name: 'root',
    component: () => import('@/layouts/DefaultLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { title: 'Dashboard', icon: 'dashboard' },
      },
      {
        path: 'profile',
        name: 'profile',
        component: () => import('@/views/profile/index.vue'),
        meta: { title: 'Profile', icon: 'user' },
      },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

// Track which dynamic routes we've already added so we don't double-register.
let dynamicLoaded = false;

// buildDynamicRoutes walks the menu tree from /auth/me and produces route
// configs. Every entry is added as a child of the `root` named route — this
// is why paths are RELATIVE (no leading slash). vue-router's addRoute rejects
// absolute paths missing the leading slash with "Invalid path".
//
// Dashboard is skipped here because it is already a static child of root.
function buildDynamicRoutes(menus: MenuTree[]): RouteRecordRaw[] {
  const out: RouteRecordRaw[] = [];
  const walk = (nodes: MenuTree[], parentPath: string) => {
    for (const n of nodes) {
      if (n.type === 2 && n.component && n.path && n.path !== '/dashboard') {
        const fullPath = n.path.startsWith('/') ? n.path : `${parentPath}/${n.path}`;
        const relative = fullPath.replace(/^\/+/, '');

        // Resolve the view file via the glob lookup. n.component looks like
        // "views/system/user/index.vue" (seed value); the glob keys are
        // "/src/views/system/user/index.vue".
        const viewKey = `/src/${n.component.replace(/^views\//, 'views/')}`;
        const loader = viewModules[viewKey];
        if (!loader) {
          // View file missing — log and skip instead of crashing the app.
          console.warn(`[router] view not found: ${viewKey} (menu ${n.name})`);
          continue;
        }
        out.push({
          path: relative,
          name: `m_${n.id}`,
          component: loader,
          meta: { title: n.title, perm: n.perm_code, icon: n.icon },
        });
      }
      if (n.children?.length) walk(n.children, n.path || parentPath);
    }
  };
  walk(menus, '');
  return out;
}

router.beforeEach(async (to) => {
  const auth = useAuthStore();
  const isPublic = to.meta.public === true;

  if (isPublic) {
    if (to.name === 'login' && auth.isAuthenticated) return { path: '/' };
    return true;
  }

  if (!getAccessToken()) return { path: '/login', query: { redirect: to.fullPath } };

  // Lazy-load the menu tree on first navigation, then cache it.
  if (!dynamicLoaded && auth.profile === null) {
    try {
      await auth.fetchMe();
    } catch {
      return { path: '/login', query: { redirect: to.fullPath } };
    }
    const me = auth.profile as MeResp | null;
    const dynamic = buildDynamicRoutes(me?.menus ?? []);
    for (const r of dynamic) router.addRoute('root', r as RouteRecordRaw);
    dynamicLoaded = true;
    // If we landed here via the wildcard redirect (deep-link like
    // /system/user before dynamic routes existed), point at the original
    // path so the now-registered dynamic route can take over. Otherwise
    // just re-resolve the current target.
    const target: string = to.redirectedFrom?.fullPath ?? to.fullPath;
    return { path: target, replace: true };
  }

  // Permission gate on per-route meta.perm codes.
  const required = to.meta.perm as string | undefined;
  if (required && !auth.hasPerm(required)) {
    return { path: '/dashboard' };
  }
  return true;
});

export default router;
