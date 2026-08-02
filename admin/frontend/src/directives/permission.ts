import type { Directive } from 'vue';
import { useAuthStore } from '@/stores/auth';

// v-permission="'user:create'" — removes the element if the user lacks the
// perm. v-permission="['user:create', 'user:update']" is OR semantics.
export const permissionDirective: Directive<HTMLElement, string | string[]> = {
  mounted(el, binding) {
    const auth = useAuthStore();
    const required = Array.isArray(binding.value) ? binding.value : [binding.value];
    const allowed = required.some((c) => auth.hasPerm(c));
    if (!allowed) el.parentNode?.removeChild(el);
  },
};
