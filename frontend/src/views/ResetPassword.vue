<template>
	<div class="min-h-screen flex items-center justify-center bg-gray-950">
		<div class="bg-gray-800 p-8 rounded-lg shadow-xl w-96">
			<h1 class="text-2xl font-bold mb-6 text-center">New Password</h1>
			<form v-if="!submitted" @submit.prevent="handleSubmit">
				<div class="mb-6">
					<label class="block text-sm font-medium mb-1">New Password</label>
					<input
						v-model="password"
						type="password"
						class="w-full p-2 rounded bg-gray-700 border border-gray-600 focus:outline-none focus:border-purple-500"
						required
						minlength="8"
					/>
				</div>
				<button
					type="submit"
					class="w-full bg-purple-600 hover:bg-purple-700 p-2 rounded font-bold transition"
				>
					Update Password
				</button>
			</form>
			<div v-else class="text-center space-y-4">
				<p class="text-gray-300">Your password has been updated successfully.</p>
				<router-link
					to="/login"
					class="inline-block bg-purple-600 hover:bg-purple-700 px-6 py-2 rounded font-bold transition"
				>
					Go to Login
				</router-link>
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
