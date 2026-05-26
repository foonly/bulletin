<template>
	<div class="min-h-screen flex items-center justify-center bg-gray-950 p-4">
		<div
			class="bg-gray-800 p-8 rounded-xl shadow-2xl w-full max-w-md border border-gray-700"
		>
			<div
				v-if="loading"
				class="flex flex-col items-center justify-center py-12 space-y-4"
			>
				<div
					class="w-12 h-12 border-4 border-purple-500 border-t-transparent rounded-full animate-spin"
				></div>
				<p class="text-gray-400">Verifying invite...</p>
			</div>

			<div v-else-if="error" class="text-center py-8">
				<div class="text-5xl mb-4">❌</div>
				<h1 class="text-2xl font-bold mb-2">Invalid Invite</h1>
				<p class="text-gray-400 mb-6">{{ error }}</p>
				<router-link
					to="/"
					class="inline-block bg-gray-700 hover:bg-gray-600 px-6 py-2 rounded-lg font-bold transition"
				>
					Go Home
				</router-link>
			</div>

			<div v-else class="text-center">
				<div
					class="w-20 h-20 bg-purple-600 rounded-2xl flex items-center justify-center font-bold text-4xl shadow-lg mx-auto mb-6"
				>
					{{
						inviteInfo.circle_name
							? inviteInfo.circle_name[0].toUpperCase()
							: "?"
					}}
				</div>
				<h1 class="text-2xl font-bold mb-2">You've been invited!</h1>
				<p class="text-gray-400 mb-8">
					Join
					<span class="text-white font-bold">{{ inviteInfo.circle_name }}</span>
					on {{ siteName }}.
				</p>

				<div class="space-y-3">
					<button
						@click="handleJoin"
						class="w-full bg-purple-600 hover:bg-purple-700 p-3 rounded-lg font-bold transition flex items-center justify-center space-x-2"
					>
						<span>Join Circle</span>
					</button>
					<p class="text-sm text-gray-500">
						{{
							auth.user
								? `Logged in as ${auth.user.username}`
								: "You will need to login or register first."
						}}
					</p>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";
import { useCircleStore } from "../stores/circles";
import { useToastStore } from "../stores/toast";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();
const circleStore = useCircleStore();
const toast = useToastStore();

const siteName = import.meta.env.VITE_SITE_NAME || "Bulletin";
const code = route.params.code;
const loading = ref(true);
const error = ref(null);
const inviteInfo = ref(null);

onMounted(async () => {
	// Ensure we know the user's auth status
	if (!auth.user) {
		try {
			await auth.fetchMe();
		} catch (e) {
			// Not logged in, that's fine
		}
	}

	try {
		const info = await circleStore.getInviteInfo(code);
		if (!info.valid) {
			error.value = "This invite link has expired or reached its maximum uses.";
		} else {
			inviteInfo.value = info;

			// If the user is already logged in, try to join automatically
			if (auth.user) {
				try {
					const res = await circleStore.joinCircle(code);
					toast.success(`Welcome to ${inviteInfo.value.circle_name}!`);
					router.push(`/circle/${res.id}`);
				} catch (err) {
					// If they are already a member, just redirect to the circle
					if (err.response?.status === 409) {
						// We need the circle ID to redirect.
						// The joinCircle endpoint returns it on 200,
						// but on 409 we might need to find it in the user's circles.
						await circleStore.fetchCircles();
						const circle = circleStore.circles.find(
							(c) => c.name === info.circle_name,
						);
						if (circle) {
							router.push(`/circle/${circle.id}`);
						} else {
							router.push("/");
						}
					}
					// For other errors, we stay on the invite page so they can see the "Join" button or error
				}
			}
		}
	} catch (err) {
		error.value =
			"We couldn't find that invite. It might be invalid or deleted.";
	} finally {
		loading.value = false;
	}
});

const handleJoin = async () => {
	if (!auth.user) {
		// If not logged in, we check if they want to register or login.
		// We'll redirect to register by default as it's an invite,
		// but we'll pass the invite code.
		router.push({
			path: "/register",
			query: { code: code },
		});
		return;
	}

	try {
		const res = await circleStore.joinCircle(code);
		toast.success(`Welcome to ${inviteInfo.value.circle_name}!`);
		router.push(`/circle/${res.id}`);
	} catch (err) {
		if (err.response?.status === 409) {
			toast.info("You are already a member of this circle.");
			// Find the circle ID if possible or just go home
			router.push("/");
		} else {
			toast.error(err.response?.data || "Failed to join circle");
		}
	}
};
</script>
