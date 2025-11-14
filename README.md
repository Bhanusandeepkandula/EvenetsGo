🗓️ Event Planner App

A simple full-stack event management application built with Go (Gin), PostgreSQL, and React.
The project currently runs fully on local development, and I'm in the process of moving the backend to Render for cloud hosting.

⭐ Overview

This app lets users create, edit, and delete events with authentication using JWT.
The backend is written in Go, the frontend in React, and PostgreSQL is used as the main database.

🚀 Features

🔐 JWT Authentication (login, protected routes)

📅 Create, Update, Delete Events

🗄️ PostgreSQL Integration

🌐 REST API using Go + Gin framework

🔌 React frontend connected to backend

⚙️ CORS configuration

❗ Error handling on API calls

🧱 Clean project structure

📦 Tech Stack
Backend

Go

Gin Web Framework

PostgreSQL

JWT

Render (Cloud hosting — in progress)

Frontend

React

Axios

GitHub Pages (deployment)

🛠️ Local Development Setup
1. Clone the Repository
git clone https://github.com/your-username/event-planner.git
cd event-planner

🔙 Backend Setup (Go + PostgreSQL)
Install Dependencies
go mod tidy

Environment Variables

Create a .env file inside the backend folder:

DB_URL=your_postgres_connection_string
JWT_SECRET=your_secret_here

Run the Backend
go run main.go


The API will start on:

http://localhost:8080

🎨 Frontend Setup (React)
Install Dependencies
npm install

Start the App
npm start


The frontend runs at:

http://localhost:3000

Change the API URL

In your React API file:

Development (local):

const API_URL = "http://localhost:8080";


Production (Render):

const API_URL = "https://your-render-backend.onrender.com";


(Cloud deployment is in progress.)

☁️ Cloud Deployment (In Progress)
Backend

Moving Go API from local to Render Web Service

Connecting to Render Managed PostgreSQL

Frontend

Deployed using GitHub Pages

Once the backend is live on Render, the frontend will use the cloud API instead of localhost.

📁 Folder Structure (Example)
/backend
    main.go
    handlers/
    models/
    middleware/
    database/
    .env

/frontend
    src/
    public/
    package.json

📝 Roadmap

Add admin user management

Add refresh tokens

Add pagination for events

Add Docker support

Add unit tests

Finalize Render deployment
