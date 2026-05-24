<template>
	<div class="min-h-screen flex items-center justify-center bg-gray-950">
		<div class="bg-gray-800 p-8 rounded-lg shadow-xl w-96">
			<h1 class="text-2xl font-bold mb-6 text-center">Bulletin Login</h1>
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
					class="w-full bg-purple-600 hover:bg-purple-700 p-2 rounded font-bold transition"
				>
					Login
				</button>
			</form>
			<p class="mt-4 text-center text-sm">
				Don't have an account?
				<router-link to="/register" class="text-purple-400 hover:underline"
					>Register with invite</router-link
				>
			</p>
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
const username = ref("");
const password = ref("");

const handleLogin = async () => {
	try {
		await auth.login(username.value, password.value);
		toast.success("Welcome back!");
		router.push("/");
	} catch (err) {
		toast.error(err.response?.data || "Login failed");
	}
};
</script>
