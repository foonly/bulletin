<template>
	<div class="auth-page">
		<div class="auth-card">
			<h1>Join {{ siteName }}</h1>
			<form @submit.prevent="handleRegister">
				<div class="field">
					<label for="reg-invite">Invite Code</label>
					<input id="reg-invite" v-model="inviteCode" type="text" required />
				</div>
				<div class="field">
					<label for="reg-username">Username</label>
					<input id="reg-username" v-model="username" type="text" required />
				</div>
				<div class="field">
					<label for="reg-email">Email Address</label>
					<input id="reg-email" v-model="email" type="email" required />
				</div>
				<div class="field">
					<label for="reg-password">Password</label>
					<input id="reg-password" v-model="password" type="password" required />
				</div>
				<button type="submit" class="btn btn-primary btn-full" style="margin-top: 1.5rem">
					Register
				</button>
			</form>
			<div class="auth-footer">
				<p>
					Already have an account?
					<router-link to="/login">Login</router-link>
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
const inviteCode = ref("");
const username = ref("");
const email = ref("");
const password = ref("");

const handleRegister = async () => {
	try {
		await auth.register(
			username.value,
			email.value,
			password.value,
			inviteCode.value,
		);
		toast.success(`Account created! Welcome to ${siteName}.`);
		router.push("/");
	} catch (err) {
		toast.error(err.response?.data || "Registration failed");
	}
};
</script>
