<script setup>
import { ref, onMounted, watch, computed } from "vue";
import { useCircleStore } from "../../stores/circles";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";

const props = defineProps(["id", "members"]);
const circleStore = useCircleStore();
const auth = useAuthStore();
const toast = useToastStore();

const isAdmin = computed(() => {
	return circleStore.activeCircle?.role === "admin";
});

const settings = ref({
	name: "",
	description: "",
	allow_freeform_tags: true,
	invite_min_role: "standard",
	chat_retention_days: 14,
	chat_retention_count: 50,
});

const inviteForm = ref({
	code: "",
	role_to_grant: "standard",
	max_uses: null,
	expires_in_hrs: null,
});

const adminNewTagName = ref("");

const syncSettings = () => {
	if (circleStore.activeCircle) {
		settings.value = { ...circleStore.activeCircle };
	}
};

onMounted(syncSettings);
watch(() => circleStore.activeCircle, syncSettings);

const saveSettings = async () => {
	try {
		await circleStore.updateCircle(props.id, settings.value);
		toast.success("Circle settings updated!");
	} catch (err) {
		toast.error("Failed to update settings");
	}
};

const generateInvite = async () => {
	try {
		await circleStore.createInvite(props.id, {
			role_to_grant: inviteForm.value.role_to_grant,
			max_uses: inviteForm.value.max_uses
				? parseInt(inviteForm.value.max_uses)
				: null,
			expires_in_hrs: inviteForm.value.expires_in_hrs
				? parseInt(inviteForm.value.expires_in_hrs)
				: null,
		});
		toast.success("Invite link generated!");
		await circleStore.fetchInvites(props.id);
	} catch (err) {
		toast.error("Failed to generate invite");
	}
};

const revokeInvite = async (inviteId) => {
	try {
		await circleStore.deleteInvite(props.id, inviteId);
		toast.success("Invite revoked");
	} catch (err) {
		toast.error("Failed to revoke invite");
	}
};

const adminAddTag = async () => {
	const tag = adminNewTagName.value.trim().toLowerCase();
	if (!tag) return;

	try {
		await circleStore.createTag(props.id, tag);
		toast.success(`Tag #${tag} created!`);
		adminNewTagName.value = "";
	} catch (err) {
		toast.error("Failed to create tag");
	}
};

const togglePin = async (tagId, isPinned) => {
	try {
		await circleStore.pinTag(props.id, tagId, isPinned);
	} catch (err) {
		toast.error("Failed to update tag pin");
	}
};

const updateMemberRole = async (userId, role) => {
	try {
		await circleStore.updateMember(props.id, userId, role);
		toast.success("Member role updated");
	} catch (err) {
		toast.error("Failed to update member role");
	}
};

const kickMember = async (userId) => {
	if (!confirm("Are you sure you want to remove this member?")) return;
	try {
		await circleStore.deleteMember(props.id, userId);
		toast.success("Member removed");
		// We'd need to refresh members list here, but it's in parent.
		// Maybe parent should provide refresh method or use window events.
		window.dispatchEvent(new CustomEvent("refresh-members"));
	} catch (err) {
		toast.error(err.response?.data || "Failed to kick member");
	}
};

const formatDate = (dateStr) => {
	return new Date(dateStr).toLocaleString();
};
</script>

