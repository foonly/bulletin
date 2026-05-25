<template>
	<div class="min-h-screen flex items-center justify-center bg-gray-950">
		<div class="bg-gray-800 p-8 rounded-lg shadow-xl w-96 text-center">
			<h1 class="text-2xl font-bold mb-6">Email Verification</h1>
			<div v-if="loading" class="space-y-4">
				<p class="text-gray-300">Verifying your email address...</p>
			</div>
			<div v-else-if="success" class="space-y-4">
				<div class="text-6xl text-green-500 mb-4">✅</div>
				<p class="text-gray-300">Your email has been verified successfully!</p>
				<router-link
					to="/"
					class="inline-block bg-purple-600 hover:bg-purple-700 px-6 py-2 rounded font-bold transition"
				>
					Go to Dashboard
				</router-link>
			</div>
			<div v-else class="space-y-4">
				<div class="text-6xl text-red-500 mb-4">❌</div>
				<p class="text-gray-300">{{ error || "Verification failed." }}</p>
				<router-link
					to="/settings"
					class="inline-block bg-gray-700 hover:bg-gray-600 px-6 py-2 rounded font-bold transition"
				>
					Back to Settings
				</router-link>
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
