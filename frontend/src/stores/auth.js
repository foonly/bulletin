import { defineStore } from "pinia";
import axios from "../api";

export const useAuthStore = defineStore("auth", {
	state: () => ({
		user: null,
		loading: false,
		error: null,
		notificationsEnabled:
			localStorage.getItem("notifications_enabled") === "true",
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
				const res = await axios.post("/api/auth/login", { username, password });
				if (res.data.status === "mfa_required") {
					return { mfaRequired: true };
				}
				await this.fetchMe();
				return { mfaRequired: false };
			} catch (err) {
				this.error = err.response?.data || "Login failed";
				throw err;
			} finally {
				this.loading = false;
			}
		},
		async loginTOTP(code) {
			this.loading = true;
			this.error = null;
			try {
				await axios.post("/api/auth/login-totp", { code });
				await this.fetchMe();
			} catch (err) {
				this.error = err.response?.data || "MFA failed";
				throw err;
			} finally {
				this.loading = false;
			}
		},
		async register(username, email, password, inviteCode) {
			this.loading = true;
			this.error = null;
			try {
				await axios.post("/api/auth/register", {
					username,
					email,
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
		async requestReset(email) {
			await axios.post("/api/auth/request-reset", { email });
		},
		async resetPassword(token, password) {
			await axios.post("/api/auth/reset-password", { token, password });
		},
		async requestVerification() {
			await axios.post("/api/auth/request-verification");
		},
		async verifyEmail(token) {
			await axios.post("/api/auth/verify-email", { token });
			await this.fetchMe();
		},
		async setupTOTP() {
			const res = await axios.post("/api/auth/totp/setup");
			return res.data;
		},
		async enableTOTP(code) {
			await axios.post("/api/auth/totp/enable", { code });
			await this.fetchMe();
		},
		async disableTOTP(password) {
			await axios.post("/api/auth/totp/disable", { password });
			await this.fetchMe();
		},
		setNotificationsEnabled(enabled) {
			this.notificationsEnabled = enabled;
			localStorage.setItem("notifications_enabled", enabled ? "true" : "false");
			if (enabled && "Notification" in window) {
				if (
					Notification.permission !== "granted" &&
					Notification.permission !== "denied"
				) {
					Notification.requestPermission();
				}
			}
		},
	},
});
