<template>
	<div class="settings-page">
		<h1 style="margin-bottom: 1.5rem">User Settings</h1>

		<div class="section-card settings-section">
			<h3 class="section-card__title">Account</h3>
			<form @submit.prevent="handleUpdate">
				<div class="field">
					<label for="set-username">Username</label>
					<input id="set-username" v-model="form.username" type="text" />
				</div>

				<div class="field">
					<div class="label-meta">
						<label for="set-email">Email Address</label>
						<span
							v-if="auth.user?.email"
							:class="auth.user.is_email_verified ? 'badge badge-success' : 'badge badge-warning'"
						>{{ auth.user.is_email_verified ? "Verified" : "Unverified" }}</span>
					</div>
					<input id="set-email" v-model="form.email" type="email" />
					<button
						v-if="auth.user?.email && !auth.user.is_email_verified"
						type="button"
						class="btn btn-ghost btn-sm"
						style="align-self: flex-start; color: var(--accent)"
						@click="handleRequestVerification"
					>Resend Verification Email</button>
				</div>

				<hr style="margin: 1.5rem 0" />

				<h3 style="font-size: var(--text-sm); font-weight: 700; color: var(--fg-2); text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 1rem">
					Change Password
				</h3>

				<div class="field">
					<label for="set-old-pw">Current Password</label>
					<input
						id="set-old-pw"
						v-model="form.oldPassword"
						type="password"
						placeholder="Required to change password"
					/>
				</div>

				<div class="field-row">
					<div class="field">
						<label for="set-new-pw">New Password</label>
						<input id="set-new-pw" v-model="form.password" type="password" />
					</div>
					<div class="field">
						<label for="set-confirm-pw">Confirm Password</label>
						<input id="set-confirm-pw" v-model="form.confirmPassword" type="password" />
					</div>
				</div>

				<div class="form-actions form-actions--between" style="margin-top: 2rem">
					<button type="submit" class="btn btn-primary">Save Changes</button>
					<router-link to="/" style="font-size: var(--text-sm); color: var(--fg-2)">
						Back to Dashboard
					</router-link>
				</div>
			</form>
		</div>

		<!-- Notifications -->
		<div class="section-card settings-section">
			<div style="display: flex; align-items: center; justify-content: space-between; padding-bottom: 0.75rem; border-bottom: 1px solid var(--border); margin-bottom: 1rem">
				<h3>Browser Notifications</h3>
				<span :class="auth.notificationsEnabled ? 'badge badge-success' : 'badge badge-neutral'">
					{{ auth.notificationsEnabled ? "Enabled" : "Disabled" }}
				</span>
			</div>
			<div class="setting-panel">
				<p>
					Receive desktop notifications for new chat messages and forum activity
					when the tab is in the background.
				</p>
				<button
					class="btn btn-full"
					:class="auth.notificationsEnabled ? 'btn-secondary' : 'btn-primary'"
					@click="auth.setNotificationsEnabled(!auth.notificationsEnabled)"
				>
					{{ auth.notificationsEnabled ? "Disable Notifications" : "Enable Notifications" }}
				</button>
			</div>
		</div>

		<!-- MFA -->
		<div class="section-card settings-section">
			<div style="display: flex; align-items: center; justify-content: space-between; padding-bottom: 0.75rem; border-bottom: 1px solid var(--border); margin-bottom: 1rem">
				<h3>Multi-Factor Authentication</h3>
				<span :class="auth.user?.totp_enabled ? 'badge badge-success' : 'badge badge-neutral'">
					{{ auth.user?.totp_enabled ? "Enabled" : "Disabled" }}
				</span>
			</div>

			<div v-if="!auth.user?.totp_enabled && !mfaSetup" class="setting-panel">
				<p>
					Add an extra layer of security to your account by requiring a code
					from your authenticator app when logging in.
				</p>
				<button class="btn btn-secondary btn-full" @click="startMfaSetup">
					Enable MFA
				</button>
			</div>

			<!-- MFA setup flow -->
			<div v-if="mfaSetup" class="mfa-setup">
				<p style="font-weight: 700">1. Scan this QR Code</p>
				<div class="mfa-qr-wrapper">
					<qrcode-vue :value="mfaSetup.url" :size="160" level="H" />
				</div>
				<p class="mfa-secret">Secret: {{ mfaSetup.secret }}</p>
				<p style="font-weight: 700; margin-top: 0.5rem">2. Enter the code from your app</p>
				<input
					v-model="mfaCode"
					placeholder="000000"
					maxlength="6"
					class="input-code"
					style="width: 100%"
				/>
				<div style="display: flex; gap: 0.75rem; width: 100%">
					<button class="btn btn-secondary" style="flex: 1" @click="mfaSetup = null">Cancel</button>
					<button
						class="btn btn-primary"
						style="flex: 1"
						:disabled="mfaCode.length !== 6"
						@click="verifyAndEnableMfa"
					>Verify & Enable</button>
				</div>
			</div>

			<!-- Disable MFA -->
			<div v-if="auth.user?.totp_enabled" class="setting-panel">
				<div v-if="!showMfaDisable">
					<button class="btn btn-danger btn-full" @click="showMfaDisable = true">
						Disable MFA
					</button>
				</div>
				<div v-else style="display: flex; flex-direction: column; gap: 0.75rem">
					<p style="font-weight: 700; color: var(--danger)">Confirm MFA Deactivation</p>
					<input
						v-model="disablePassword"
						type="password"
						placeholder="Enter password to confirm"
					/>
					<div style="display: flex; gap: 0.75rem">
						<button class="btn btn-secondary" style="flex: 1" @click="showMfaDisable = false">Cancel</button>
						<button class="btn btn-danger" style="flex: 1" @click="handleDisableMfa">Confirm Disable</button>
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

const mfaSetup = ref(null);
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
