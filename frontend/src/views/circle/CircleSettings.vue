<script setup>
import { ref, onMounted, watch, computed } from "vue";
import { useCircleStore } from "../../stores/circles";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";
import InviteModal from "../../components/InviteModal.vue";

const props = defineProps(["id", "members"]);
const circleStore = useCircleStore();
const auth = useAuthStore();
const toast = useToastStore();

const showInviteModal = ref(false);

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

const adminNewTagName = ref("");

const editingTag = ref(null);
const editTagName = ref("");

const mergingTag = ref(null);
const mergeTargetTagId = ref("");

const deletingTag = ref(null);

const startEditTag = (tag) => {
	editingTag.value = tag;
	editTagName.value = tag.name;
};

const cancelEditTag = () => {
	editingTag.value = null;
	editTagName.value = "";
};

const handleUpdateTag = async () => {
	const name = editTagName.value.trim().toLowerCase();
	if (!name || name === editingTag.value.name) {
		cancelEditTag();
		return;
	}

	try {
		await circleStore.updateTag(props.id, editingTag.value.id, name);
		toast.success(`Tag renamed to #${name}`);
		editingTag.value = null;
	} catch (err) {
		toast.error("Failed to rename tag");
	}
};

const startMergeTag = (tag) => {
	mergingTag.value = tag;
	mergeTargetTagId.value = "";
};

const handleMergeTag = async () => {
	if (!mergeTargetTagId.value) return;

	try {
		await circleStore.mergeTags(
			props.id,
			mergingTag.value.id,
			mergeTargetTagId.value,
		);
		toast.success("Tags merged successfully");
		mergingTag.value = null;
	} catch (err) {
		toast.error("Failed to merge tags");
	}
};

const startDeleteTag = (tag) => {
	deletingTag.value = tag;
};

const handleDeleteTag = async () => {
	try {
		await circleStore.deleteTag(props.id, deletingTag.value.id);
		toast.success("Tag and all its posts deleted");
		deletingTag.value = null;
	} catch (err) {
		toast.error("Failed to delete tag");
	}
};

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

const revokeInvite = async (inviteId) => {
	try {
		await circleStore.deleteInvite(props.id, inviteId);
		toast.success("Invite removed");
	} catch (err) {
		toast.error("Failed to remove invite");
	}
};

