<template>
	<div class="min-h-screen flex items-center justify-center bg-gray-950">
		<div class="bg-gray-800 p-8 rounded-lg shadow-xl w-96">
			<h1 class="text-2xl font-bold mb-6 text-center">Join {{ siteName }}</h1>
			<form @submit.prevent="handleRegister">
				<div class="mb-4">
					<label class="block text-sm font-medium mb-1">Invite Code</label>
					<input
						v-model="inviteCode"
						type="text"
						class="w-full p-2 rounded bg-gray-700 border border-gray-600 focus:outline-none focus:border-purple-500"
						required
					/>
				</div>
				<div class="mb-4">
					<label class="block text-sm font-medium mb-1">Username</label>
					<input
						v-model="username"
						type="text"
						class="w-full p-2 rounded bg-gray-700 border border-gray-600 focus:outline-none focus:border-purple-500"
						required
					/>
				</div>
				<div class="mb-4">
					<label class="block text-sm font-medium mb-1">Email Address</label>
					<input
						v-model="email"
						type="email"
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
					class="w-full bg-green-600 hover:bg-green-700 p-2 rounded font-bold transition"
				>
					Register
				</button>
			</form>
			<p class="mt-4 text-center text-sm">
				Already have an account?
				<router-link to="/login" class="text-purple-400 hover:underline"
					>Login</router-link
				>
			</p>
		</div>
	</div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useAuthStore } from "../stores/auth";
import { useToastStore } from "../stores/toast";
import { useRouter, useRoute } from "vue-router";

const auth = useAuthStore();
const toast = useToastStore();
const router = useRouter();
const route = useRoute();
const siteName = import.meta.env.VITE_SITE_NAME || "Bulletin";
const inviteCode = ref("");
const username = ref("");
const email = ref("");
const password = ref("");

onMounted(() => {
	if (route.query.code) {
		inviteCode.value = route.query.code;
	}
});

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
