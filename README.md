# ChittyChat

## Description
ChittyChat is a chat application that allows users to communicate with each other through server-side streaming. Users can join a chat and send messages to the server, which will then broadcast the message to all other active users. The application is built using golang and gRPC.

## To get started
1. Clone the repository
2. Run the server
    ```bash
    cd ChittyChat/server
    go run .
    ```
3. Run the client
    ```bash
    cd ChittyChat/client
    go run . <client name>
    ```
    **NOTE:** The client name is required and must be unique. Run the command multiple times with different client names to simulate multiple users. Please note that you have to use seperate terminal windows for each client.
4. Type your message and press enter to send it to the server. The server will then broadcast the message to all other active users.
5. To leave the chat type
    ```bash
    .exit
    ```

## Requirements
- The application is built using go version 1.23.1