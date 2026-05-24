import { defineStore } from "pinia";
import axios from "axios";

export const useCircleStore = defineStore("circles", {
	state: () => ({
		circles: [],
		activeCircle: null,
		threads: [],
		activeThread: null,
		posts: [],
		tags: [],
		invites: [],
		chatMessages: [],
	}),
	actions: {
		async fetchCircles() {
			const res = await axios.get("/api/circles");
			this.circles = res.data;
		},
		async fetchInvites(circleId) {
			const res = await axios.get(`/api/circles/${circleId}/invites`);
			this.invites = res.data;
		},
		async deleteInvite(circleId, inviteId) {
			await axios.delete(`/api/circles/${circleId}/invites/${inviteId}`);
			await this.fetchInvites(circleId);
		},
		async createCircle(circleData) {
			const res = await axios.post("/api/circles", circleData);
			await this.fetchCircles();
			return res.data;
		},
		async fetchTags(circleId) {
			const res = await axios.get(`/api/circles/${circleId}/tags`);
			this.tags = res.data;
		},
		async createTag(circleId, name) {
			await axios.post(`/api/circles/${circleId}/tags`, { name });
			await this.fetchTags(circleId);
		},
		async pinTag(circleId, tagId, isPinned) {
			await axios.post(`/api/circles/${circleId}/tags/${tagId}/pin`, {
				is_pinned: isPinned,
			});
			await this.fetchTags(circleId);
		},
		async fetchThreads(circleId, tag = "") {
			let url = `/api/circles/${circleId}/threads`;
			if (tag) {
				url += `?tag=${encodeURIComponent(tag)}`;
			}
			const res = await axios.get(url);
			this.threads = res.data;
		},
		async fetchThread(circleId, postId) {
			const res = await axios.get(`/api/circles/${circleId}/threads/${postId}`);
			this.activeThread = res.data;
		},
		async updatePost(circleId, postId, content) {
			await axios.put(`/api/circles/${circleId}/threads/${postId}`, {
				content,
			});
		},
		async markRead(circleId, entityId) {
			await axios.post(`/api/circles/${circleId}/read/${entityId}`);
		},
		async fetchPosts(circleId) {
			const res = await axios.get(`/api/circles/${circleId}/posts`);
			this.posts = res.data;
		},
		async fetchChatHistory(circleId) {
			const res = await axios.get(`/api/circles/${circleId}/chat/history`);
			this.chatMessages = res.data;
		},
		async createPost(circleId, postData) {
			await axios.post(`/api/circles/${circleId}/posts`, postData);
			await this.fetchPosts(circleId);
		},
		async updateCircle(circleId, circleData) {
			await axios.put(`/api/circles/${circleId}`, circleData);
			await this.fetchCircles();
		},
		async createInvite(circleId, inviteData) {
			await axios.post(`/api/circles/${circleId}/invites`, inviteData);
		},
		async updateMember(circleId, userId, role) {
			await axios.put(`/api/circles/${circleId}/members/${userId}`, { role });
		},
		async deleteMember(circleId, userId) {
			await axios.delete(`/api/circles/${circleId}/members/${userId}`);
		},
		addChatMessage(msg) {
			this.chatMessages.push(msg);
		},
	},
});