const clearAllInactive = async () => {
	if (!confirm("Clear all inactive and depleted invites?")) return;
	try {
		for (const invite of inactiveInvites.value) {
			await circleStore.deleteInvite(props.id, invite.id);
		}
		toast.success("Inactive invites cleared");
	} catch (err) {
		toast.error("Failed to clear some invites");
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

const activeInvites = computed(() => {
	return circleStore.invites.filter((i) => {
		const notExpired = !i.expires_at || new Date(i.expires_at) > new Date();
		const notDepleted = !i.max_uses || i.used_count < i.max_uses;
		return notExpired && notDepleted;
	});
});

const inactiveInvites = computed(() => {
	return circleStore.invites.filter((i) => {
		const expired = i.expires_at && new Date(i.expires_at) <= new Date();
		const depleted = i.max_uses && i.used_count >= i.max_uses;
		return expired || depleted;
	});
});

const formatDate = (dateStr) => {
	return new Date(dateStr).toLocaleString();
};

const copyInviteLink = (code) => {
	const url = `${window.location.origin}/join/${code}`;
	navigator.clipboard.writeText(url);
	toast.success("Invite link copied to clipboard!");
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
			<div
				class="flex items-center justify-between border-b border-gray-700 pb-2"
			>
				<h3 class="text-xl font-bold">Invites</h3>
				<button
					@click="showInviteModal = true"
					class="bg-purple-600 hover:bg-purple-700 px-4 py-1.5 rounded font-bold transition text-xs"
				>
					+ New Invite
				</button>
			</div>

			<div v-if="activeInvites.length > 0" class="mt-6 space-y-3">
				<h4 class="text-sm font-bold text-gray-400">Active Invites</h4>
				<div
					v-for="invite in activeInvites"
					:key="invite.id"
					class="flex items-center justify-between bg-gray-900 p-3 rounded border border-gray-700"
				>
					<div class="space-y-1">
						<div class="flex items-center space-x-2">
							<code
								class="bg-purple-900/30 text-purple-400 px-2 py-0.5 rounded font-mono font-bold"
								>{{ invite.code }}</code
							>
							<button
								@click="copyInviteLink(invite.code)"
								class="p-1 hover:bg-gray-700 rounded transition"
								title="Copy Invite Link"
							>
								🔗
							</button>
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

			<div v-if="inactiveInvites.length > 0" class="mt-6 space-y-3 opacity-60">
				<div class="flex items-center justify-between">
					<h4 class="text-sm font-bold text-gray-500">
						Inactive/Depleted Invites
					</h4>
					<button
						@click="clearAllInactive"
						class="text-[10px] uppercase font-bold text-gray-600 hover:text-red-400 transition-colors"
					>
						Clear All
					</button>
				</div>
				<div
					v-for="invite in inactiveInvites"
					:key="invite.id"
					class="flex items-center justify-between bg-gray-950 p-3 rounded border border-gray-800"
				>
					<div class="space-y-1">
						<div class="flex items-center space-x-2">
							<code
								class="bg-gray-800 text-gray-500 px-2 py-0.5 rounded font-mono"
								>{{ invite.code }}</code
							>
							<span class="text-xs text-gray-600"
								>grants {{ invite.role_to_grant }} • Issued by
								{{ invite.created_by }}</span
							>
						</div>
						<div class="text-[10px] text-gray-600">
							Uses: {{ invite.used_count }} / {{ invite.max_uses || "∞" }} •
							Status:
							<span class="text-red-900/80 font-bold uppercase">
								{{
									invite.max_uses && invite.used_count >= invite.max_uses
										? "Depleted"
										: "Expired"
								}}
							</span>
						</div>
					</div>
					<button
						@click="revokeInvite(invite.id)"
						class="text-xs text-gray-500 hover:text-red-400 hover:underline"
					>
						Clear
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

			<div class="space-y-2 pt-2">
				<div
					v-for="tag in circleStore.tags"
					:key="tag.id"
					class="bg-gray-900 border border-gray-700 p-3 rounded-lg flex items-center justify-between group"
				>
					<div class="flex items-center space-x-3">
						<span
							class="text-lg"
							:class="tag.is_pinned ? 'text-purple-400' : 'text-gray-400'"
							>#</span
						>
						<div
							v-if="editingTag?.id === tag.id"
							class="flex items-center space-x-2"
						>
							<input
								v-model="editTagName"
								class="bg-gray-800 border border-purple-500 rounded px-2 py-0.5 text-sm focus:outline-none"
								@keyup.enter="handleUpdateTag"
								@keyup.esc="cancelEditTag"
								v-focus
							/>
							<button @click="handleUpdateTag" class="text-green-500 text-xs">
								Save
							</button>
							<button @click="cancelEditTag" class="text-gray-500 text-xs">
								Cancel
							</button>
						</div>
						<div v-else>
							<span
								class="font-bold"
								:class="tag.is_pinned ? 'text-purple-400' : ''"
								>{{ tag.name }}</span
							>
							<span class="text-[10px] text-gray-500 ml-2 uppercase font-bold"
								>{{ tag.use_count }} posts</span
							>
						</div>
					</div>

					<div
						class="flex items-center space-x-2 opacity-0 group-hover:opacity-100 transition-opacity"
					>
						<button
							@click="togglePin(tag.id, !tag.is_pinned)"
							class="p-1 hover:bg-gray-800 rounded transition text-xs"
							:title="tag.is_pinned ? 'Unpin' : 'Pin'"
						>
							{{ tag.is_pinned ? "📌" : "📍" }}
						</button>
						<button
							@click="startEditTag(tag)"
							class="p-1 hover:bg-gray-800 rounded transition text-xs"
							title="Rename"
						>
							✏️
						</button>
						<button
							@click="startMergeTag(tag)"
							class="p-1 hover:bg-gray-800 rounded transition text-xs"
							title="Merge"
						>
							🔄
						</button>
						<button
							v-if="isAdmin"
							@click="startDeleteTag(tag)"
							class="p-1 hover:bg-red-900/20 rounded transition text-xs text-red-400"
							title="Delete"
						>
							🗑️
						</button>
					</div>
				</div>
			</div>
		</section>

		<!-- Merge Tag Modal -->
		<div
			v-if="mergingTag"
			class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4"
		>
			<div
				class="bg-gray-800 rounded-lg shadow-2xl w-full max-w-md border border-gray-700 p-6 space-y-4"
			>
				<h2 class="text-xl font-bold">Merge Tag #{{ mergingTag.name }}</h2>
				<p class="text-sm text-gray-400">
					This will move all posts from
					<span class="text-white font-bold">#{{ mergingTag.name }}</span> to
					another tag and then delete
					<span class="text-white font-bold">#{{ mergingTag.name }}</span
					>.
				</p>

				<div class="space-y-1">
					<label class="text-xs text-gray-400 font-bold uppercase"
						>Target Tag</label
					>
					<select
						v-model="mergeTargetTagId"
						class="w-full bg-gray-900 border border-gray-700 p-2 rounded focus:outline-none focus:border-purple-500"
					>
						<option value="" disabled>Select a tag to merge into</option>
						<option
							v-for="tag in circleStore.tags.filter(
								(t) => t.id !== mergingTag.id,
							)"
							:key="tag.id"
							:value="tag.id"
						>
							#{{ tag.name }} ({{ tag.use_count }} posts)
						</option>
					</select>
				</div>

				<div class="flex justify-end space-x-3 pt-2">
					<button
						@click="mergingTag = null"
						class="px-4 py-2 text-sm text-gray-400 hover:text-white transition"
					>
						Cancel
					</button>
					<button
						@click="handleMergeTag"
						:disabled="!mergeTargetTagId"
						class="px-4 py-2 bg-purple-600 hover:bg-purple-700 rounded font-bold transition disabled:opacity-50"
					>
						Merge Tags
					</button>
				</div>
			</div>
		</div>

		<!-- Delete Tag Modal -->
		<div
			v-if="deletingTag"
			class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4"
		>
			<div
				class="bg-gray-800 rounded-lg shadow-2xl w-full max-w-md border border-red-900/50 p-6 space-y-4"
			>
				<div class="flex items-center space-x-3 text-red-500">
					<span class="text-3xl">⚠️</span>
					<h2 class="text-xl font-bold">Delete Tag #{{ deletingTag.name }}?</h2>
				</div>

				<div class="bg-red-900/20 border border-red-900/50 p-4 rounded-lg">
					<p class="text-sm text-red-200 font-bold mb-2">
						WARNING: This action is irreversible!
					</p>
					<p class="text-sm text-red-300">
						Deleting this tag will also
						<span class="underline"
							>permanently delete all posts and threads</span
						>
						that belong to it.
					</p>
				</div>

				<p class="text-sm text-gray-400">
					If you want to keep the posts, you should
					<button
						@click="
							startMergeTag(deletingTag);
							deletingTag = null;
						"
						class="text-purple-400 hover:underline"
					>
						merge this tag
					</button>
					into another tag instead.
				</p>

				<div class="flex justify-end space-x-3 pt-2">
					<button
						@click="deletingTag = null"
						class="px-4 py-2 text-sm text-gray-400 hover:text-white transition"
					>
						Cancel
					</button>
					<button
						@click="handleDeleteTag"
						class="px-4 py-2 bg-red-600 hover:bg-red-700 rounded font-bold transition"
					>
						Delete Permanently
					</button>
				</div>
			</div>
		</div>

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

	<InviteModal
		:show="showInviteModal"
		:id="id"
		@close="showInviteModal = false"
		@created="circleStore.fetchInvites(id)"
	/>
</template>
