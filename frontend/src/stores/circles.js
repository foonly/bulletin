import { defineStore } from "pinia";
import axios from "axios";

export const useCircleStore = defineStore("circles", {
	state: () => ({
		circles: [],
		activeCircle: null,
		posts: [],
		chatMessages: [],
	}),
	actions: {
		async fetchCircles() {
			const res = await axios.get("/api/circles");
			this.circles = res.data;
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