<template>
	<div class="space-y-8 w-full max-w-2xl mx-auto">
		<!-- Circle Config -->
		<section class="bg-gray-800 p-6 rounded-lg space-y-4">
			<h3 class="text-xl font-bold border-b border-gray-700 pb-2">
				Circle Settings
			</h3>
			<div class="grid grid-cols-2 gap-4">
				<div class="flex flex-col">
					<label class="text-xs text-gray-400 mb-1">Name</label>
					<input
						v-model="settings.name"
						class="bg-gray-700 p-2 rounded focus:outline-none focus:ring-1 focus:ring-purple-500"
					/>
				</div>
				<div class="flex flex-col">
					<label class="text-xs text-gray-400 mb-1">Invite Min Role</label>
					<select
						v-model="settings.invite_min_role"
						class="bg-gray-700 p-2 rounded focus:outline-none"
					>
						<option value="guest">Guest</option>
						<option value="standard">Standard</option>
						<option value="mod">Moderator</option>
						<option value="admin">Admin</option>
					</select>
				</div>
				<div class="flex flex-col col-span-2">
					<label class="text-xs text-gray-400 mb-1">Description</label>
					<textarea
						v-model="settings.description"
						class="bg-gray-700 p-2 rounded focus:outline-none h-20"
					></textarea>
				</div>
				<div class="flex items-center space-x-2">
					<input
						type="checkbox"
						v-model="settings.allow_freeform_tags"
						id="freeform"
					/>
					<label for="freeform" class="text-sm"
						>Allow users to create tags</label
					>
				</div>
			</div>
			<div class="flex justify-end">
				<button
					@click="saveSettings"
					class="bg-purple-600 hover:bg-purple-700 px-6 py-2 rounded font-bold transition"
				>
					Save Changes
				</button>
			</div>
		</section>

		<!-- Invites -->
		<section class="bg-gray-800 p-6 rounded-lg space-y-4">
			<h3 class="text-xl font-bold border-b border-gray-700 pb-2">Invites</h3>
			<div class="grid grid-cols-2 gap-4 text-sm">
				<div class="flex flex-col">
					<label class="text-xs text-gray-400 mb-1">Role to Grant</label>
					<select
						v-model="inviteForm.role_to_grant"
						class="bg-gray-700 p-2 rounded focus:outline-none"
					>
						<option value="guest">Guest</option>
						<option value="standard">Standard</option>
						<option value="mod">Moderator</option>
						<option value="admin">Admin</option>
					</select>
				</div>
				<div class="flex flex-col">
					<label class="text-xs text-gray-400 mb-1">Max Uses (optional)</label>
					<input
						v-model="inviteForm.max_uses"
						type="number"
						class="bg-gray-700 p-2 rounded focus:outline-none"
					/>
				</div>
				<div class="flex flex-col">
					<label class="text-xs text-gray-400 mb-1"
						>Expires in Hours (optional)</label
					>
					<input
						v-model="inviteForm.expires_in_hrs"
						type="number"
						class="bg-gray-700 p-2 rounded focus:outline-none"
					/>
				</div>
			</div>
			<div class="flex justify-end">
				<button
					@click="generateInvite"
					class="bg-purple-600 hover:bg-purple-700 px-6 py-2 rounded font-bold transition"
				>
					Generate Invite
				</button>
			</div>

			<div v-if="circleStore.invites.length > 0" class="mt-6 space-y-3">
				<h4 class="text-sm font-bold text-gray-400">Active Invites</h4>
				<div
					v-for="invite in circleStore.invites"
					:key="invite.id"
					class="flex items-center justify-between bg-gray-900 p-3 rounded border border-gray-700"
				>
					<div class="space-y-1">
						<div class="flex items-center space-x-2">
							<code
								class="bg-purple-900/30 text-purple-400 px-2 py-0.5 rounded font-mono font-bold"
								>{{ invite.code }}</code
							>
							<span class="text-xs text-gray-500"
								>grants {{ invite.role_to_grant }} • Issued by
								{{ invite.created_by }}</span
							>
						</div>
						<div class="text-[10px] text-gray-500">
							Uses: {{ invite.used_count }} / {{ invite.max_uses || "∞" }} •
							Expires:
							{{ invite.expires_at ? formatDate(invite.expires_at) : "Never" }}
						</div>
					</div>
					<button
						@click="revokeInvite(invite.id)"
						class="text-xs text-red-400 hover:underline"
					>
						Revoke
					</button>
				</div>
			</div>
		</section>

		<!-- Manage Tags -->
		<section class="bg-gray-800 p-6 rounded-lg space-y-4">
			<h3 class="text-xl font-bold border-b border-gray-700 pb-2">
				Manage Tags
			</h3>
			<div class="flex items-center space-x-2">
				<input
					v-model="adminNewTagName"
					placeholder="New tag name"
					class="flex-1 bg-gray-700 p-2 rounded focus:outline-none"
				/>
				<button
					@click="adminAddTag"
					class="bg-purple-600 hover:bg-purple-700 px-4 py-2 rounded font-bold transition"
				>
					Add Tag
				</button>
			</div>
			<div class="flex flex-wrap gap-2 pt-2">
				<div
					v-for="tag in circleStore.tags"
					:key="tag.id"
					class="bg-gray-900 border border-gray-700 px-3 py-1.5 rounded-lg flex items-center"
				>
					<span :class="tag.is_pinned ? 'text-purple-400' : ''"
						>#{{ tag.name }}</span
					>
					<button
						@click="togglePin(tag.id, !tag.is_pinned)"
						class="ml-2 text-xs text-gray-500 hover:text-purple-400"
					>
						{{ tag.is_pinned ? "Unpin" : "Pin" }}
					</button>
				</div>
			</div>
		</section>

		<section
			v-if="isAdmin"
			class="bg-gray-800 p-6 rounded-lg space-y-4 border border-red-900/20"
		>
			<h3
				class="text-xl font-bold border-b border-red-900/50 pb-2 text-red-400"
			>
				Danger Zone
			</h3>
			<div class="space-y-2">
				<div
					v-for="m in members"
					:key="m.id"
					class="flex items-center justify-between bg-gray-900 p-3 rounded border border-gray-700"
				>
					<div>
						<span class="font-bold">{{ m.username }}</span>
						<span class="text-xs text-gray-400 ml-2"
							>Invited by {{ m.invited_by }}</span
						>
					</div>
					<div class="flex items-center space-x-2">
						<select
							v-model="m.role"
							@change="updateMemberRole(m.id, m.role)"
							class="bg-gray-700 text-xs p-1 rounded focus:outline-none"
						>
							<option value="guest">Guest</option>
							<option value="standard">Standard</option>
							<option value="mod">Mod</option>
							<option value="admin">Admin</option>
						</select>
						<button
							v-if="m.id !== auth.user?.id"
							@click="kickMember(m.id)"
							class="bg-red-900/20 text-red-400 px-2 py-1 rounded text-xs hover:bg-red-900/40 transition"
						>
							Kick
						</button>
					</div>
				</div>
			</div>
		</section>
	</div>
</template>
