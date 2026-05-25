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
			<div class="mb-4">
				<div class="flex items-center justify-between mb-1">
					<label class="block text-sm font-medium">Email Address</label>
					<span
						v-if="auth.user?.email"
						:class="[
							'text-[10px] uppercase font-bold px-1.5 py-0.5 rounded',
							auth.user.is_email_verified
								? 'bg-green-900/30 text-green-400 border border-green-800/50'
								: 'bg-yellow-900/30 text-yellow-400 border border-yellow-800/50',
						]"
					>
						{{ auth.user.is_email_verified ? "Verified" : "Unverified" }}
					</span>
				</div>
				<div class="space-y-2">
					<input
						v-model="form.email"
						type="email"
						class="w-full p-2 rounded bg-gray-700 border border-gray-600 focus:outline-none focus:border-purple-500"
					/>
					<button
						v-if="auth.user?.email && !auth.user.is_email_verified"
						type="button"
						@click="handleRequestVerification"
						class="text-xs text-purple-400 hover:text-purple-300 font-bold"
					>
						Resend Verification Email
					</button>
				</div>
			</div>

			<hr class="my-6 border-gray-700" />

			<div class="space-y-4">
				<h3 class="text-sm font-bold text-gray-400 uppercase tracking-wider">
					Change Password
				</h3>
				<div class="mb-4">
					<label class="block text-sm font-medium mb-1">Current Password</label>
					<input
						v-model="form.oldPassword"
						type="password"
						placeholder="Required to change password"
						class="w-full p-2 rounded bg-gray-700 border border-gray-600 focus:outline-none focus:border-purple-500"
					/>
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium mb-1">New Password</label>
						<input
							v-model="form.password"
							type="password"
							class="w-full p-2 rounded bg-gray-700 border border-gray-600 focus:outline-none focus:border-purple-500"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium mb-1"
							>Confirm Password</label
						>
						<input
							v-model="form.confirmPassword"
							type="password"
							class="w-full p-2 rounded bg-gray-700 border border-gray-600 focus:outline-none focus:border-purple-500"
						/>
					</div>
				</div>
			</div>

			<div class="flex items-center justify-between mt-8">
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

		<hr class="my-8 border-gray-700" />

		<!-- Notifications Section -->
		<div class="space-y-4">
			<div class="flex items-center justify-between">
				<h3 class="text-sm font-bold text-gray-400 uppercase tracking-wider">
					Browser Notifications
				</h3>
				<span
					:class="[
						'text-[10px] uppercase font-bold px-1.5 py-0.5 rounded',
						auth.notificationsEnabled
							? 'bg-green-900/30 text-green-400 border border-green-800/50'
							: 'bg-gray-700 text-gray-400',
					]"
				>
					{{ auth.notificationsEnabled ? "Enabled" : "Disabled" }}
				</span>
			</div>
			<div class="bg-gray-900/50 p-4 rounded-lg border border-gray-700">
				<p class="text-sm text-gray-400 mb-4">
					Receive desktop notifications for new chat messages and forum activity
					when the tab is in the background.
				</p>
				<button
					@click="auth.setNotificationsEnabled(!auth.notificationsEnabled)"
					:class="[
						'w-full py-2 rounded font-bold transition text-sm border',
						auth.notificationsEnabled
							? 'bg-gray-800 border-gray-700 text-gray-300 hover:bg-gray-700'
							: 'bg-purple-600 border-purple-500 text-white hover:bg-purple-700',
					]"
				>
					{{
						auth.notificationsEnabled
							? "Disable Notifications"
							: "Enable Notifications"
					}}
				</button>
			</div>
		</div>

		<hr class="my-8 border-gray-700" />

		<!-- MFA Section -->
		<div class="space-y-4">
			<div class="flex items-center justify-between">
				<h3 class="text-sm font-bold text-gray-400 uppercase tracking-wider">
					Multi-Factor Authentication
				</h3>
				<span
					:class="[
						'text-[10px] uppercase font-bold px-1.5 py-0.5 rounded',
						auth.user?.totp_enabled
							? 'bg-green-900/30 text-green-400 border border-green-800/50'
							: 'bg-gray-700 text-gray-400',
					]"
				>
					{{ auth.user?.totp_enabled ? "Enabled" : "Disabled" }}
				</span>
			</div>

			<div
				v-if="!auth.user?.totp_enabled && !mfaSetup"
				class="bg-gray-900/50 p-4 rounded-lg border border-gray-700"
			>
				<p class="text-sm text-gray-400 mb-4">
					Add an extra layer of security to your account by requiring a code
					from your authenticator app when logging in.
				</p>
				<button
					@click="startMfaSetup"
					class="w-full bg-gray-700 hover:bg-gray-600 text-white py-2 rounded font-bold transition text-sm"
				>
					Enable MFA
				</button>
			</div>

			<!-- MFA Setup Flow -->
			<div
				v-if="mfaSetup"
				class="bg-gray-900 p-6 rounded-lg border border-purple-500/30 space-y-6"
			>
				<div class="text-center space-y-4">
					<p class="text-sm font-bold">1. Scan this QR Code</p>
					<div class="inline-block bg-white p-2 rounded">
						<qrcode-vue :value="mfaSetup.url" :size="160" level="H" />
					</div>
					<p class="text-xs text-gray-500 font-mono break-all">
						Secret: {{ mfaSetup.secret }}
					</p>
				</div>

				<div class="space-y-4">
					<p class="text-sm font-bold text-center">
						2. Enter the code from your app
					</p>
					<input
						v-model="mfaCode"
						placeholder="000000"
						maxlength="6"
						class="w-full bg-gray-800 border border-gray-700 p-3 rounded text-center text-xl tracking-[0.5em] focus:outline-none focus:border-purple-500"
					/>
					<div class="flex space-x-3">
						<button
							@click="mfaSetup = null"
							class="flex-1 px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded font-bold transition text-sm"
						>
							Cancel
						</button>
						<button
							@click="verifyAndEnableMfa"
							:disabled="mfaCode.length !== 6"
							class="flex-1 px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded font-bold transition text-sm disabled:opacity-50"
						>
							Verify & Enable
						</button>
					</div>
				</div>
			</div>

			<!-- Disable MFA -->
			<div
				v-if="auth.user?.totp_enabled"
				class="bg-gray-900/50 p-4 rounded-lg border border-gray-700"
			>
				<div v-if="!showMfaDisable">
					<button
						@click="showMfaDisable = true"
						class="w-full bg-red-900/20 text-red-400 hover:bg-red-900/30 border border-red-900/50 py-2 rounded font-bold transition text-sm"
					>
						Disable MFA
					</button>
				</div>
				<div v-else class="space-y-4">
					<p class="text-sm font-bold text-red-400">Confirm MFA Deactivation</p>
					<input
						v-model="disablePassword"
						type="password"
						placeholder="Enter password to confirm"
						class="w-full bg-gray-800 border border-gray-700 p-2 rounded focus:outline-none focus:border-red-500"
					/>
					<div class="flex space-x-3">
						<button
							@click="showMfaDisable = false"
							class="flex-1 px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded font-bold transition text-sm"
						>
							Cancel
						</button>
						<button
							@click="handleDisableMfa"
							class="flex-1 px-4 py-2 bg-red-600 hover:bg-red-700 rounded font-bold transition text-sm"
						>
							Confirm Disable
						</button>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useAuthStore } from "../stores/auth";
