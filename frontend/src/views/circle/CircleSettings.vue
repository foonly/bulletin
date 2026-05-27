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

const isAdmin = computed(() => circleStore.activeCircle?.role === "admin");

const settings = ref({
	name: "",
	description: "",
	allow_freeform_tags: true,
	invite_min_role: "standard",
	chat_retention_days: 14,
	chat_retention_count: 50,
	palette: "violet",
});

const adminNewTagName = ref("");

const palettes = [
	{ key: "violet", label: "Violet" },
	{ key: "ocean",  label: "Ocean" },
	{ key: "ember",  label: "Ember" },
	{ key: "forest", label: "Forest" },
	{ key: "rose",   label: "Rose" },
	{ key: "slate",  label: "Slate" },
];

const syncSettings = () => {
	if (circleStore.activeCircle) {
		settings.value = {
			palette: "violet",
			...circleStore.activeCircle,
		};
	}
};

onMounted(syncSettings);
watch(() => circleStore.activeCircle, syncSettings);

const applyPalette = (palette) => {
	palettes.forEach((p) => document.body.classList.remove(`palette-${p.key}`));
	if (palette !== "violet") {
		document.body.classList.add(`palette-${palette}`);
	}
	// Persist per-circle in localStorage
	localStorage.setItem(`circle-palette-${props.id}`, palette);
};

const selectPalette = (key) => {
	settings.value.palette = key;
	applyPalette(key);
};

const saveSettings = async () => {
	try {
		await circleStore.updateCircle(props.id, settings.value);
		applyPalette(settings.value.palette);
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
		window.dispatchEvent(new CustomEvent("refresh-members"));
	} catch (err) {
		toast.error(err.response?.data || "Failed to kick member");
	}
};

const activeInvites = computed(() =>
	circleStore.invites.filter((i) => {
		const notExpired = !i.expires_at || new Date(i.expires_at) > new Date();
		const notDepleted = !i.max_uses || i.used_count < i.max_uses;
		return notExpired && notDepleted;
	}),
);

const inactiveInvites = computed(() =>
	circleStore.invites.filter((i) => {
		const expired = i.expires_at && new Date(i.expires_at) <= new Date();
		const depleted = i.max_uses && i.used_count >= i.max_uses;
		return expired || depleted;
	}),
);

const formatDate = (dateStr) => new Date(dateStr).toLocaleString();
</script>

