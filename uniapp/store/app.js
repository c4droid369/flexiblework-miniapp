import { defineStore } from 'pinia'
import { categoryApi } from '@/api/category'

// App-wide read-mostly data — currently just the cached category list so the
// home page can render the category strip without an extra round trip.
export const useAppStore = defineStore('app', {
	state: () => ({
		categories: [],
		loadedAt: 0,
	}),
	getters: {
		activeCategories: (s) => s.categories.filter(c => c.status === 1),
	},
	actions: {
		async bootstrap() {
			// Lazy-load on first call from a page; nothing to do here yet.
		},
		async loadCategories(force = false) {
			// Cache for 5 minutes; pass `true` to bypass.
			if (!force && Date.now() - this.loadedAt < 5 * 60 * 1000 && this.categories.length) {
				return this.categories
			}
			try {
				const list = await categoryApi.list()
				this.categories = Array.isArray(list) ? list : []
				this.loadedAt = Date.now()
			} catch (e) {
				// Silent — page-level UI will show empty state.
			}
			return this.categories
		},
	},
})
