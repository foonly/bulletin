<script setup>
import { ref } from "vue";
import { useCircleStore } from "../../stores/circles";
import { useRouter } from "vue-router";
import { renderMarkdown } from "../../utils/markdown";

const props = defineProps(["id"]);
const circleStore = useCircleStore();
const router = useRouter();

const newPost = ref({ title: "", content: "", tags: [] });
const newTagName = ref("");
const showPreview = ref(false);

const handleCreatePost = async () => {
	if (!newPost.value.content.trim()) return;
	await circleStore.createPost(props.id, {
		title: newPost.value.title,
		content: newPost.value.content,
		tags: newPost.value.tags,
	});
	router.push({ name: "circle-posts", params: { id: props.id } });
};

const toggleTagSelection = (tagName) => {
	const index = newPost.value.tags.indexOf(tagName);
	if (index > -1) {
		newPost.value.tags.splice(index, 1);
	} else {
		newPost.value.tags.push(tagName);
	}
};

const addCustomTag = () => {
	const tag = newTagName.value.trim().toLowerCase();
	if (tag && !newPost.value.tags.includes(tag)) {
		newPost.value.tags.push(tag);
	}
	newTagName.value = "";
};
</script>

<template>
	<div style="max-width: var(--content-max); margin: 0 auto; width: 100%">
		<div class="section-card">
			<div style="display: flex; align-items: center; justify-content: space-between">
				<h3>Start a new thread</h3>
				<button
					class="btn-icon"
					@click="router.push({ name: 'circle-posts', params: { id } })"
				>✕</button>
			</div>

			<div class="field">
				<label class="label-uppercase">Title</label>
				<input v-model="newPost.title" placeholder="Give your thread a clear title" />
			</div>

			<div class="field">
				<div class="label-meta">
					<label class="label-uppercase">Content</label>
					<button
						class="btn btn-ghost btn-sm"
						style="color: var(--accent)"
						@click="showPreview = !showPreview"
					>
						{{ showPreview ? "Edit Content" : "Show Preview" }}
					</button>
				</div>
				<div
					v-if="showPreview"
					class="markdown-content"
					style="min-height: 200px; background: var(--bg-sunken); padding: 0.75rem; border-radius: var(--r-md); border: 1px solid var(--border)"
					v-html="renderMarkdown(newPost.content)"
				></div>
				<textarea
					v-else
					v-model="newPost.content"
					placeholder="What's on your mind? (Markdown supported)"
					style="min-height: 200px"
				></textarea>
			</div>

			<div class="field">
				<label class="label-uppercase">Tags (at least one required)</label>
				<div style="display: flex; flex-wrap: wrap; gap: 0.5rem; margin-top: 0.375rem">
					<button
						v-for="tag in circleStore.tags"
						:key="tag.id"
						class="tag-chip"
						:class="{ active: newPost.tags.includes(tag.name) }"
						@click="toggleTagSelection(tag.name)"
					>{{ tag.name }}</button>

					<input
						v-if="circleStore.activeCircle?.allow_freeform_tags"
						v-model="newTagName"
						@keyup.enter="addCustomTag"
						placeholder="Add custom tag…"
						style="width: 140px; font-size: var(--text-xs); padding: 0.2em 0.625em; border-radius: var(--r-pill)"
					/>
				</div>

				<div v-if="newPost.tags.length > 0" style="display: flex; flex-wrap: wrap; gap: 0.375rem; margin-top: 0.5rem">
					<span
						v-for="tag in newPost.tags"
						:key="tag"
						class="tag-chip tag-chip-selected"
					>
						{{ tag }}
						<button class="tag-remove" @click="toggleTagSelection(tag)">×</button>
					</span>
				</div>
			</div>

			<div class="form-actions">
				<button
					class="btn btn-ghost"
					@click="router.push({ name: 'circle-posts', params: { id } })"
				>Cancel</button>
				<button
					class="btn btn-primary"
					@click="handleCreatePost"
					:disabled="!newPost.content.trim() || !newPost.title.trim() || newPost.tags.length === 0"
				>Create Thread</button>
			</div>
		</div>
	</div>
</template>
