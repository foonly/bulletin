import { defineStore } from "pinia";

export const useToastStore = defineStore("toast", {
	state: () => ({
		toasts: [],
	}),
	actions: {
		addToast(message, type = "info", duration = 15000) {
			const id = Date.now();
			this.toasts.push({ id, message, type, duration });

			if (duration > 0) {
				setTimeout(() => {
					this.removeToast(id);
				}, duration);
			}
		},
		removeToast(id) {
			this.toasts = this.toasts.filter((t) => t.id !== id);
		},
		success(message, duration = 15000) {
			this.addToast(message, "success", duration);
		},
		error(message, duration = 20000) {
			this.addToast(message, "error", duration);
		},
		info(message, duration = 15000) {
			this.addToast(message, "info", duration);
		},
		warning(message, duration = 20000) {
			this.addToast(message, "warning", duration);
		},
	},
});
