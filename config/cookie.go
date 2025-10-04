package config

import (
	"net/http"
)

type CookieConfig struct {
	Name     string
	Path     string
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

func NewCookieConfig(cfg *Config) *CookieConfig {
	sameSite := http.SameSiteDefaultMode
	switch cfg.Cookie.SessionSameSite {
	case "Strict":
		sameSite = http.SameSiteStrictMode
	case "Lax":
		sameSite = http.SameSiteLaxMode
	case "None":
		sameSite = http.SameSiteNoneMode
	}

	return &CookieConfig{
		Name:     cfg.Cookie.SessionCookieName,
		Path:     cfg.Cookie.SessionCookiePath,
		Domain:   cfg.Cookie.SessionDomain,
		Secure:   cfg.Cookie.SessionSecure,
		SameSite: sameSite,
	}
}

func (c *CookieConfig) ToHTTPCookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     c.Name,
		Value:    value,
		Path:     c.Path,
		Domain:   c.Domain,
		HttpOnly: true,
		Secure:   c.Secure,
		SameSite: c.SameSite,
		MaxAge:   maxAge,
	}
}