<template>
	<div class="circle-settings">
		<!-- Circle Config -->
		<section class="section-card">
			<h3 class="section-card__title">Circle Settings</h3>

			<div class="field-row">
				<div class="field">
					<label>Name</label>
					<input v-model="settings.name" />
				</div>
				<div class="field">
					<label>Invite Min Role</label>
					<select v-model="settings.invite_min_role">
						<option value="guest">Guest</option>
						<option value="standard">Standard</option>
						<option value="mod">Moderator</option>
						<option value="admin">Admin</option>
					</select>
				</div>
			</div>

			<div class="field">
				<label>Description</label>
				<textarea v-model="settings.description" style="height: 80px"></textarea>
			</div>

			<div class="checkbox-row">
				<input type="checkbox" v-model="settings.allow_freeform_tags" id="freeform" />
				<label for="freeform">Allow users to create tags</label>
			</div>

			<!-- Palette picker -->
			<div class="field" style="margin-top: 0.5rem">
				<label class="label-uppercase">Color Palette</label>
				<div class="palette-picker" style="margin-top: 0.5rem">
					<button
						v-for="p in palettes"
						:key="p.key"
						class="palette-swatch"
						:class="[p.key, { active: settings.palette === p.key }]"
						:title="p.label"
						@click="selectPalette(p.key)"
					></button>
				</div>
			</div>

			<div class="form-actions">
				<button class="btn btn-primary" @click="saveSettings">Save Changes</button>
			</div>
		</section>

		<!-- Invites -->
		<section class="section-card">
			<div style="display: flex; align-items: center; justify-content: space-between; padding-bottom: 0.75rem; border-bottom: 1px solid var(--border)">
				<h3>Invites</h3>
				<button class="btn btn-primary btn-sm" @click="showInviteModal = true">+ New Invite</button>
			</div>

			<div v-if="activeInvites.length > 0" style="display: flex; flex-direction: column; gap: 0.75rem">
				<h4 style="font-size: var(--text-sm); font-weight: 700; color: var(--fg-2)">Active Invites</h4>
				<div v-for="invite in activeInvites" :key="invite.id" class="invite-row">
					<div style="display: flex; flex-direction: column; gap: 0.25rem">
						<div style="display: flex; align-items: center; gap: 0.5rem">
							<code class="invite-code">{{ invite.code }}</code>
							<span style="font-size: var(--text-xs); color: var(--fg-3)">
								grants {{ invite.role_to_grant }} · Issued by {{ invite.created_by }}
							</span>
						</div>
						<div style="font-size: 10px; color: var(--fg-3)">
							Uses: {{ invite.used_count }} / {{ invite.max_uses || "∞" }} ·
							Expires: {{ invite.expires_at ? formatDate(invite.expires_at) : "Never" }}
						</div>
					</div>
					<button class="btn btn-ghost btn-sm" style="color: var(--danger)" @click="revokeInvite(invite.id)">Revoke</button>
				</div>
			</div>

			<div v-if="inactiveInvites.length > 0" style="display: flex; flex-direction: column; gap: 0.75rem; opacity: 0.6">
				<div style="display: flex; align-items: center; justify-content: space-between">
					<h4 style="font-size: var(--text-sm); font-weight: 700; color: var(--fg-3)">Inactive / Depleted</h4>
					<button class="btn btn-ghost btn-sm" style="color: var(--danger)" @click="clearAllInactive">Clear All</button>
				</div>
				<div v-for="invite in inactiveInvites" :key="invite.id" class="invite-row inactive">
					<div style="display: flex; flex-direction: column; gap: 0.25rem">
						<div style="display: flex; align-items: center; gap: 0.5rem">
							<code class="invite-code inactive">{{ invite.code }}</code>
							<span style="font-size: var(--text-xs); color: var(--fg-3)">
								grants {{ invite.role_to_grant }} · Issued by {{ invite.created_by }}
							</span>
						</div>
						<div style="font-size: 10px; color: var(--fg-3)">
							Uses: {{ invite.used_count }} / {{ invite.max_uses || "∞" }} ·
							<span style="font-weight: 700; text-transform: uppercase">
								{{ invite.max_uses && invite.used_count >= invite.max_uses ? "Depleted" : "Expired" }}
							</span>
						</div>
					</div>
					<button class="btn btn-ghost btn-sm" style="color: var(--fg-3)" @click="revokeInvite(invite.id)">Clear</button>
				</div>
			</div>
		</section>

		<!-- Manage Tags -->
		<section class="section-card">
			<h3 class="section-card__title">Manage Tags</h3>
			<div style="display: flex; gap: 0.5rem">
				<input v-model="adminNewTagName" placeholder="New tag name" style="flex: 1" />
				<button class="btn btn-primary" @click="adminAddTag">Add Tag</button>
			</div>
			<div style="display: flex; flex-wrap: wrap; gap: 0.5rem; padding-top: 0.5rem">
				<div
					v-for="tag in circleStore.tags"
					:key="tag.id"
					style="display: flex; align-items: center; gap: 0.25rem; background: var(--bg-sunken); border: 1px solid var(--border); padding: 0.375rem 0.75rem; border-radius: var(--r-lg)"
				>
					<span :style="tag.is_pinned ? 'color: var(--accent)' : ''">#{{ tag.name }}</span>
					<button
						class="btn btn-ghost btn-sm"
						style="padding: 0 0.25rem"
						@click="togglePin(tag.id, !tag.is_pinned)"
					>{{ tag.is_pinned ? "Unpin" : "Pin" }}</button>
				</div>
			</div>
		</section>

		<!-- Danger Zone -->
		<section v-if="isAdmin" class="section-card danger">
			<h3 class="section-card__title">Danger Zone</h3>
			<div style="display: flex; flex-direction: column; gap: 0.5rem">
				<div
					v-for="m in members"
					:key="m.id"
					style="display: flex; align-items: center; justify-content: space-between; background: var(--bg-sunken); padding: 0.75rem; border-radius: var(--r-md); border: 1px solid var(--border)"
				>
					<div>
						<span style="font-weight: 700">{{ m.username }}</span>
						<span style="font-size: var(--text-xs); color: var(--fg-3); margin-left: 0.5rem">
							Invited by {{ m.invited_by }}
						</span>
					</div>
					<div style="display: flex; align-items: center; gap: 0.5rem">
						<select
							v-model="m.role"
							@change="updateMemberRole(m.id, m.role)"
							style="font-size: var(--text-xs); padding: 0.25rem 0.5rem; width: auto"
						>
							<option value="guest">Guest</option>
							<option value="standard">Standard</option>
							<option value="mod">Mod</option>
							<option value="admin">Admin</option>
						</select>
						<button
							v-if="m.id !== auth.user?.id"
							class="btn btn-danger btn-sm"
							@click="kickMember(m.id)"
						>Kick</button>
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
