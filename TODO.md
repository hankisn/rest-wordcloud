# Todolist
- Sjekke om database-fil eksisterer
- Returnere melding om reglene for ord
- Redusere println-statements

```
docker build -t wordcloud .
docker run -it --rm -p 9090:9090 --name cloud wordcloud
docker exec -it cloud sh
```