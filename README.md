# rest-wordcloud
Rest-service to provide a wordcloud in Golang. It uses a SQLite database located in the `db`-folder inside the container. There are not persistant storage for the database(by design).

## Requirements
### Docker
Tested with Docker version: 23.0.1.

## Usage
### Build
Build with the following command:
```
docker build -t wordcloud .
```
### Example run
```
docker run -it --rm -p 9090:9090 --name cloud wordcloud
```
