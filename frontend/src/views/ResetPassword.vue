<template>
	<div class="auth-page">
		<div class="auth-card">
			<h1>New Password</h1>
			<form v-if="!submitted" @submit.prevent="handleSubmit">
				<div class="field">
					<label for="new-password">New Password</label>
					<input
						id="new-password"
						v-model="password"
						type="password"
						required
						minlength="8"
					/>
				</div>
				<button type="submit" class="btn btn-primary btn-full" style="margin-top: 1.5rem">
					Update Password
				</button>
			</form>
			<div v-else style="text-align: center; display: flex; flex-direction: column; gap: 1rem">
				<p style="color: var(--fg-2)">Your password has been updated successfully.</p>
				<router-link to="/login" class="btn btn-primary">Go to Login</router-link>
			</div>
		</div>
	</div>
</template>

<script setup>
import { ref } from "vue";
import { useAuthStore } from "../stores/auth";
import { useToastStore } from "../stores/toast";
import { useRoute } from "vue-router";

const auth = useAuthStore();
const toast = useToastStore();
const route = useRoute();
const password = ref("");
const submitted = ref(false);

const handleSubmit = async () => {
	const token = route.query.token;
	if (!token) {
		toast.error("Invalid reset link");
		return;
	}

	try {
		await auth.resetPassword(token, password.value);
		submitted.value = true;
		toast.success("Password updated!");
	} catch (err) {
		toast.error(err.response?.data || "Failed to update password");
	}
};
</script>
