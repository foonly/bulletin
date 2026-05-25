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
	<div class="max-w-3xl mx-auto w-full">
		<div class="bg-gray-800 p-6 rounded-lg border border-gray-700 shadow-xl">
			<div class="flex items-center justify-between mb-6">
				<h3 class="text-xl font-bold">Start a new thread</h3>
				<button
					@click="router.push({ name: 'circle-posts', params: { id: id } })"
					class="text-gray-400 hover:text-gray-200"
				>
					✕
				</button>
			</div>

			<div class="space-y-4">
				<div class="space-y-1">
					<label class="text-xs text-gray-500 font-bold uppercase">Title</label>
					<input
						v-model="newPost.title"
						placeholder="Give your thread a clear title"
						class="w-full bg-gray-900 p-3 rounded-lg border border-gray-700 focus:outline-none focus:border-purple-500 transition-colors"
					/>
				</div>

				<div class="space-y-1">
					<div class="flex items-center justify-between">
						<label class="text-xs text-gray-500 font-bold uppercase"
							>Content</label
						>
						<button
							@click="showPreview = !showPreview"
							class="text-[10px] uppercase font-bold text-purple-400 hover:text-purple-300"
						>
							{{ showPreview ? "Edit Content" : "Show Preview" }}
						</button>
					</div>
					<div
						v-if="showPreview"
						class="min-h-[200px] bg-gray-900 p-3 rounded-lg border border-gray-700 markdown-content text-sm"
						v-html="renderMarkdown(newPost.content)"
					></div>
					<textarea
						v-else
						v-model="newPost.content"
						placeholder="What's on your mind? (Markdown supported)"
						class="w-full bg-gray-900 p-3 rounded-lg border border-gray-700 focus:outline-none focus:border-purple-500 transition-colors min-h-[200px]"
					></textarea>
				</div>

				<div class="space-y-2">
					<label class="text-xs text-gray-500 font-bold uppercase"
						>Tags (at least one required)</label
					>
					<div class="flex flex-wrap gap-2">
						<button
							v-for="tag in circleStore.tags"
							:key="tag.id"
							@click="toggleTagSelection(tag.name)"
							:class="[
								'px-3 py-1 rounded-full text-xs border transition-colors',
								newPost.tags.includes(tag.name)
									? 'bg-purple-600 border-purple-500 text-white'
									: 'bg-gray-900 border-gray-700 text-gray-400 hover:border-gray-500',
							]"
						>
							{{ tag.name }}
						</button>
						<div
							v-if="circleStore.activeCircle?.allow_freeform_tags"
							class="flex items-center space-x-1"
						>
							<input
								v-model="newTagName"
								@keyup.enter="addCustomTag"
								placeholder="Add custom tag..."
								class="bg-gray-900 border border-gray-700 text-xs px-3 py-1 rounded-full focus:outline-none focus:border-purple-500 w-32"
							/>
						</div>
					</div>
					<div v-if="newPost.tags.length > 0" class="flex flex-wrap gap-1.5 mt-2">
						<span
							v-for="tag in newPost.tags"
							:key="tag"
							class="bg-purple-900/30 text-purple-400 text-xs px-2.5 py-1 rounded-full border border-purple-500/30 flex items-center"
						>
							{{ tag }}
							<button
								@click="toggleTagSelection(tag)"
								class="ml-2 text-purple-600 hover:text-purple-400 font-bold"
							>
								×
							</button>
						</span>
					</div>
				</div>

				<div class="flex justify-end space-x-3 pt-4 border-t border-gray-700">
					<button
						@click="router.push({ name: 'circle-posts', params: { id: id } })"
						class="px-6 py-2 rounded-lg font-bold text-gray-400 hover:bg-gray-700 transition-colors"
					>
						Cancel
					</button>
					<button
						@click="handleCreatePost"
						:disabled="
							!newPost.content.trim() ||
							!newPost.title.trim() ||
							newPost.tags.length === 0
						"
						class="bg-purple-600 px-8 py-2 rounded-lg font-bold hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg"
					>
						Create Thread
					</button>
				</div>
			</div>
		</div>
	</div>
</template>
