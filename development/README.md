# Development Docker Stack

This project includes a development Docker stack that provides a local LibreNMS instance for testing and development purposes. The stack consists of several containers:

- **librenms**: The main LibreNMS web application
- **db**: MariaDB database for LibreNMS
- **redis**: Redis cache for LibreNMS
- **librenms_dispatcher**: LibreNMS dispatcher service for polling/background tasks

## Available Make Commands

The following make commands are available for managing the development environment:

```shell
# Start the Docker stack in detached mode
make dev-start

# Stop the Docker stack without removing containers
make dev-stop

# Restart the Docker stack
make dev-restart

# View real-time logs from all containers
make dev-debug

# Access a bash shell in the LibreNMS container
make dev-cli

# Completely remove the Docker stack and volumes
make dev-destroy

# Set up a test environment with an admin user and API token for acceptance tests
make dev-testacc
```

The LibreNMS web interface will be available at http://localhost:8000 once the stack is running.

For manual testing, you can use the token specified in your `.env` file by setting the environment variable:

```shell
export LIBRENMS_TOKEN=your_api_token_here
```

This development environment provides a convenient way to test your provider code against a real LibreNMS instance without needing to set up a production environment.
