<template>
	<div class="auth-page">
		<div class="auth-card">
			<h1>Reset Password</h1>
			<form v-if="!submitted" @submit.prevent="handleSubmit">
				<div class="field">
					<label for="reset-email">Email Address</label>
					<input id="reset-email" v-model="email" type="email" required />
				</div>
				<button type="submit" class="btn btn-primary btn-full" style="margin-top: 1.5rem">
					Send Reset Link
				</button>
			</form>
			<div v-else style="text-align: center; display: flex; flex-direction: column; gap: 0.75rem">
				<p style="color: var(--fg-2)">
					If an account with that email exists, we've sent a password reset link.
				</p>
				<p style="font-size: var(--text-sm); color: var(--fg-3)">
					Be sure to check your spam folder if you don't see it within a few minutes.
				</p>
			</div>
			<div class="auth-footer">
				<p><router-link to="/login">Back to Login</router-link></p>
			</div>
		</div>
	</div>
</template>

<script setup>
import { ref } from "vue";
import { useAuthStore } from "../stores/auth";
import { useToastStore } from "../stores/toast";

const auth = useAuthStore();
const toast = useToastStore();
const email = ref("");
const submitted = ref(false);

const handleSubmit = async () => {
	try {
		await auth.requestReset(email.value);
		submitted.value = true;
	} catch (err) {
		toast.error("Failed to request password reset");
	}
};
</script>
