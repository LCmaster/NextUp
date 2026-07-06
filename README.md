# NextUp

A modern, efficient ticketing and to-do list application designed to help teams and individuals track tasks, manage workflows, and stay organized.

## 🚀 Features

- **Issue Ticketing**: Create, edit, and track tickets with statuses, priorities, and assignees.
- **To-Do Management**: Personal and team-based to-do lists to keep track of daily tasks.
- **Workflow Automation**: Simple state transitions (e.g., To Do -> In Progress -> Done).
- **Responsive Design**: Accessible on desktop and mobile devices.
- **Real-time Updates**: Instant synchronization of tickets and tasks across clients via WebSockets.

## 🛠 Technologies Used

- **Frontend**: SvelteKit
- **Backend**: Golang
- **Database**: PostgreSQL
- **Styling**: Tailwind CSS
- **Containerization**: Docker
- **Real-time**: WebSockets

## 📦 Getting Started

Follow these instructions to get a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

You will need the following tools installed on your system:

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/) (Included with Docker Desktop)

### Installation

1. Clone the repository
   ```bash
   git clone https://github.com/LCmaster/NextUp.git
   ```
2. Navigate to the project directory
   ```bash
   cd NextUp
   ```
3. (Optional) Configure environment variables by copying `.env.example` to `.env`.

### Running Locally

To start the entire application stack (Frontend, Backend, and Database):

```bash
docker-compose up --build
```

## 🤝 Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.
