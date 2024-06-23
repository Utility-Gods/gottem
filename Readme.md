# Gottem

Gottem is a multi-API CLI application that allows users to query different APIs using keyboard shortcuts.

## Prerequisites

- Go 1.22.1 or later

## Building the Application

To build the application, follow these steps:

1. Clone the repository:
   ```
   git clone https://github.com/Utility-Gods/gottem.git
   cd gottem
   ```

2. Build the application:
   ```
   go build -o gottem ./cmd/cli
   ```

   This will create an executable named `gottem` in your current directory.

## Running the Application

After building, you can run the application using:

```
./gottem
```

On Windows, use:

```
gottem.exe
```

## Usage

Once the application is running:

1. You will see a list of available APIs and their shortcuts.
2. Enter your query in the format: `<API_SHORTCUT> <QUERY>`
3. To exit the application, type `exit`

Example:
```
> 1 hello world
Response from Mock API 1: HELLO WORLD
```

## Development

### Adding a New API

To add a new API:

1. Create a new file in `internal/api/` (e.g., `newapi.go`) with your API implementation.
2. Add the new API to the `GetAPIHandlers()` function in `internal/api/handler.go`.

### Running Tests

To run tests:

```
go test ./...
```

## License

[Add your license information here]

## Contributing

[Add contribution guidelines here]

## Contact

[Add contact information or links to issue trackers here]
