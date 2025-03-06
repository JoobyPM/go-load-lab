package handlers

import (
	"fmt"
	"net/http"
	"os"
)

// InfoHandler returns some environment-based information
func InfoHandler(w http.ResponseWriter, r *http.Request) {
	hubLink := os.Getenv("HUB_LINK")
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <title>Image Info</title>
  <link rel="stylesheet" href="/style.css" />
</head>
<body>
  <h1>Image info</h1>
  <p>View on Docker Hub: <a href="%s">%s</a></p>
</body>
</html>
`, hubLink, hubLink)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
