# Bulletin - plan for a communications system

I want to design a communication platform that is based on social circles. A user would have one user on the system, and this user would be a member of one of more circles. A circle would be a group set up for a circle of friends, or a specific purpose, like a RPG campaign or a book club. The Circles could display a bit like "servers" in discord, where you switch between them.

The main communication would be through posts, that are always posted in a Circle, and start a persistent threaded conversation. All posts use one or more "tags" to organize them.

In addition to the threaded posts, all circles should have a real-time chat room. The chat rooms should have settings for how much of the chat history will be saved. The default should be 50 messages or 14 days, whichever is longer.

The user system should be invite only. If you get an invite code to a circle, you can create an account and join the circle. If you already have an account, you can join the circle with the invite.

Circles should only be visible to members of the circle. If you are not a member, you should not be able to see the circle or join it.

The system should keep an invite chain, so all users should have an invited_by field. And all user / circle links should also track who invited them to the circle. This should be visible in the UI, but if the user viewing the information doesn't have a connection to the user that invited someone, that user should just be listed as "Unknown".

## The technology

The backend should be written in Go, and use a REST API, combined with websockets for the realtime chat functionality.

We should use PostgreSQL as a database, and perhaps a Redis cache for storing chat history.

The frontend should be written in Vue with a pinia store for state management.

The project will be in a monorepo containing both the backend and frontend code, as well as docker-compose files for running the application locally.
