<script setup>
import { ref, onMounted, watch, computed } from "vue";
import { useCircleStore } from "../../stores/circles";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";

const props = defineProps(["id", "members"]);
const circleStore = useCircleStore();
const auth = useAuthStore();
const toast = useToastStore();

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
	{ key: "ocean", label: "Ocean" },
	{ key: "ember", label: "Ember" },
	{ key: "forest", label: "Forest" },
	{ key: "rose", label: "Rose" },
	{ key: "slate", label: "Slate" },
];

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
				<textarea
					v-model="settings.description"
					style="height: 80px"
				></textarea>
			</div>

			<div class="checkbox-row">
				<input
					type="checkbox"
					v-model="settings.allow_freeform_tags"
					id="freeform"
				/>
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
				<button class="btn btn-primary" @click="saveSettings">
					Save Changes
				</button>
			</div>
		</section>

		<!-- Manage Tags -->
		<section class="section-card">
			<h3 class="section-card__title">Manage Tags</h3>
			<div style="display: flex; gap: 0.5rem">
				<input
					v-model="adminNewTagName"
					placeholder="New tag name"
					style="flex: 1"
				/>
				<button class="btn btn-primary" @click="adminAddTag">Add Tag</button>
			</div>

			<div
				style="
					display: flex;
					flex-direction: column;
					gap: 0.375rem;
					padding-top: 0.5rem;
				"
			>
				<div
					v-for="tag in circleStore.tags"
					:key="tag.id"
					class="tag-manage-row"
				>
					<div style="display: flex; align-items: center; gap: 0.625rem">
						<span
							:style="
								tag.is_pinned
									? 'color: var(--accent); font-size: var(--text-lg)'
									: 'color: var(--fg-3); font-size: var(--text-lg)'
							"
							>#</span
						>

						<div
							v-if="editingTag?.id === tag.id"
							style="display: flex; align-items: center; gap: 0.375rem"
						>
							<input
								v-model="editTagName"
								v-focus
								style="
									padding: 0.2rem 0.5rem;
									font-size: var(--text-sm);
									width: 140px;
								"
								@keyup.enter="handleUpdateTag"
								@keyup.esc="cancelEditTag"
							/>
							<button class="btn btn-primary btn-sm" @click="handleUpdateTag">
								Save
							</button>
							<button class="btn btn-ghost btn-sm" @click="cancelEditTag">
								Cancel
							</button>
						</div>
						<div v-else>
							<span
								:style="
									tag.is_pinned
										? 'font-weight: 700; color: var(--accent)'
										: 'font-weight: 700'
								"
							>
								{{ tag.name }}
							</span>
							<span
								style="
									font-size: 10px;
									color: var(--fg-3);
									margin-left: 0.5rem;
									text-transform: uppercase;
									font-weight: 700;
								"
							>
								{{ tag.use_count }} posts
							</span>
						</div>
					</div>

					<div class="tag-manage-row__actions">
						<button
							class="btn-icon"
							:title="tag.is_pinned ? 'Unpin' : 'Pin'"
							@click="togglePin(tag.id, !tag.is_pinned)"
						>
							{{ tag.is_pinned ? "📌" : "📍" }}
						</button>
						<button class="btn-icon" title="Rename" @click="startEditTag(tag)">
							✏️
						</button>
						<button
							class="btn-icon"
							title="Merge into another tag"
							@click="startMergeTag(tag)"
						>
							🔄
						</button>
						<button
							v-if="isAdmin"
							class="btn-icon"
							style="color: var(--danger)"
							title="Delete tag and all its posts"
							@click="startDeleteTag(tag)"
						>
							🗑️
						</button>
					</div>
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
					style="
						display: flex;
						align-items: center;
						justify-content: space-between;
						background: var(--bg-sunken);
						padding: 0.75rem;
						border-radius: var(--r-md);
						border: 1px solid var(--border);
					"
				>
					<div>
						<span style="font-weight: 700">{{ m.username }}</span>
						<span
							style="
								font-size: var(--text-xs);
								color: var(--fg-3);
								margin-left: 0.5rem;
							"
						>
							Invited by {{ m.invited_by }}
						</span>
					</div>
					<div style="display: flex; align-items: center; gap: 0.5rem">
						<select
							v-model="m.role"
							@change="updateMemberRole(m.id, m.role)"
							style="
								font-size: var(--text-xs);
								padding: 0.25rem 0.5rem;
								width: auto;
							"
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
						>
							Kick
						</button>
					</div>
				</div>
			</div>
		</section>

		<!-- Merge Tag Modal -->
		<div
			v-if="mergingTag"
			class="modal-backdrop"
			@click.self="mergingTag = null"
		>
			<div class="modal">
				<div class="modal__header">
					<h2>Merge Tag #{{ mergingTag.name }}</h2>
					<button class="btn-icon" @click="mergingTag = null">✕</button>
				</div>
				<p style="font-size: var(--text-sm); color: var(--fg-2)">
					This will move all posts from
					<strong>#{{ mergingTag.name }}</strong> to another tag and then delete
					<strong>#{{ mergingTag.name }}</strong
					>.
				</p>
				<div class="field">
					<label class="label-uppercase">Target Tag</label>
					<select v-model="mergeTargetTagId">
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
				<div class="form-actions">
					<button class="btn btn-ghost" @click="mergingTag = null">
						Cancel
					</button>
					<button
						class="btn btn-primary"
						:disabled="!mergeTargetTagId"
						@click="handleMergeTag"
					>
						Merge Tags
					</button>
				</div>
			</div>
		</div>

		<!-- Delete Tag Modal -->
		<div
			v-if="deletingTag"
			class="modal-backdrop"
			@click.self="deletingTag = null"
		>
			<div class="modal" style="border-color: var(--error-border)">
				<div class="modal__header">
					<h2 style="color: var(--danger)">
						⚠️ Delete Tag #{{ deletingTag.name }}?
					</h2>
					<button class="btn-icon" @click="deletingTag = null">✕</button>
				</div>
				<div
					style="
						background: var(--danger-bg);
						border: 1px solid var(--error-border);
						border-radius: var(--r-md);
						padding: 1rem;
					"
				>
					<p
						style="
							font-weight: 700;
							color: var(--error-fg);
							margin-bottom: 0.5rem;
						"
					>
						WARNING: This action is irreversible!
					</p>
					<p style="font-size: var(--text-sm); color: var(--error-fg)">
						Deleting this tag will also
						<u>permanently delete all posts and threads</u> that belong to it.
					</p>
				</div>
				<p style="font-size: var(--text-sm); color: var(--fg-2)">
					If you want to keep the posts, you should
					<button
						class="btn btn-ghost btn-sm"
						style="color: var(--accent); display: inline"
						@click="
							startMergeTag(deletingTag);
							deletingTag = null;
						"
					>
						merge this tag
					</button>
					into another tag instead.
				</p>
				<div class="form-actions">
					<button class="btn btn-ghost" @click="deletingTag = null">
						Cancel
					</button>
					<button class="btn btn-danger" @click="handleDeleteTag">
						Delete Permanently
					</button>
				</div>
			</div>
		</div>
	</div>
</template>
