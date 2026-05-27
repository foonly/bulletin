<template>
	<div class="auth-page">
		<div class="auth-card" style="text-align: center">
			<h1>Email Verification</h1>
			<div v-if="loading" style="color: var(--fg-2)">
				Verifying your email address...
			</div>
			<div v-else-if="success" style="display: flex; flex-direction: column; align-items: center; gap: 1rem">
				<div style="font-size: 3rem; color: var(--success)">✅</div>
				<p style="color: var(--fg-2)">Your email has been verified successfully!</p>
				<router-link to="/" class="btn btn-primary">Go to Dashboard</router-link>
			</div>
			<div v-else style="display: flex; flex-direction: column; align-items: center; gap: 1rem">
				<div style="font-size: 3rem; color: var(--error)">❌</div>
				<p style="color: var(--fg-2)">{{ error || "Verification failed." }}</p>
				<router-link to="/settings" class="btn btn-secondary">Back to Settings</router-link>
			</div>
		</div>
	</div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useAuthStore } from "../stores/auth";
import { useRoute } from "vue-router";

const auth = useAuthStore();
const route = useRoute();
const loading = ref(true);
const success = ref(false);
const error = ref(null);

onMounted(async () => {
	const token = route.query.token;
	if (!token) {
		error.value = "Missing verification token.";
		loading.value = false;
		return;
	}

	try {
		await auth.verifyEmail(token);
		success.value = true;
	} catch (err) {
		error.value = err.response?.data || "Verification failed.";
	} finally {
		loading.value = false;
	}
});
</script>
