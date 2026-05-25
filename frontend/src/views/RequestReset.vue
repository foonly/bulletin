<template>
	<div class="min-h-screen flex items-center justify-center bg-gray-950">
		<div class="bg-gray-800 p-8 rounded-lg shadow-xl w-96">
			<h1 class="text-2xl font-bold mb-6 text-center">Reset Password</h1>
			<form v-if="!submitted" @submit.prevent="handleSubmit">
				<div class="mb-6">
					<label class="block text-sm font-medium mb-1">Email Address</label>
					<input
						v-model="email"
						type="email"
						class="w-full p-2 rounded bg-gray-700 border border-gray-600 focus:outline-none focus:border-purple-500"
						required
					/>
				</div>
				<button
					type="submit"
					class="w-full bg-purple-600 hover:bg-purple-700 p-2 rounded font-bold transition"
				>
					Send Reset Link
				</button>
			</form>
			<div v-else class="text-center space-y-4">
				<p class="text-gray-300">
					If an account with that email exists, we've sent a password reset
					link.
				</p>
				<p class="text-sm text-gray-500">
					Be sure to check your spam folder if you don't see it within a few
					minutes.
				</p>
			</div>
			<p class="mt-6 text-center text-sm">
				<router-link to="/login" class="text-purple-400 hover:underline"
					>Back to Login</router-link
				>
			</p>
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
