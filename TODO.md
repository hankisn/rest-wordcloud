# Todolist
- Sjekke om database-fil eksisterer
- Returnere 401 om ord ikke kommer gjennom regexp
    - Returnere melding om reglene for ord
- open /usr/share/nginx/html/cloud.png: permission denied
    - nginx-mappen er readonly...

```
docker build -t wordcloud .
docker run -it --rm -p 8080:8080 -p 9090:9090 --name mordi wordcloud
docker exec -it mordi sh

docker run -it --rm -p 81:80 --name mordi wordcloud

```