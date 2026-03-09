package http

import (
	"fmt"
	"net/http"
	"os"
)

// SwaggerUIHandler returns an HTTP handler that serves the Swagger UI HTML page.
func SwaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Nexora API - Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3/swagger-ui.css" >
    <style>
      html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
      *, *:before, *:after { box-sizing: inherit; }
      body { margin:0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3/swagger-ui-bundle.js"> </script>
    <script src="https://unpkg.com/swagger-ui-dist@3/swagger-ui-standalone-preset.js"> </script>
    <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "/swagger.yaml",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      })
      window.ui = ui
    }
    </script>
</body>
</html>
`)
}

// SwaggerYamlHandler returns an HTTP handler that serves the swagger.yaml file.
func SwaggerYamlHandler(w http.ResponseWriter, r *http.Request) {
	// Try to find swagger.yaml in common locations relative to the backend
	paths := []string{
		"../swagger.yaml",      // relative to backend/cmd/app
		"swagger.yaml",         // in current dir
		"../../swagger.yaml",   // another relative
		"backend/swagger.yaml", // if in root
	}

	var content []byte
	var err error
	for _, p := range paths {
		content, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}

	if err != nil {
		// Try absolute path as last resort if we can determine project root
		// For now, return error if not found in common relative paths
		http.Error(w, "swagger.yaml not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Write(content)
}
