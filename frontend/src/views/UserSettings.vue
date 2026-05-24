<template>
	<div class="max-w-md mx-auto mt-10 p-6 bg-gray-800 rounded-lg shadow-xl">
		<h1 class="text-2xl font-bold mb-6">User Settings</h1>
		<form @submit.prevent="handleUpdate">
			<div class="mb-4">
				<label class="block text-sm font-medium mb-1">Username</label>
				<input
					v-model="form.username"
					type="text"
					class="w-full p-2 rounded bg-gray-700 border border-gray-600 focus:outline-none focus:border-purple-500"
				/>
			</div>
			<div class="mb-6">
				<label class="block text-sm font-medium mb-1"
					>New Password (leave blank to keep current)</label
				>
				<input
					v-model="form.password"
					type="password"
					class="w-full p-2 rounded bg-gray-700 border border-gray-600 focus:outline-none focus:border-purple-500"
				/>
			</div>
			<div class="flex items-center justify-between">
				<button
					type="submit"
					class="bg-purple-600 hover:bg-purple-700 px-4 py-2 rounded font-bold transition"
				>
					Save Changes
				</button>
				<router-link to="/" class="text-sm text-gray-400 hover:underline"
					>Back to Dashboard</router-link
				>
			</div>
		</form>
	</div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useAuthStore } from "../stores/auth";
import { useToastStore } from "../stores/toast";
import { useRouter } from "vue-router";

const auth = useAuthStore();
const toast = useToastStore();
const router = useRouter();
const form = ref({
	username: "",
	password: "",
});

onMounted(() => {
	if (auth.user) {
		form.value.username = auth.user.username;
	}
});

const handleUpdate = async () => {
	try {
		await auth.updateMe(form.value);
		toast.success("Settings updated successfully!");
	} catch (err) {
		toast.error(err.response?.data || "Failed to update settings");
	}
};
</script>
