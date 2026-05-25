<template>
	<div class="min-h-screen flex items-center justify-center bg-gray-950">
		<div class="bg-gray-800 p-8 rounded-lg shadow-xl w-96">
			<h1 class="text-2xl font-bold mb-6 text-center">{{ siteName }} Login</h1>

			<template v-if="!mfaRequired">
				<form @submit.prevent="handleLogin">
					<div class="mb-4">
						<label class="block text-sm font-medium mb-1">Username</label>
						<input
							v-model="username"
							type="text"
							class="w-full p-2 rounded bg-gray-700 border border-gray-600 focus:outline-none focus:border-purple-500"
							required
						/>
					</div>
					<div class="mb-6">
						<label class="block text-sm font-medium mb-1">Password</label>
						<input
							v-model="password"
							type="password"
							class="w-full p-2 rounded bg-gray-700 border border-gray-600 focus:outline-none focus:border-purple-500"
							required
						/>
					</div>
					<button
						type="submit"
						:disabled="auth.loading"
						class="w-full bg-purple-600 hover:bg-purple-700 p-2 rounded font-bold transition disabled:opacity-50"
					>
						Login
					</button>
				</form>
			</template>

			<template v-else>
				<div class="text-center mb-6">
					<div class="text-4xl mb-2">🔐</div>
					<p class="text-gray-300">Enter your 6-digit MFA code</p>
				</div>
				<form @submit.prevent="handleLoginTOTP">
					<div class="mb-6">
						<input
							v-model="mfaCode"
							type="text"
							maxlength="6"
							placeholder="000000"
							class="w-full p-3 text-center text-2xl tracking-[0.5em] rounded bg-gray-700 border border-gray-600 focus:outline-none focus:border-purple-500"
							required
							autofocus
						/>
					</div>
					<button
						type="submit"
						:disabled="auth.loading"
						class="w-full bg-purple-600 hover:bg-purple-700 p-2 rounded font-bold transition disabled:opacity-50"
					>
						Verify Code
					</button>
					<button
						type="button"
						@click="mfaRequired = false"
						class="w-full mt-2 text-sm text-gray-400 hover:text-gray-300"
					>
						Back to Login
					</button>
				</form>
			</template>

			<div class="mt-4 text-center text-sm space-y-2">
				<p>
					Don't have an account?
					<router-link to="/register" class="text-purple-400 hover:underline"
						>Register with invite</router-link
					>
				</p>
				<p>
					<router-link
						to="/request-reset"
						class="text-gray-500 hover:text-gray-400"
						>Forgot password?</router-link
					>
				</p>
			</div>
		</div>
	</div>
</template>

<script setup>
import { ref } from "vue";
import { useAuthStore } from "../stores/auth";
import { useToastStore } from "../stores/toast";
import { useRouter } from "vue-router";

const auth = useAuthStore();
const toast = useToastStore();
const router = useRouter();
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
