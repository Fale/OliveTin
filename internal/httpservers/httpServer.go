package httpservers

import (
	config "github.com/jamesread/OliveTin/internal/config"
)

// StartServers will start 3 HTTP servers. The WebUI, the Rest API, and a proxy
// for both of them.
func StartServers(cfg *config.Config) {
	go startWebUIServer(cfg)

	if cfg.UseSingleHTTPFrontend {
		go StartSingleHTTPFrontend(cfg)
	}

	startRestAPIServer(cfg)
}
