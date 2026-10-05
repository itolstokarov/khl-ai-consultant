# KHL AI Consultant

An AI-powered smart consultant and analytical assistant for the **KHL (Kontinental Hockey League)**. The project is built with **Go** (Golang), containerized using **Docker**, and integrates **YandexGPT** with active **Web Search capabilities** to deliver real-time hockey statistics, match insights, and live sports commentary.

## 🚀 Features

- **Real-Time Data Integration**: Utilizes YandexGPT's built-in web search tool to fetch the latest scores, schedules, and league updates dynamically.
- **Context-Aware Dialogue**: Maintains a structured chat history with intelligent filters to avoid repeating old introduction details and focus strictly on the user's current question.
- **Production-Ready Docker Setup**: Features a multi-stage Docker build optimized for minimal image size and fast deployments.
- **Built-in Error Recovery**: Includes server panic recovery mechanisms to ensure high availability and robust runtime stability.
- **Static Frontend UI**: Serves a clean web interface directly from the Go backend.

## 🛠️ Tech Stack

- **Backend**: Go (Golang)
- **AI Integration**: YandexGPT API (Asynchronous Task Processing Flow)
- **Containerization**: Docker & Docker Compose
- **Frontend**: HTML/JS (Served statically via Go standard library)

## 📋 Prerequisites

Before running the application, make sure you have installed:
- [Docker](https://docs.docker.com/engine/install/) and [Docker Compose](https://docs.docker.com/compose/install/)

You also need an active Yandex Cloud account with a folder ID and an authorized API key.

## ⚙️ Installation & Configuration

1. **Clone the repository:**
   ```bash
   git clone https://github.com/itolstokarov/khl-ai-consultant.git
   cd khl-ai-consultant
   ```

2. **Setup environment variables:**
   Create a `.env` file in the root directory of the project. **Do not commit this file to GitHub** (it is protected by `.gitignore`).

   ```bash
   touch .env
   ```

   Open the `.env` file and populate it with your actual Yandex credentials:
   ```text
   YANDEX_API_KEY=your_secret_api_key_here
   YANDEX_FOLDER_ID=your_folder_id_here
   ```
   *(Note: You can check `.env.example` for the reference structure).*

## 🏃 Running the Application

Launch the application instantly using Docker Compose:

```bash
docker compose up --build
```

The server will initialize and start listening on port `8080`. 

- Open your web browser and navigate to: **`http://localhost:8080`**
- To stop the application, press `Ctrl + C` or run: `docker compose down`

## 🔌 API Reference

### Send a Message to the AI
* **URL**: `/api/ask`
* **Method**: `POST`
* **Headers**: `Content-Type: application/json`
* **Request Body**:
  ```json
  {
    "message": "Как вчера сыграл Автомобилист против Лады?"
  }
  ```
* **Response Body**:
  ```json
  {
    "answer": "Вчера в упорном матче Автомобилист одержал победу со счетом..."
  }
  ```

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](https://github.com/itolstokarov/khl-ai-consultant/blob/main/LICENSE) file for details.