![Avazon Banner](https://avazon.s3.us-west-1.amazonaws.com/ppts/avazon_banner.png)

# Avazon API Server

Avazon is your ultimate partner in unleashing creativity for multimedia content creation. The Avatar acts as a unique agent, enhancing user interaction with the platform. Whether you're creating mesmerizing music tracks, generating captivating videos, or remixing stunning images, the Avazon API has everything you need. By leveraging cutting-edge third-party APIs, we provide a comprehensive toolkit designed for passionate content creators.

## Features

- **Music Production**: Effortlessly generate unique music tracks using the innovative JENAI API.
- **Text-to-Speech**: Bring your words to life by converting text to speech with the advanced ElevenLabs API.
- **Image Remixing**: Transform your images artistically with the capabilities of OpenArt and OpenAI tools.
- **Video Production**: Seamlessly create dynamic videos from images using the powerful RunwayML API.
- **User Management**: Benefit from robust user authentication and authorization to keep your creations secure.
- **Asynchronous Processing**: Utilize Celery for efficient background task processing.
- **Real-time Interaction**: Engage in real-time interactions with the Avatar, making the content creation process more dynamic and enjoyable.

## What You Can Do with This API

![Avatar Creation](https://avazon.s3.us-west-1.amazonaws.com/ppts/avatar_creation.png)

1. **Create Avatars**: Use the API to generate unique avatars that represent you or your creative ideas.
2. **Enjoy and Create Avatar Content**: Engage with the content created by your avatars, whether it's music, videos, or images, and create new content effortlessly. [Video](https://avazon.s3.us-west-1.amazonaws.com/ppts/video_creation.png)

   <div style="display: flex; justify-content: space-around; height: 300px;">
       <img src="https://avazon.s3.us-west-1.amazonaws.com/ppts/remixed_original.png" alt="Original Avatar" style="height: 100%;"/>
       <img src="https://avazon.s3.us-west-1.amazonaws.com/ppts/remixed.png" alt="Remixed Avatar" style="height: 100%;"/>
   </div>

3. **Remix Each Other's Avatars**: Collaborate with others by remixing their avatars, allowing for a dynamic and interactive creative experience. The above images showcase a remixed avatar alongside its original version, highlighting the creative transformations possible with the Avazon API.

4. **Remix Each Other's Avatars**: Collaborate with others by remixing their avatars, allowing for a dynamic and interactive creative experience.

## Frameworks & Tools

- **Go**: The backend is built with Go, providing a robust and efficient runtime environment. It must be version 1.22 or higher.
- **Docker**: The application is containerized with Docker, ensuring easy deployment and scalability.

## Installation

1. **Clone the repository**:

   ```bash
   git clone git@github.com:avazon-eth/avazon-server.git
   cd avazon-api
   ```

2. **Set up environment variables**:
   Create a `.env` file in the root directory and add your API keys and other necessary environment variables.

3. **Build the Docker image**:

   ```bash
   docker build -t avazon-api:latest .
   ```

4. **Run the application**:
   ```bash
   docker-compose up
   ```

## Usage

- **API Endpoints**: The API offers various endpoints for managing system prompts, user sessions, and multimedia content creation. Refer to the `main.go` file for a complete list of routes and their purposes.

- **Authentication**: The API employs JWT for authentication. Make sure to include a valid token in the `Authorization` header for protected routes.

## Development

- **Dependencies**: The project utilizes Go modules. Ensure you have Go installed and run `go mod tidy` to install the dependencies.

- **Database**: The application uses SQLite for local development, with the database schema automatically migrated on startup.

## Contributing

1. Fork the repository.
2. Create a new branch: `git checkout -b feature/YourFeature`.
3. Commit your changes: `git commit -m 'Add some feature'`.
4. Push to the branch: `git push origin feature/YourFeature`.
5. Open a pull request.

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for more details.

## Contact

For questions or support, please reach out to [jkya02@gmail.com](mailto:jkya02@gmail.com).
