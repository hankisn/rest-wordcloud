# Todolist
- Sjekke om database-fil eksisterer
- Returnere melding om reglene for ord
- Fjerne Config-filen og la all config være i koden
- Default-bilde som gir beskjed om at en må skrive inn ord
- Lage et reset-endepunkt
- Lage et endepunkt for å fjerne enkelte ord

```
docker build -t wordcloud .
docker build --build-arg NEW_LISTEN_PORT=9091 -t wordcloud .
docker run -it --rm -p 9090:9090 --name cloud wordcloud
docker exec -it cloud sh

docker build --build-arg NEW_LISTEN_PORT=9099 -t wordcloud .; docker run -it --rm -p 9099:9099 --name cloud wordcloud
```