import { useToastStore } from "../stores/toast";
import { useRouter } from "vue-router";
import QrcodeVue from "qrcode.vue";

const auth = useAuthStore();
const toast = useToastStore();
const router = useRouter();
const form = ref({
	username: "",
	email: "",
	oldPassword: "",
	password: "",
	confirmPassword: "",
});

const mfaSetup = ref(null); // { secret: string, url: string }
const mfaCode = ref("");
const showMfaDisable = ref(false);
const disablePassword = ref("");

onMounted(() => {
	if (auth.user) {
		form.value.username = auth.user.username;
		form.value.email = auth.user.email || "";
	}
});

const handleRequestVerification = async () => {
	try {
		await auth.requestVerification();
		toast.success("Verification email sent!");
	} catch (err) {
		toast.error("Failed to send verification email");
	}
};

const handleUpdate = async () => {
	if (form.value.password || form.value.confirmPassword) {
		if (form.value.password !== form.value.confirmPassword) {
			toast.error("New passwords do not match");
			return;
		}
		if (!form.value.oldPassword) {
			toast.error("Current password is required to change password");
			return;
		}
	}

	try {
		await auth.updateMe({
			username: form.value.username,
			email: form.value.email,
			old_password: form.value.oldPassword,
			password: form.value.password,
		});
		toast.success("Settings updated successfully!");
		// Clear password fields
		form.value.oldPassword = "";
		form.value.password = "";
		form.value.confirmPassword = "";
	} catch (err) {
		toast.error(err.response?.data || "Failed to update settings");
	}
};

const startMfaSetup = async () => {
	try {
		mfaSetup.value = await auth.setupTOTP();
	} catch (err) {
		toast.error("Failed to start MFA setup");
	}
};

const verifyAndEnableMfa = async () => {
	try {
		await auth.enableTOTP(mfaCode.value);
		toast.success("MFA enabled successfully!");
		mfaSetup.value = null;
		mfaCode.value = "";
	} catch (err) {
		toast.error(err.response?.data || "Invalid MFA code");
	}
};

const handleDisableMfa = async () => {
	try {
		await auth.disableTOTP(disablePassword.value);
		toast.success("MFA disabled");
		showMfaDisable.value = false;
		disablePassword.value = "";
	} catch (err) {
		toast.error(err.response?.data || "Failed to disable MFA");
	}
};
</script>
