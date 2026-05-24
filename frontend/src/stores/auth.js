import { defineStore } from "pinia";
import axios from "axios";

export const useAuthStore = defineStore("auth", {
	state: () => ({
		user: null,
		loading: false,
		error: null,
	}),
	actions: {
		async fetchMe() {
			try {
				const res = await axios.get("/api/auth/me");
				this.user = res.data;
			} catch (err) {
				this.user = null;
			}
		},
		async login(username, password) {
			this.loading = true;
			this.error = null;
			try {
				await axios.post("/api/auth/login", { username, password });
				await this.fetchMe();
			} catch (err) {
				this.error = err.response?.data || "Login failed";
				throw err;
			} finally {
				this.loading = false;
			}
		},
		async register(username, password, inviteCode) {
			this.loading = true;
			this.error = null;
			try {
				await axios.post("/api/auth/register", {
					username,
					password,
					invite_code: inviteCode,
				});
				await this.fetchMe();
			} catch (err) {
				this.error = err.response?.data || "Registration failed";
				throw err;
			} finally {
				this.loading = false;
			}
		},
		async logout() {
			await axios.post("/api/auth/logout");
			this.user = null;
		},
		async updateMe(userData) {
			await axios.put("/api/auth/me", userData);
			await this.fetchMe();
		},
	},
});
