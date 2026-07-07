# NextUp 🚀

A modern, efficient ticketing and Kanban-style project management application designed to help teams and individuals track tasks, manage workflows, and collaborate seamlessly in real-time.

## ✨ Features

- **Real-Time Collaboration**: Instant synchronization of tickets, projects, and users across all clients via WebSockets.
- **Interactive Kanban Boards**: Drag-and-drop tickets across statuses (To Do, In Progress, Done, Archived) with smooth UI interactions.
- **AI-Powered Sub-tasks**: Automatically break down complex tickets into actionable sub-tasks using AI.
- **Role-Based Access Control**: Granular permissions (Owner, Admin, Member) with secure email-based project invitations.
- **Responsive & Accessible UI**: Beautiful light/dark mode interface built with Svelte 5 runes and Tailwind CSS.
- **Robust Testing**: Comprehensive test coverage including unit testing (Vitest), backend integration testing (Testcontainers for Postgres), and E2E testing (Playwright).
- **Automated CI/CD**: Full GitHub Actions pipeline for backend and frontend linting, building, and testing.

## 🛠️ Technologies Used

### Frontend
- **Framework**: SvelteKit (with Svelte 5 Runes)
- **Styling**: Tailwind CSS
- **Testing**: Vitest & Playwright E2E
- **State Management**: Reactive Websocket Stores

### Backend
- **Language**: Go 1.22
- **Router**: `go-chi`
- **Database**: PostgreSQL (Migrations and queries managed via `sqlc`)
- **Authentication**: JWT via HttpOnly Cookies
- **Testing**: Go testing library with Testcontainers

### Infrastructure
- **Containerization**: Docker & multi-stage `Dockerfile`s
- **CI/CD**: GitHub Actions

## 📦 Getting Started

Follow these instructions to get a copy of the project up and running on your local machine for development and testing.

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/install/)
- (Optional, for local non-Docker development) Node.js 22+ and Go 1.22+

### Installation

1. Clone the repository
   ```bash
   git clone https://github.com/LCmaster/NextUp.git
   cd NextUp
   ```

2. Start the entire application stack via Docker Compose:
   ```bash
   docker-compose up --build
   ```
   *The frontend will be available at `http://localhost:5173` and the backend API at `http://localhost:8080`.*

## 🧪 Testing

We take quality seriously. Here is how to run the automated test suites:

### Backend Tests
The backend uses **Testcontainers** to automatically spin up a temporary PostgreSQL database for integration tests.
```bash
cd backend
go test -v ./...
```

### Frontend Tests
```bash
cd frontend
npm ci

# Run unit tests
npm run test:unit

# Run Playwright End-to-End tests
npx playwright install chromium --with-deps
npx playwright test
```

## 🤝 Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'feat: Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.
