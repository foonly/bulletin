<template>
	<div class="auth-page">
		<div class="auth-card">
			<h1>{{ siteName }} Login</h1>

			<template v-if="!mfaRequired">
				<form @submit.prevent="handleLogin">
					<div class="field">
						<label for="login-username">Username</label>
						<input
							id="login-username"
							v-model="username"
							type="text"
							required
						/>
					</div>
					<div class="field">
						<label for="login-password">Password</label>
						<input
							id="login-password"
							v-model="password"
							type="password"
							required
						/>
					</div>
					<button
						type="submit"
						class="btn btn-primary btn-full"
						:disabled="auth.loading"
						style="margin-top: 1.5rem"
					>
						Login
					</button>
				</form>
			</template>

			<template v-else>
				<div style="text-align: center; margin-bottom: 1.5rem">
					<div style="font-size: 2.5rem; margin-bottom: 0.5rem">🔐</div>
					<p style="color: var(--fg-2)">Enter your 6-digit MFA code</p>
				</div>
				<form @submit.prevent="handleLoginTOTP">
					<div class="field">
						<input
							v-model="mfaCode"
							type="text"
							maxlength="6"
							placeholder="000000"
							class="input-code"
							required
							autofocus
						/>
					</div>
					<button
						type="submit"
						class="btn btn-primary btn-full"
						:disabled="auth.loading"
						style="margin-top: 1rem"
					>
						Verify Code
					</button>
					<button
						type="button"
						class="btn btn-ghost btn-full"
						@click="mfaRequired = false"
						style="margin-top: 0.5rem"
					>
						Back to Login
					</button>
				</form>
			</template>

			<div class="auth-footer">
				<p>
					Don't have an account?
					<router-link to="/register">Register with invite</router-link>
				</p>
				<p>
					<router-link to="/request-reset" style="color: var(--fg-3)"
						>Forgot password?</router-link
					>
				</p>
				<p class="app-version" style="margin-top: 0.5rem">v{{ version }}</p>
			</div>
		</div>
	</div>
</template>

<script setup>
import { ref } from "vue";
import pkg from "../../package.json";
import { useAuthStore } from "../stores/auth";
import { useToastStore } from "../stores/toast";
import { useRouter } from "vue-router";

const auth = useAuthStore();
const toast = useToastStore();
const router = useRouter();
const version = pkg.version;
const siteName = import.meta.env.VITE_SITE_NAME || "Bulletin";
const username = ref("");
const password = ref("");
const mfaRequired = ref(false);
const mfaCode = ref("");

const handleLogin = async () => {
	try {
		const result = await auth.login(username.value, password.value);
		if (result.mfaRequired) {
			mfaRequired.value = true;
			return;
		}
		toast.success("Welcome back!");
		router.push("/");
	} catch (err) {
		toast.error(err.response?.data || "Login failed");
	}
};

const handleLoginTOTP = async () => {
	try {
		await auth.loginTOTP(mfaCode.value);
		toast.success("MFA Verified. Welcome back!");
		router.push("/");
	} catch (err) {
		toast.error(err.response?.data || "Invalid MFA code");
	}
};
</script>
